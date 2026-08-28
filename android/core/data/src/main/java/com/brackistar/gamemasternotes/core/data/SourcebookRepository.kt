package com.brackistar.gamemasternotes.core.data

import com.brackistar.gamemasternotes.core.retrieval.RetrievalQuery
import com.brackistar.gamemasternotes.core.retrieval.RetrievalRepository
import com.brackistar.gamemasternotes.core.retrieval.RetrievalResult
import kotlinx.coroutines.flow.Flow

class SourcebookRepository(
    private val database: AppDatabase,
) : RetrievalRepository {
    private val dao = database.sourcebookDao()

    fun observePackCount(): Flow<Int> = dao.observePackCount()

    fun observePacks(): Flow<List<SourcebookPackSummary>> = dao.observePacks()

    suspend fun replaceImportedPack(pack: ImportedPack) {
        database.replaceImportedPack(pack)
    }

    suspend fun existingFingerprint(packId: String): String? = dao.archiveFingerprint(packId)

    suspend fun packIds(): Set<String> = dao.packIds().toSet()

    suspend fun pruneToAvailablePacks(packIds: List<String>) {
        if (packIds.isEmpty()) {
            dao.deleteAllPacks()
        } else {
            dao.deletePacksNotIn(packIds)
        }
    }

    override suspend fun search(query: RetrievalQuery): List<RetrievalResult> {
        val terms = query.text.significantTerms()
        val ftsQuery = terms
            .take(MAX_QUERY_TERMS)
            .joinToString(" AND ") { "\"$it\"" }
        if (ftsQuery.isBlank()) return emptyList()

        val resultLimit = query.limit.coerceIn(1, 8)
        val candidateLimit = resultLimit * CANDIDATE_MULTIPLIER
        val strictRows = dao.searchChunks(ftsQuery, limit = candidateLimit)
        val (rows, minimumScore) = if (strictRows.isNotEmpty()) {
            strictRows to MIN_EXCERPT_SCORE
        } else {
            val relaxedQuery = terms
                .take(MAX_QUERY_TERMS)
                .joinToString(" OR ") { "\"$it\"" }
            dao.searchChunks(relaxedQuery, limit = candidateLimit) to
                minOf(MIN_RELAXED_EXCERPT_SCORE, terms.size)
        }

        return rows
            .mapNotNull { row ->
                val excerpt = row.text.bestExcerptFor(terms, minimumScore) ?: return@mapNotNull null
                RetrievalResult(
                    sourceId = row.chunkId,
                    title = row.packTitle,
                    snippet = excerpt.text,
                    citationLabel = "${row.packTitle}, pp. ${row.pageStart}-${row.pageEnd}",
                    score = excerpt.score + -row.rank,
                )
            }
            .sortedByDescending { it.score }
            .diversifyResults(resultLimit)
    }

}

private fun List<RetrievalResult>.diversifyResults(limit: Int): List<RetrievalResult> {
    val byCitation = mutableMapOf<String, Int>()
    val byTitle = mutableMapOf<String, Int>()
    return asSequence()
        .filter { result ->
            val citationKey = result.citationLabel ?: result.sourceId
            val citationCount = byCitation[citationKey] ?: 0
            val titleCount = byTitle[result.title] ?: 0
            if (citationCount >= MAX_RESULTS_PER_CITATION || titleCount >= MAX_RESULTS_PER_TITLE) {
                false
            } else {
                byCitation[citationKey] = citationCount + 1
                byTitle[result.title] = titleCount + 1
                true
            }
        }
        .take(limit)
        .toList()
}

private const val MAX_QUERY_TERMS = 5
private const val MAX_SNIPPET_CHARS = 1_800
private const val FALLBACK_SENTENCE_RADIUS = 2
private const val MAX_PARAGRAPHS_PER_SNIPPET = 3
private const val MIN_EXCERPT_SCORE = 1
private const val MIN_RELAXED_EXCERPT_SCORE = 2
private const val CANDIDATE_MULTIPLIER = 3
private const val MAX_RESULTS_PER_CITATION = 2
private const val MAX_RESULTS_PER_TITLE = 3

private val STOP_WORDS = setOf(
    "the",
    "and",
    "for",
    "from",
    "what",
    "when",
    "where",
    "which",
    "with",
    "about",
    "that",
    "this",
    "does",
    "have",
    "how",
    "are",
    "was",
    "were",
    "into",
    "book",
    "books",
    "tell",
    "explain",
    "describe",
)

private fun String.significantTerms(): List<String> =
    lowercase()
        .split(Regex("""[^a-z0-9]+"""))
        .map { it.trim() }
        .filter { it.length >= 3 && it !in STOP_WORDS }
        .distinct()

private fun String.bestExcerptFor(terms: List<String>, minimumScore: Int): RankedExcerpt? {
    if (terms.isEmpty()) return null
    val allParagraphs = paragraphs()
    val paragraph = allParagraphs
        .map { text -> text to text.scoreAgainst(terms) }
        .filter { it.second >= minimumScore }
        .sortedByDescending { it.second }
        .firstOrNull()
    if (paragraph != null) {
        val usefulBlock = allParagraphs
            .filter { text -> text.scoreAgainst(terms) >= minimumScore }
            .take(MAX_PARAGRAPHS_PER_SNIPPET)
            .joinToString("\n")
        return RankedExcerpt(text = usefulBlock.takeCleanly(MAX_SNIPPET_CHARS), score = paragraph.second.toDouble())
    }

    return bestSentenceWindowFor(terms, minimumScore)
}

private fun String.paragraphs(): List<String> =
    split(Regex("""\n\s*\n+"""))
        .map { it.normalizeWhitespace() }
        .filter { it.isNotBlank() }

private fun String.bestSentenceWindowFor(terms: List<String>, minimumScore: Int): RankedExcerpt? {
    val sentences = split(Regex("""(?<=[.!?])\s+"""))
        .map { it.normalizeWhitespace() }
        .filter { it.isNotBlank() }
    val bestIndex = sentences
        .indices
        .map { index -> index to sentences[index].scoreAgainst(terms) }
        .filter { it.second >= minimumScore }
        .maxByOrNull { it.second }
        ?: return null
    val start = maxOf(0, bestIndex.first - FALLBACK_SENTENCE_RADIUS)
    val endExclusive = minOf(sentences.size, bestIndex.first + FALLBACK_SENTENCE_RADIUS + 1)
    val text = sentences
        .subList(start, endExclusive)
        .joinToString(" ")
        .takeCleanly(MAX_SNIPPET_CHARS)
    return RankedExcerpt(text = text, score = bestIndex.second.toDouble())
}

private fun String.scoreAgainst(terms: List<String>): Int {
    val lower = lowercase()
    return terms.count { term -> lower.contains(term) }
}

private fun String.normalizeWhitespace(): String =
    replace(Regex("""\s+"""), " ").trim()

private fun String.takeCleanly(maxChars: Int): String {
    if (length <= maxChars) return this
    val clipped = take(maxChars)
    return clipped.substringBeforeLast(" ").ifBlank { clipped }.trimEnd() + "..."
}

private data class RankedExcerpt(
    val text: String,
    val score: Double,
)

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

        return dao.searchChunks(ftsQuery, limit = query.limit.coerceIn(1, 8) * CANDIDATE_MULTIPLIER)
            .mapNotNull { row ->
                val excerpt = row.text.bestExcerptFor(terms) ?: return@mapNotNull null
                RetrievalResult(
                    sourceId = row.chunkId,
                    title = row.packTitle,
                    snippet = excerpt.text,
                    citationLabel = "${row.packTitle}, pp. ${row.pageStart}-${row.pageEnd}",
                    score = excerpt.score + -row.rank,
                )
            }
            .sortedByDescending { it.score }
            .take(query.limit.coerceIn(1, 8))
    }

}

private const val MAX_QUERY_TERMS = 5
private const val MAX_SNIPPET_CHARS = 900
private const val FALLBACK_SENTENCE_RADIUS = 1
private const val MIN_EXCERPT_SCORE = 1
private const val CANDIDATE_MULTIPLIER = 3

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

private fun String.bestExcerptFor(terms: List<String>): RankedExcerpt? {
    if (terms.isEmpty()) return null
    val paragraph = paragraphs()
        .map { text -> text to text.scoreAgainst(terms) }
        .filter { it.second >= MIN_EXCERPT_SCORE }
        .sortedByDescending { it.second }
        .firstOrNull()
    if (paragraph != null) {
        return RankedExcerpt(text = paragraph.first.takeCleanly(MAX_SNIPPET_CHARS), score = paragraph.second.toDouble())
    }

    return bestSentenceWindowFor(terms)
}

private fun String.paragraphs(): List<String> =
    split(Regex("""\n\s*\n+"""))
        .map { it.normalizeWhitespace() }
        .filter { it.isNotBlank() }

private fun String.bestSentenceWindowFor(terms: List<String>): RankedExcerpt? {
    val sentences = split(Regex("""(?<=[.!?])\s+"""))
        .map { it.normalizeWhitespace() }
        .filter { it.isNotBlank() }
    val bestIndex = sentences
        .indices
        .map { index -> index to sentences[index].scoreAgainst(terms) }
        .filter { it.second >= MIN_EXCERPT_SCORE }
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

package com.brackistar.gamemasternotes.core.data

import com.brackistar.gamemasternotes.core.retrieval.RetrievalQuery
import com.brackistar.gamemasternotes.core.retrieval.RetrievalRepository
import com.brackistar.gamemasternotes.core.retrieval.RetrievalResult
import com.brackistar.gamemasternotes.core.domain.diagnostics.DiagnosticEvent
import com.brackistar.gamemasternotes.core.domain.diagnostics.DiagnosticsJournal
import com.brackistar.gamemasternotes.core.domain.diagnostics.NoOpDiagnosticsJournal
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.withContext

class SourcebookRepository(
    private val database: AppDatabase,
    private val rankingDispatcher: CoroutineDispatcher = Dispatchers.Default,
    private val diagnosticsJournal: DiagnosticsJournal = NoOpDiagnosticsJournal,
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
        database.pruneToAvailablePacks(packIds)
    }

    suspend fun resultsBySourceIds(sourceIds: List<String>): List<RetrievalResult> {
        if (sourceIds.isEmpty()) return emptyList()
        return dao.chunksByIds(sourceIds).map { row ->
            RetrievalResult(
                sourceId = row.chunkId,
                packId = row.packId,
                documentId = row.documentId,
                pageStart = row.pageStart,
                pageEnd = row.pageEnd,
                title = row.packTitle,
                snippet = row.text.takeCleanly(MAX_SNIPPET_CHARS),
                citationLabel = row.citationLabel,
                score = 0.0,
            )
        }
    }

    suspend fun relatedResults(sourceIds: List<String>, limit: Int): List<RetrievalResult> {
        if (sourceIds.isEmpty() || limit <= 0) return emptyList()
        return dao.relatedChunks(sourceIds, limit).map { row ->
            RetrievalResult(
                sourceId = row.chunkId,
                packId = row.packId,
                documentId = row.documentId,
                pageStart = row.pageStart,
                pageEnd = row.pageEnd,
                title = row.packTitle,
                snippet = row.text.takeCleanly(MAX_SNIPPET_CHARS),
                citationLabel = row.citationLabel,
                score = 0.0,
            )
        }
    }

    override suspend fun search(query: RetrievalQuery): List<RetrievalResult> = withContext(rankingDispatcher) {
        val startedAt = System.currentTimeMillis()
        val terms = query.text.significantTerms()
        val ftsQuery = terms
            .take(MAX_QUERY_TERMS)
            .joinToString(" AND ") { "\"$it\"" }
        if (ftsQuery.isBlank()) return@withContext emptyList()

        val resultLimit = query.limit.coerceIn(1, 8)
        val strictRows = dao.searchChunks(ftsQuery, limit = MAX_FTS_SCAN_ROWS)
        val (rows, minimumScore, retrievalMode) = if (strictRows.isNotEmpty()) {
            Triple(strictRows, MIN_EXCERPT_SCORE, "strict")
        } else {
            val relaxedQuery = terms
                .take(MAX_QUERY_TERMS)
                .joinToString(" OR ") { "\"$it\"" }
            Triple(
                dao.searchChunks(relaxedQuery, limit = MAX_FTS_SCAN_ROWS),
                minOf(MIN_RELAXED_EXCERPT_SCORE, terms.size),
                "relaxed",
            )
        }

        val rankedCandidates = rows
            .mapNotNull { row ->
                val excerpt = row.text.bestExcerptFor(terms, minimumScore) ?: return@mapNotNull null
                RetrievalResult(
                    sourceId = row.chunkId,
                    packId = row.packId,
                    documentId = row.documentId,
                    pageStart = row.pageStart,
                    pageEnd = row.pageEnd,
                    title = row.packTitle,
                    snippet = excerpt.text,
                    citationLabel = row.citationLabel,
                    score = row.lexicalScore(query.text, terms, excerpt),
                )
            }
            .sortedWith(compareByDescending<RetrievalResult> { it.score }.thenBy { it.sourceId })
            .distinctBy { it.stableIdentity() }
            .take(MAX_RANKED_CANDIDATES)

        query.requestId?.let { requestId ->
            runCatching {
                diagnosticsJournal.record(
                    DiagnosticEvent(
                        requestId = requestId,
                        timestampEpochMillis = System.currentTimeMillis(),
                        stage = "lexical-retrieval",
                        outcome = "completed",
                        elapsedMs = System.currentTimeMillis() - startedAt,
                        fields = mapOf(
                            "retrievalMode" to retrievalMode,
                            "scannedCount" to rows.size.toString(),
                            "candidateCount" to rankedCandidates.size.toString(),
                            "candidateIds" to rankedCandidates.joinToString(",") { it.sourceId },
                            "lexicalRanks" to rankedCandidates.mapIndexed { index, result ->
                                "${result.sourceId}:${index + 1}"
                            }.joinToString(","),
                        ),
                    ),
                )
            }
        }

        rankedCandidates
            .diversifyResults(resultLimit)
    }

}

private fun SourceChunkSearchRow.lexicalScore(
    queryText: String,
    terms: List<String>,
    excerpt: RankedExcerpt,
): Double {
    val normalizedText = text.lowercase()
    val normalizedQuery = queryText.lowercase().trim()
    val exactPhrase = if (normalizedQuery.length >= 3 && normalizedText.contains(normalizedQuery)) 8.0 else 0.0
    val allTerms = if (terms.isNotEmpty() && terms.all(normalizedText::contains)) 6.0 else 0.0
    val frequency = terms.sumOf { term -> normalizedText.nonOverlappingCount(term).coerceAtMost(3) }.coerceAtMost(10)
    val proximity = normalizedText.termProximityScore(terms)
    val titleMatch = if (terms.any { packTitle.contains(it, ignoreCase = true) }) 3.0 else 0.0
    val systemMatch = if (terms.any { system.contains(it, ignoreCase = true) }) 2.0 else 0.0
    val structuralQuality = if (pageStart > 0 && pageEnd >= pageStart && citationLabel.isNotBlank()) 1.0 else 0.0
    return excerpt.score + exactPhrase + allTerms + frequency + proximity + titleMatch + systemMatch + structuralQuality - rank
}

private fun String.nonOverlappingCount(term: String): Int {
    var count = 0
    var start = 0
    while (true) {
        val index = indexOf(term, start)
        if (index < 0) return count
        count += 1
        start = index + term.length
    }
}

private fun String.termProximityScore(terms: List<String>): Double {
    if (terms.size < 2) return 0.0
    val positions = terms.map { term -> indexOf(term) }
    if (positions.any { it < 0 }) return 0.0
    val span = positions.max() - positions.min()
    return when {
        span <= 40 -> 4.0
        span <= 120 -> 2.0
        else -> 0.0
    }
}

private fun RetrievalResult.stableIdentity(): String = "$packId:$documentId:$sourceId"

private fun List<RetrievalResult>.diversifyResults(limit: Int): List<RetrievalResult> {
    val byCitation = mutableMapOf<String, Int>()
    val byTitle = mutableMapOf<String, Int>()
    return asSequence()
        .filter { result ->
            val citationKey = result.citationLabel
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
private const val MAX_FTS_SCAN_ROWS = 256
private const val MAX_RANKED_CANDIDATES = 32
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

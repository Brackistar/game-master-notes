package com.brackistar.gamemasternotes.core.retrieval

import com.brackistar.gamemasternotes.core.domain.diagnostics.DiagnosticEvent
import com.brackistar.gamemasternotes.core.domain.diagnostics.DiagnosticsJournal
import com.brackistar.gamemasternotes.core.domain.diagnostics.NoOpDiagnosticsJournal
import kotlin.math.sqrt

interface QueryEmbeddingEncoder {
    val modelId: String
    val dimensions: Int
    suspend fun encode(text: String): FloatArray
}

interface VectorRetrievalStore {
    val modelId: String
    val dimensions: Int
    suspend fun search(queryVector: FloatArray, limit: Int): List<RankedCandidate>
}

fun interface RetrievalResultResolver {
    suspend fun resolve(sourceIds: List<String>): List<RetrievalResult>
}

fun interface RelatedResultResolver {
    suspend fun resolve(seedIds: List<String>, limit: Int): List<RetrievalResult>
}

class HybridRetrievalRepository(
    private val lexicalRepository: RetrievalRepository,
    private val encoder: QueryEmbeddingEncoder,
    private val vectorStore: VectorRetrievalStore,
    private val resultResolver: RetrievalResultResolver,
    private val relatedResultResolver: RelatedResultResolver? = null,
    private val diagnosticsJournal: DiagnosticsJournal = NoOpDiagnosticsJournal,
) : RetrievalRepository {
    init {
        require(encoder.modelId == vectorStore.modelId) { "Query and pack embedding model IDs must match." }
        require(encoder.dimensions == vectorStore.dimensions) { "Query and pack embedding dimensions must match." }
    }

    override suspend fun search(query: RetrievalQuery): List<RetrievalResult> {
        val startedAt = System.currentTimeMillis()
        val lexical = lexicalRepository.search(query.copy(limit = CANDIDATE_LIMIT))
        val vector = runCatching {
            vectorStore.search(encoder.encode(query.text), CANDIDATE_LIMIT)
        }.getOrElse { error ->
            record(query, startedAt, "lexical-fallback", lexical, emptyList(), emptyList(), error::class.java.simpleName)
            return lexical.take(query.limit)
        }
        val relationshipQuestion = query.text.isRelationshipQuestion()
        val fused = ReciprocalRankFusion.fuse(
            lexical = lexical.map { RankedCandidate(it.sourceId, it.score) },
            vector = vector,
            limit = if (relationshipQuestion && query.limit >= 4) query.limit - 1 else query.limit,
        )
        val semanticResults = resultResolver.resolve(vector.map { it.sourceId })
        val byId = (lexical + semanticResults).associateBy { it.sourceId }
        val primary = fused.mapNotNull { rank -> byId[rank.sourceId]?.copy(score = rank.score) }
        val results = if (relationshipQuestion && relatedResultResolver != null) {
            val adjacent = relatedResultResolver.resolve(primary.map { it.sourceId }, MAX_RELATIONSHIP_EXPANSION)
            val names = primary.flatMap { it.snippet.properNameTerms() }.distinct().take(2)
            val repeatedName = if (names.isEmpty()) emptyList() else {
                lexicalRepository.search(query.copy(text = names.joinToString(" "), limit = MAX_RELATIONSHIP_EXPANSION))
            }
            (primary + adjacent + repeatedName)
                .distinctBy { it.sourceId }
                .take(query.limit)
        } else {
            primary
        }
        record(query, startedAt, "hybrid", lexical, vector, fused, null)
        return results
    }

    private suspend fun record(
        query: RetrievalQuery,
        startedAt: Long,
        mode: String,
        lexical: List<RetrievalResult>,
        vector: List<RankedCandidate>,
        fused: List<RankedCandidate>,
        errorType: String?,
    ) {
        val requestId = query.requestId ?: return
        runCatching {
            diagnosticsJournal.record(
                DiagnosticEvent(
                    requestId = requestId,
                    timestampEpochMillis = System.currentTimeMillis(),
                    stage = "hybrid-retrieval",
                    outcome = mode,
                    elapsedMs = System.currentTimeMillis() - startedAt,
                    fields = buildMap {
                        put("lexicalRanks", lexical.mapIndexed { i, item -> "${item.sourceId}:${i + 1}" }.joinToString(","))
                        put("vectorRanks", vector.mapIndexed { i, item -> "${item.sourceId}:${i + 1}" }.joinToString(","))
                        put("fusedRanks", fused.mapIndexed { i, item -> "${item.sourceId}:${i + 1}" }.joinToString(","))
                        errorType?.let { put("errorType", it) }
                    },
                ),
            )
        }
    }

    private companion object {
        const val CANDIDATE_LIMIT = 20
        const val MAX_RELATIONSHIP_EXPANSION = 4
    }
}

private fun String.isRelationshipQuestion(): Boolean {
    val value = lowercase()
    return listOf("connected", "connection", "relationship", "related", "between", "how does", "how is")
        .any(value::contains)
}

private fun String.properNameTerms(): List<String> =
    Regex("\\b[A-Z][A-Za-z]{3,}\\b").findAll(this).map { it.value }.toList()

data class RankedCandidate(val sourceId: String, val score: Double)

object ReciprocalRankFusion {
    fun fuse(
        lexical: List<RankedCandidate>,
        vector: List<RankedCandidate>,
        limit: Int,
        rankConstant: Int = 60,
    ): List<RankedCandidate> {
        require(limit > 0 && rankConstant > 0)
        val scores = linkedMapOf<String, Double>()
        listOf(lexical, vector).forEach { ranking ->
            ranking.distinctBy { it.sourceId }.forEachIndexed { index, candidate ->
                scores[candidate.sourceId] = scores.getOrDefault(candidate.sourceId, 0.0) +
                    1.0 / (rankConstant + index + 1)
            }
        }
        return scores.map { RankedCandidate(it.key, it.value) }
            .sortedWith(compareByDescending<RankedCandidate> { it.score }.thenBy { it.sourceId })
            .take(limit)
    }
}

object BruteForceVectorSearch {
    fun search(query: FloatArray, rows: List<Pair<String, FloatArray>>, limit: Int): List<RankedCandidate> {
        require(limit > 0)
        require(query.isNotEmpty() && query.all(Float::isFinite))
        return rows.asSequence()
            .map { (sourceId, vector) ->
                require(vector.size == query.size && vector.all(Float::isFinite))
                RankedCandidate(sourceId, cosine(query, vector))
            }
            .sortedWith(compareByDescending<RankedCandidate> { it.score }.thenBy { it.sourceId })
            .take(limit)
            .toList()
    }

    private fun cosine(left: FloatArray, right: FloatArray): Double {
        var dot = 0.0
        var leftNorm = 0.0
        var rightNorm = 0.0
        for (index in left.indices) {
            dot += left[index] * right[index]
            leftNorm += left[index] * left[index]
            rightNorm += right[index] * right[index]
        }
        if (leftNorm == 0.0 || rightNorm == 0.0) return 0.0
        return dot / (sqrt(leftNorm) * sqrt(rightNorm))
    }
}

data class RelationshipCandidate(
    val sourceId: String,
    val documentId: String,
    val rowIndex: Int,
    val pageStart: Int,
    val pageEnd: Int,
    val normalizedNames: Set<String> = emptySet(),
)

object BoundedRelationshipExpansion {
    fun expand(
        seeds: List<RelationshipCandidate>,
        corpus: List<RelationshipCandidate>,
        maxAdded: Int = 4,
    ): List<RelationshipCandidate> {
        if (maxAdded <= 0 || seeds.isEmpty()) return emptyList()
        val seedIds = seeds.mapTo(hashSetOf()) { it.sourceId }
        return corpus.asSequence()
            .filterNot { it.sourceId in seedIds }
            .mapNotNull { candidate ->
                val score = seeds.maxOf { seed -> relationshipScore(seed, candidate) }
                candidate.takeIf { score > 0 }?.let { it to score }
            }
            .sortedWith(compareByDescending<Pair<RelationshipCandidate, Int>> { it.second }.thenBy { it.first.sourceId })
            .map { it.first }
            .take(maxAdded)
            .toList()
    }

    private fun relationshipScore(left: RelationshipCandidate, right: RelationshipCandidate): Int {
        var score = 0
        if (left.documentId == right.documentId && kotlin.math.abs(left.rowIndex - right.rowIndex) == 1) score += 3
        if (left.documentId == right.documentId && left.pageStart <= right.pageEnd && right.pageStart <= left.pageEnd) score += 2
        if (left.normalizedNames.intersect(right.normalizedNames).isNotEmpty()) score += 1
        return score
    }
}

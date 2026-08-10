package com.brackistar.gamemasternotes.core.retrieval

data class RetrievalQuery(
    val text: String,
    val campaignId: String? = null,
    val systemId: String? = null,
)

data class RetrievalResult(
    val sourceId: String,
    val title: String,
    val snippet: String,
    val citationLabel: String?,
    val score: Double,
)

interface RetrievalRepository {
    suspend fun search(query: RetrievalQuery): List<RetrievalResult>
}

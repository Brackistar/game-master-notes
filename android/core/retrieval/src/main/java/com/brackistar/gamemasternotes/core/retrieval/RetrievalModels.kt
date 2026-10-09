package com.brackistar.gamemasternotes.core.retrieval

data class RetrievalQuery(
    val text: String,
    val requestId: String? = null,
    val campaignId: String? = null,
    val systemId: String? = null,
    val limit: Int = 4,
)

data class RetrievalResult(
    val sourceId: String,
    val packId: String,
    val documentId: String,
    val pageStart: Int,
    val pageEnd: Int,
    val title: String,
    val snippet: String,
    val citationLabel: String,
    val score: Double,
)

interface RetrievalRepository {
    suspend fun search(query: RetrievalQuery): List<RetrievalResult>
}

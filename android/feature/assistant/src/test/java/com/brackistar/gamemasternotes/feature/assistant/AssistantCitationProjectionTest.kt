package com.brackistar.gamemasternotes.feature.assistant

import com.brackistar.gamemasternotes.core.ai.AiResponse
import com.brackistar.gamemasternotes.core.ai.AiAnswerProvenance
import com.brackistar.gamemasternotes.core.retrieval.RetrievalResult
import org.junit.Assert.assertEquals
import org.junit.Test

class AssistantCitationProjectionTest {
    @Test
    fun responseCitationIdsSelectOnlyMatchingSources() {
        val first = result("chunk-1", "Book, p. 1")
        val second = result("chunk-2", "Book, p. 2")
        val response = AiResponse(
            requestId = "request-1",
            text = "Supported answer [Book, p. 2]",
            citationIds = listOf("chunk-2"),
            provenance = AiAnswerProvenance.systemMessage("test"),
        )

        val (citations, retrievedEvidence) = resolveAnswerSources(response, listOf(first, second))

        assertEquals(listOf(second), citations)
        assertEquals(listOf(first), retrievedEvidence)
    }

    @Test
    fun unknownResponseCitationIdsNeverCreateDisplayedCitations() {
        val result = result("chunk-1", "Book, p. 1")
        val response = AiResponse(
            requestId = "request-1",
            text = "Unsupported answer [Unknown, p. 9]",
            citationIds = listOf("missing-chunk"),
            provenance = AiAnswerProvenance.systemMessage("test"),
        )

        val (citations, retrievedEvidence) = resolveAnswerSources(response, listOf(result))

        assertEquals(emptyList<RetrievalResult>(), citations)
        assertEquals(listOf(result), retrievedEvidence)
    }

    private fun result(sourceId: String, citationLabel: String) = RetrievalResult(
        sourceId = sourceId,
        packId = "pack-1",
        documentId = "doc-1",
        pageStart = 1,
        pageEnd = 1,
        title = "Book",
        snippet = "Evidence text.",
        citationLabel = citationLabel,
        score = 1.0,
    )
}

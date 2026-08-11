package com.brackistar.gamemasternotes.core.ai

import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class GroundedMvpAiEngineTest {
    @Test
    fun generateReportsMissingContext() = runTest {
        val response = GroundedMvpAiEngine().generate(
            AiRequest(prompt = "What is paradox?", context = ""),
        )

        assertEquals(emptyList<String>(), response.citationIds)
        assertTrue(response.text.contains("could not find relevant passages"))
    }

    @Test
    fun generateUsesOnlyProvidedContextWithCitations() = runTest {
        val response = GroundedMvpAiEngine().generate(
            AiRequest(
                prompt = "What is paradox?",
                context = "[Core Book, pp. 10-11]\nParadox follows vulgar magic.",
            ),
        )

        assertEquals(listOf("Core Book, pp. 10-11"), response.citationIds)
        assertTrue(response.text.contains("Paradox follows vulgar magic."))
        assertTrue(response.text.contains("[Core Book, pp. 10-11]"))
    }
}

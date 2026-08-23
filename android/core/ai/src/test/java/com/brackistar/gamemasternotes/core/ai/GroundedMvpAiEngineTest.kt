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

    @Test
    fun generateSeparatesCitedExcerpts() = runTest {
        val response = GroundedMvpAiEngine().generate(
            AiRequest(
                prompt = "What is paradox?",
                context = """
                    [Core Book, pp. 10-11]
                    Paradox follows vulgar magic.

                    [Core Book, pp. 40-41]
                    A cabal is a group of mages.
                """.trimIndent(),
            ),
        )

        assertTrue(response.text.contains("[Core Book, pp. 10-11]\nParadox follows vulgar magic."))
        assertTrue(response.text.contains("[Core Book, pp. 40-41]\nA cabal is a group of mages."))
    }

    @Test
    fun evidenceBriefKeepsCompactQuestionRelevantSourceText() {
        val brief = EvidenceBriefBuilder.build(
            question = "What is paradox?",
            context = """
                [Core Book, pp. 10-11]
                Paradox follows vulgar magic. This sentence is less relevant.

                [Core Book, pp. 40-41]
                A cabal is a group of mages.
            """.trimIndent(),
        )

        assertEquals(listOf("Core Book, pp. 10-11", "Core Book, pp. 40-41"), brief.citationIds)
        assertTrue(brief.text.contains("Evidence"))
        assertTrue(brief.text.contains("Paradox follows vulgar magic."))
    }

    @Test
    fun evidenceBriefCanParseExistingNumberedBrief() {
        val brief = EvidenceBriefBuilder.build(
            question = "What is paradox?",
            context = """
                Evidence:
                1. [Core Book, pp. 10-11] Paradox follows vulgar magic.

                2. [Core Book, pp. 40-41] A cabal is a group of mages.
            """.trimIndent(),
        )

        assertEquals(listOf("Core Book, pp. 10-11", "Core Book, pp. 40-41"), brief.citationIds)
        assertEquals(2, brief.items.size)
    }

    @Test
    fun citationIdsCanBeExtractedFromNumberedLines() {
        val citations = """
            1. [Core Book, pp. 10-11] Paradox follows vulgar magic.
            2. [Core Book, pp. 40-41] A cabal is a group of mages.
        """.trimIndent().extractCitationIds()

        assertEquals(listOf("Core Book, pp. 10-11", "Core Book, pp. 40-41"), citations)
    }
}

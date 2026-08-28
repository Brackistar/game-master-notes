package com.brackistar.gamemasternotes.core.ai

import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class GroundedMvpAiEngineTest {
    @Test
    fun promptTemplateUsesSelectedAnswerMode() {
        val prompt = AiRequest(
            prompt = "Compare the two factions.",
            context = "[Core Book, p. 1]\nFaction A values secrecy.",
            answerMode = AnswerMode.Compare,
        ).withPromptTemplate(PromptStyle.Plain).prompt

        assertTrue(prompt.contains("Compare the requested subjects"))
    }

    @Test
    fun groundedAnswerRejectsUnsupportedCitation() {
        val evidence = EvidenceBriefBuilder.build(
            "What is paradox?",
            "[Core Book, p. 1]\nParadox follows vulgar magic.",
        )

        val quality = validateGroundedAnswer(
            "What is paradox?",
            "Paradox is harmless. [Other Book, p. 2] This is a complete answer.",
            evidence,
        )

        assertTrue(!quality.usable)
        assertEquals("unsupported-citation", quality.reason)
    }

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
    fun evidenceBriefPreservesMultipleRelevantParagraphs() {
        val brief = EvidenceBriefBuilder.build(
            question = "What does the Silver Ladder protect?",
            context = """
                [Core Book, pp. 10-11]
                The Silver Ladder guards the hidden library.
                Its members preserve the rites and laws of awakened society.
                The Silver Ladder trains archivists to protect dangerous lore.
            """.trimIndent(),
        )

        assertEquals(1, brief.items.size)
        assertTrue(brief.items.single().text.contains("guards the hidden library"))
        assertTrue(brief.items.single().text.contains("protect dangerous lore"))
    }

    @Test
    fun evidenceBriefKeepsPromptEvidenceWithinBound() {
        val brief = EvidenceBriefBuilder.build(
            question = "What is the Silver Ladder?",
            context = (1..4).joinToString("\n\n") { index ->
                "[Book, p. $index]\n" + "The Silver Ladder protects the library and trains archivists. ".repeat(20)
            },
        )

        assertTrue(brief.text.length <= 1_900)
        assertTrue(brief.items.isNotEmpty())
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

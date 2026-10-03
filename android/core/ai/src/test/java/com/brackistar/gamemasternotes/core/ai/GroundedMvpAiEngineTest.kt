package com.brackistar.gamemasternotes.core.ai

import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class GroundedMvpAiEngineTest {
    @Test
    fun promptTemplatePreservesOriginalQuestionAndStructuredEvidence() {
        val request = request(
            question = "Compare the two factions.",
            evidence = listOf(evidence("chunk-1", "Core Book, p. 1", "Faction A values secrecy.")),
            answerMode = AnswerMode.Compare,
        )

        val templated = request.withPromptTemplate(PromptStyle.Plain)

        assertEquals(request.originalQuestion, templated.originalQuestion)
        assertEquals("E1", templated.evidence.single().evidenceId)
        assertEquals(AiMessageRole.System, templated.messages.first().role)
        assertTrue(templated.messages.last().content.contains("[E1]"))
        assertTrue(templated.renderedPrompt!!.contains("Compare the requested subjects"))
        assertTrue(templated.renderedPrompt!!.contains("Question: Compare the two factions."))
    }

    @Test
    fun localModelEnginePassesSeparateQuestionEvidenceAndRenderedPrompt() = runTest {
        var received: AiRequest? = null
        val runtime = object : LocalModelRuntime {
            override suspend fun load(model: LocalModelProfile) = AiRuntimeStatus(model.model.id, false)
            override suspend fun unload() = Unit
            override suspend fun cancel() = Unit
            override suspend fun generate(model: LocalModelProfile, request: AiRequest): AiResponse {
                received = request
                return AiResponse(
                    requestId = request.requestId,
                    text = "Supported answer [E1]",
                    citationIds = listOf("chunk-1"),
                    provenance = AiAnswerProvenance.systemMessage("runtime"),
                )
            }
        }
        val engine = object : LocalModelAiEngine(LocalModelProfiles.Lfm25TinyQ2, runtime) {}
        val original = request(
            question = "What is paradox?",
            evidence = listOf(evidence("chunk-1", "Core Book, p. 1", "Paradox follows vulgar magic.")),
        )

        val response = engine.generate(original)

        assertEquals(original.originalQuestion, received?.originalQuestion)
        assertEquals("E1", received?.evidence?.single()?.evidenceId)
        assertEquals(listOf(AiMessageRole.System, AiMessageRole.User), received?.messages?.map { it.role })
        assertTrue(received?.renderedPrompt?.contains("Question: What is paradox?") == true)
        assertEquals(AiResponseMode.LocalModel, response.provenance.mode)
        assertFalse(response.provenance.isDeterministic)
        assertEquals(LocalModelProfiles.Lfm25TinyQ2.model.id, response.provenance.modelId)
    }

    @Test
    fun groundedAnswerRejectsUnsupportedCitation() {
        val brief = EvidenceBriefBuilder.build(
            "What is paradox?",
            listOf(evidence("chunk-1", "Core Book, p. 1", "Paradox follows vulgar magic.")),
        )

        val quality = validateGroundedAnswer(
            "What is paradox?",
            "Paradox is harmless. [E9] This is a complete answer.",
            brief,
        )

        assertTrue(!quality.usable)
        assertEquals("unsupported-citation", quality.reason)
    }

    @Test
    fun generateReportsMissingEvidence() = runTest {
        val response = GroundedMvpAiEngine().generate(request(question = "What is paradox?"))

        assertEquals("request-1", response.requestId)
        assertEquals(emptyList<String>(), response.citationIds)
        assertEquals(GroundedMvpAiEngine.NO_CLEAR_ANSWER_MESSAGE, response.text)
        assertEquals(AiResponseMode.DeterministicFallback, response.provenance.mode)
        assertTrue(response.provenance.isDeterministic)
    }

    @Test
    fun generateReturnsStableSourceIdsAndReadableLabels() = runTest {
        val response = GroundedMvpAiEngine().generate(
            request(
                question = "What is paradox?",
                evidence = listOf(evidence("chunk-42", "Core Book, pp. 10-11", "Paradox follows vulgar magic.")),
            ),
        )

        assertEquals(listOf("chunk-42"), response.citationIds)
        assertTrue(response.text.contains("Paradox follows vulgar magic."))
        assertTrue(response.text.contains("[Core Book, pp. 10-11]"))
        assertEquals(AiResponseMode.DeterministicFallback, response.provenance.mode)
    }

    @Test
    fun evidenceBriefKeepsQuestionRelevantSourcesAndIdentity() {
        val brief = EvidenceBriefBuilder.build(
            question = "What is paradox?",
            evidence = listOf(
                evidence("chunk-1", "Core Book, pp. 10-11", "Paradox follows vulgar magic. This sentence is less relevant."),
                evidence("chunk-2", "Core Book, pp. 40-41", "A cabal is a group of mages."),
            ),
        )

        assertEquals(listOf("chunk-1"), brief.sourceIds)
        assertEquals(listOf("Core Book, pp. 10-11"), brief.citationLabels)
        assertTrue(brief.text.contains("[E1]"))
        assertTrue(brief.text.contains("Paradox follows vulgar magic."))
    }

    @Test
    fun evidenceBriefPreservesMultipleRelevantLines() {
        val brief = EvidenceBriefBuilder.build(
            question = "What does the Silver Ladder protect?",
            evidence = listOf(
                evidence(
                    "chunk-1",
                    "Core Book, pp. 10-11",
                    """
                        The Silver Ladder guards the hidden library.
                        Its members preserve the rites and laws of awakened society.
                        The Silver Ladder trains archivists to protect dangerous lore.
                    """.trimIndent(),
                ),
            ),
        )

        assertEquals(1, brief.items.size)
        assertTrue(brief.items.single().text.contains("guards the hidden library"))
        assertTrue(brief.items.single().text.contains("protect dangerous lore"))
    }

    @Test
    fun questionIntentSeparatesIntentWordsFromSubjectTerms() {
        val intent = EvidenceBriefBuilder.analyze("How do I learn a rote?")

        assertEquals(QuestionIntentType.Procedure, intent.type)
        assertEquals(setOf("rote"), intent.subjectTerms)
        assertEquals(setOf("learn"), intent.supportingTerms)
        assertTrue("learn rote" in intent.subjectPhrases)
    }

    @Test
    fun evidenceBriefKeepsPromptEvidenceWithinBound() {
        val brief = EvidenceBriefBuilder.build(
            question = "What is the Silver Ladder?",
            evidence = (1..4).map { index ->
                evidence(
                    "chunk-$index",
                    "Book, p. $index",
                    "The Silver Ladder protects the library and trains archivists. ".repeat(20),
                )
            },
        )

        assertTrue(brief.text.length <= 1_900)
        assertTrue(brief.items.isNotEmpty())
    }

    @Test
    fun evidenceBriefDropsIrrelevantEvidenceWhenSubjectTermDoesNotMatch() {
        val brief = EvidenceBriefBuilder.build(
            question = "How do I learn a rote?",
            evidence = listOf(
                evidence("chunk-1", "Core Book, p. 10", "A mage can learn sympathetic tracking through prolonged study."),
                evidence("chunk-2", "Core Book, p. 11", "Scrying lets the caster witness a distant subject."),
            ),
        )

        assertTrue(brief.isEmpty)
    }

    @Test
    fun groundedEngineFallsBackGracefullyWhenNoEvidenceClearsThreshold() = runTest {
        val response = GroundedMvpAiEngine().generate(
            request(
                question = "How do I learn a rote?",
                evidence = listOf(
                    evidence("chunk-1", "Core Book, p. 10", "A mage can learn sympathetic tracking through prolonged study."),
                    evidence("chunk-2", "Core Book, p. 11", "Scrying lets the caster witness a distant subject."),
                ),
            ),
        )

        assertEquals(GroundedMvpAiEngine.NO_CLEAR_ANSWER_MESSAGE, response.text)
        assertEquals(emptyList<String>(), response.citationIds)
        assertTrue(response.provenance.isDeterministic)
    }

    @Test
    fun citationLabelsCanBeExtractedFromNumberedLines() {
        val citations = """
            1. [Core Book, pp. 10-11] Paradox follows vulgar magic.
            2. [Core Book, pp. 40-41] A cabal is a group of mages.
        """.trimIndent().extractCitationIds()

        assertEquals(listOf("Core Book, pp. 10-11", "Core Book, pp. 40-41"), citations)
    }

    @Test
    fun multipleEvidenceIdsCanBeExtractedFromOneLine() {
        assertEquals(listOf("E1", "E2"), "A comparison [E1] with an exception [E2].".extractCitationIds())
    }

    private fun request(
        question: String,
        evidence: List<AiEvidence> = emptyList(),
        answerMode: AnswerMode = AnswerMode.Explain,
    ) = AiRequest(
        requestId = "request-1",
        originalQuestion = question,
        evidence = evidence,
        answerMode = answerMode,
    )

    private fun evidence(sourceId: String, citationLabel: String, text: String) =
        AiEvidence(sourceId = sourceId, citationLabel = citationLabel, text = text)
}

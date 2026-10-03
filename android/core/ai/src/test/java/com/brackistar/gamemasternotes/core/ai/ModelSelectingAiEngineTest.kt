package com.brackistar.gamemasternotes.core.ai

import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class ModelSelectingAiEngineTest {
    @Test
    fun availableModelsShowsOnlyFallbackWhenNoModelFilesAreInstalled() = runTest {
        val engine = ModelSelectingAiEngine(
            deviceProfile = DeviceAiProfile(
                totalRamMb = 3_500,
                supportedAbis = listOf("arm64-v8a"),
                isLowRamDevice = true,
            ),
        )

        val modelIds = engine.availableModels().map { it.id }

        assertEquals(listOf("grounded-mvp"), modelIds)
    }

    @Test
    fun unsupportedInstalledModelDoesNotAppear() = runTest {
        val engine = ModelSelectingAiEngine(
            deviceProfile = DeviceAiProfile(
                totalRamMb = 3_500,
                supportedAbis = listOf("arm64-v8a"),
                isLowRamDevice = true,
            ),
            isModelFileInstalled = { it.model.id == "qwen3-1.7b-instruct-q4" },
        )

        val modelIds = engine.availableModels().map { it.id }

        assertTrue("qwen3-1.7b-instruct-q4" !in modelIds)
    }

    @Test
    fun installedCompatibleLfmModelCanBeSelected() = runTest {
        val engine = ModelSelectingAiEngine(
            deviceProfile = DeviceAiProfile(
                totalRamMb = 8_000,
                supportedAbis = listOf("arm64-v8a"),
                isLowRamDevice = false,
            ),
            isModelFileInstalled = { it.model.id == "lfm2.5-350m-q2-tiny" },
        )

        val lfm = engine.availableModels().first { it.id == "lfm2.5-350m-q2-tiny" }

        assertEquals(AiModelAvailability.Ready, lfm.availability)
    }

    @Test
    fun smallestLfmIsFirstInstalledLocalModelByDefaultOrder() = runTest {
        val engine = ModelSelectingAiEngine(
            deviceProfile = DeviceAiProfile(
                totalRamMb = 8_000,
                supportedAbis = listOf("arm64-v8a"),
                isLowRamDevice = false,
            ),
            isModelFileInstalled = { it.model.id == "lfm2.5-350m-q2-tiny" || it.model.id == "lfm2.5-350m-q4" || it.model.id == "gemma-3-1b-it-q4" },
        )

        val firstReadyLocalModel = engine.availableModels()
            .filterNot { it.isFallback }
            .first { it.availability == AiModelAvailability.Ready }

        assertEquals("lfm2.5-350m-q2-tiny", firstReadyLocalModel.id)
    }

    @Test
    fun fallbackStillAnswersFromRetrievedContext() = runTest {
        val engine = ModelSelectingAiEngine(
            deviceProfile = DeviceAiProfile(
                totalRamMb = 2_000,
                supportedAbis = listOf("arm64-v8a"),
                isLowRamDevice = true,
            ),
        )

        engine.load("grounded-mvp")
        val response = engine.generate(
            AiRequest(
                requestId = "request-1",
                originalQuestion = "What is paradox?",
                evidence = listOf(
                    AiEvidence(
                        sourceId = "chunk-1",
                        citationLabel = "Core Book, p. 10",
                        text = "Paradox follows vulgar magic.",
                    ),
                ),
            ),
        )

        assertEquals(listOf("chunk-1"), response.citationIds)
        assertTrue(response.text.contains("Paradox follows vulgar magic."))
        assertEquals(AiResponseMode.DeterministicFallback, response.provenance.mode)
        assertTrue(response.provenance.isDeterministic)
    }

    @Test
    fun selectedLocalModelAnswersAreMarkedAsLocalModelResponses() = runTest {
        val runtime = object : LocalModelRuntime {
            override suspend fun load(model: LocalModelProfile) = AiRuntimeStatus(model.model.id, false)
            override suspend fun unload() = Unit
            override suspend fun cancel() = Unit
            override suspend fun generate(model: LocalModelProfile, request: AiRequest): AiResponse =
                AiResponse(
                    requestId = request.requestId,
                    text = "Supported answer [E1]",
                    citationIds = listOf("chunk-1"),
                    provenance = AiAnswerProvenance.systemMessage("runtime"),
                )
        }
        val engine = ModelSelectingAiEngine(
            deviceProfile = DeviceAiProfile(
                totalRamMb = 8_000,
                supportedAbis = listOf("arm64-v8a"),
                isLowRamDevice = false,
            ),
            runtime = runtime,
            isModelFileInstalled = { it.model.id == LocalModelProfiles.Lfm25TinyQ2.model.id },
        )

        engine.load(LocalModelProfiles.Lfm25TinyQ2.model.id)
        val response = engine.generate(
            AiRequest(
                requestId = "request-2",
                originalQuestion = "What is paradox?",
                evidence = listOf(
                    AiEvidence(
                        sourceId = "chunk-1",
                        citationLabel = "Core Book, p. 10",
                        text = "Paradox follows vulgar magic.",
                    ),
                ),
            ),
        )

        assertEquals(AiResponseMode.LocalModel, response.provenance.mode)
        assertFalse(response.provenance.isDeterministic)
        assertEquals(LocalModelProfiles.Lfm25TinyQ2.model.id, response.provenance.modelId)
    }
}

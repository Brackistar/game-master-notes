package com.brackistar.gamemasternotes.core.ai

import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
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
                prompt = "What is paradox?",
                context = "[Core Book, p. 10]\nParadox follows vulgar magic.",
            ),
        )

        assertEquals(listOf("Core Book, p. 10"), response.citationIds)
        assertTrue(response.text.contains("Paradox follows vulgar magic."))
    }
}

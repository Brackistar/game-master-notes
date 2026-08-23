package com.brackistar.gamemasternotes.core.ai

import android.util.Log

class ModelSelectingAiEngine(
    private val deviceProfile: DeviceAiProfile,
    runtime: LocalModelRuntime = MissingLocalModelRuntime(),
    private val fallback: AiEngine = GroundedMvpAiEngine(),
    private val isModelFileInstalled: (LocalModelProfile) -> Boolean = { false },
) : AiEngine {
    private val engines: Map<String, AiEngine> =
        mapOf(
            LocalModelProfiles.Lfm25TinyQ2.model.id to Lfm25TinyQ2AiEngine(runtime),
            LocalModelProfiles.Lfm25FastQ4.model.id to Lfm25FastQ4AiEngine(runtime),
            LocalModelProfiles.Lfm25TinyQ3.model.id to Lfm25TinyQ3AiEngine(runtime),
            LocalModelProfiles.Lfm25.model.id to Lfm25AiEngine(runtime),
            LocalModelProfiles.Gemma3.model.id to Gemma3AiEngine(runtime),
            LocalModelProfiles.Qwen3.model.id to Qwen3AiEngine(runtime),
            LocalModelProfiles.Phi4.model.id to Phi4AiEngine(runtime),
        )

    private var selectedEngine: AiEngine = fallback

    override suspend fun availableModels(): List<AiModel> {
        val models = fallback.availableModels() + LocalModelProfiles.Lfm25Family
            .filter { deviceProfile.canRun(it.model) }
            .filter { isModelFileInstalled(it) }
            .map { profile ->
                profile.model.copy(
                    availability = AiModelAvailability.Ready,
                )
            }
        logDebug(
            TAG,
            "Installed models discovered total=${models.size} local=${models.count { !it.isFallback }} lowRam=${deviceProfile.isLowRamDevice} ramMb=${deviceProfile.totalRamMb} supportedAbis=${deviceProfile.supportedAbis.joinToString()}",
        )
        return models
    }

    override suspend fun load(modelId: String): AiRuntimeStatus {
        val startedAt = System.currentTimeMillis()
        logInfo(TAG, "Load requested modelId=$modelId")
        val model = availableModels().firstOrNull { it.id == modelId }
        require(model != null && model.availability == AiModelAvailability.Ready) {
            "${model?.displayName ?: modelId} is not ready: ${model?.availability ?: AiModelAvailability.UnsupportedDevice}."
        }
        selectedEngine = engines[modelId] ?: fallback
        return selectedEngine.load(modelId).also {
            logInfo(TAG, "Load completed modelId=$modelId loadedModelId=${it.loadedModelId} elapsedMs=${System.currentTimeMillis() - startedAt}")
        }
    }

    override suspend fun unload() {
        logInfo(TAG, "Unload requested")
        selectedEngine.unload()
    }

    override suspend fun generate(request: AiRequest): AiResponse {
        val startedAt = System.currentTimeMillis()
        logInfo(TAG, "Generate routed promptChars=${request.prompt.length} contextChars=${request.context.length}")
        return selectedEngine.generate(request).also {
            logInfo(TAG, "Generate routed finished outputChars=${it.text.length} citationCount=${it.citationIds.size} elapsedMs=${System.currentTimeMillis() - startedAt}")
        }
    }

    override suspend fun cancel() {
        logWarn(TAG, "Cancel requested")
        selectedEngine.cancel()
    }

    companion object {
        private const val TAG = "GmnAiSelector"
    }
}

private fun logDebug(tag: String, message: String) {
    runCatching { Log.d(tag, message) }
}

private fun logInfo(tag: String, message: String) {
    runCatching { Log.i(tag, message) }
}

private fun logWarn(tag: String, message: String) {
    runCatching { Log.w(tag, message) }
}

class MissingLocalModelRuntime : LocalModelRuntime {
    override suspend fun load(model: LocalModelProfile): AiRuntimeStatus =
        AiRuntimeStatus(loadedModelId = model.model.id, isGenerating = false)

    override suspend fun unload() = Unit

    override suspend fun generate(model: LocalModelProfile, request: AiRequest): AiResponse {
        error(
            "${model.model.displayName} is available for this device, but native model inference is not installed yet. " +
                "Add the llama.cpp runtime and a matching ${model.model.quantization ?: "quantized"} model file.",
        )
    }

    override suspend fun cancel() = Unit
}

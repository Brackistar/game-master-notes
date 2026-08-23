package com.brackistar.gamemasternotes.core.ai

interface LocalModelRuntime {
    suspend fun load(model: LocalModelProfile): AiRuntimeStatus
    suspend fun unload()
    suspend fun generate(model: LocalModelProfile, request: AiRequest): AiResponse
    suspend fun cancel()
}

abstract class LocalModelAiEngine(
    private val profile: LocalModelProfile,
    private val runtime: LocalModelRuntime,
) : AiEngine {
    override suspend fun availableModels(): List<AiModel> = listOf(profile.model)

    override suspend fun load(modelId: String): AiRuntimeStatus {
        require(modelId == profile.model.id) { "Unsupported model id: $modelId" }
        return runtime.load(profile)
    }

    override suspend fun unload() = runtime.unload()

    override suspend fun generate(request: AiRequest): AiResponse =
        runtime.generate(profile, request.withPromptTemplate(profile.promptStyle))

    override suspend fun cancel() = runtime.cancel()
}

class Qwen3AiEngine(runtime: LocalModelRuntime) : LocalModelAiEngine(LocalModelProfiles.Qwen3, runtime)

class Gemma3AiEngine(runtime: LocalModelRuntime) : LocalModelAiEngine(LocalModelProfiles.Gemma3, runtime)

class Phi4AiEngine(runtime: LocalModelRuntime) : LocalModelAiEngine(LocalModelProfiles.Phi4, runtime)

class Lfm25AiEngine(runtime: LocalModelRuntime) : LocalModelAiEngine(LocalModelProfiles.Lfm25, runtime)

class Lfm25FastQ4AiEngine(runtime: LocalModelRuntime) : LocalModelAiEngine(LocalModelProfiles.Lfm25FastQ4, runtime)

class Lfm25TinyQ3AiEngine(runtime: LocalModelRuntime) : LocalModelAiEngine(LocalModelProfiles.Lfm25TinyQ3, runtime)

class Lfm25TinyQ2AiEngine(runtime: LocalModelRuntime) : LocalModelAiEngine(LocalModelProfiles.Lfm25TinyQ2, runtime)

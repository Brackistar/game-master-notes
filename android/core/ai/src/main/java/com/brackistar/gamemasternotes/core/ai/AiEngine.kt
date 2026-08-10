package com.brackistar.gamemasternotes.core.ai

interface AiEngine {
    suspend fun availableModels(): List<AiModel>
    suspend fun load(modelId: String): AiRuntimeStatus
    suspend fun unload()
    suspend fun generate(request: AiRequest): AiResponse
    suspend fun cancel()
}

data class AiModel(
    val id: String,
    val displayName: String,
    val fileSizeBytes: Long?,
    val quantization: String?,
)

data class AiRequest(
    val prompt: String,
    val context: String,
)

data class AiResponse(
    val text: String,
    val citationIds: List<String>,
)

data class AiRuntimeStatus(
    val loadedModelId: String?,
    val isGenerating: Boolean,
)

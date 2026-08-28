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
    val minimumRamMb: Long = 0,
    val description: String = "",
    val isFallback: Boolean = false,
    val availability: AiModelAvailability = AiModelAvailability.Ready,
)

enum class AiModelAvailability {
    Ready,
    MissingModelFile,
    UnsupportedDevice,
}

data class AiRequest(
    val prompt: String,
    val context: String,
    val answerMode: AnswerMode = AnswerMode.Explain,
)

enum class AnswerMode(
    val displayName: String,
    val description: String,
    val promptInstruction: String,
) {
    Lookup(
        displayName = "Look up a rule",
        description = "Find and explain what the books say.",
        promptInstruction = "Give a concise factual answer based on the source material.",
    ),
    Explain(
        displayName = "Explain",
        description = "Synthesize several passages into a clear explanation.",
        promptInstruction = "Synthesize the relevant passages into a clear explanation.",
    ),
    Summarize(
        displayName = "Summarize",
        description = "Condense the retrieved source material.",
        promptInstruction = "Summarize the relevant source material without adding new facts.",
    ),
    Brainstorm(
        displayName = "Brainstorm",
        description = "Suggest ideas while separating them from book facts.",
        promptInstruction = "Start with what the sources establish, then label any creative suggestions as ideas.",
    ),
    Compare(
        displayName = "Compare",
        description = "Organize similarities and differences from the sources.",
        promptInstruction = "Compare the requested subjects using only supported differences and similarities.",
    ),
}

data class AiResponse(
    val text: String,
    val citationIds: List<String>,
)

data class AiRuntimeStatus(
    val loadedModelId: String?,
    val isGenerating: Boolean,
)

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

class GroundedMvpAiEngine : AiEngine {
    override suspend fun availableModels(): List<AiModel> =
        listOf(
            AiModel(
                id = MODEL_ID,
                displayName = "Grounded MVP responder",
                fileSizeBytes = null,
                quantization = null,
            ),
        )

    override suspend fun load(modelId: String): AiRuntimeStatus =
        AiRuntimeStatus(loadedModelId = MODEL_ID, isGenerating = false)

    override suspend fun unload() = Unit

    override suspend fun generate(request: AiRequest): AiResponse {
        if (request.context.isBlank()) {
            return AiResponse(
                text = "I could not find relevant passages in the loaded books for that question.",
                citationIds = emptyList(),
            )
        }

        val sections = request.context
            .split("\n\n")
            .filter { it.isNotBlank() }
            .take(4)
        val answer = buildString {
            appendLine("Based on the loaded books:")
            sections.forEachIndexed { index, section ->
                val citation = section.substringBefore("\n").removePrefix("[")
                    .substringBefore("]")
                    .ifBlank { "source ${index + 1}" }
                val body = section.lines().drop(1).joinToString(" ").trim()
                appendLine()
                append("- ")
                append(body.take(420))
                if (body.length > 420) append("...")
                append(" [")
                append(citation)
                append("]")
            }
        }.trim()

        return AiResponse(
            text = answer,
            citationIds = sections.mapIndexed { index, section ->
                section.substringBefore("\n").removePrefix("[")
                    .substringBefore("]")
                    .ifBlank { "source-${index + 1}" }
            },
        )
    }

    override suspend fun cancel() = Unit

    companion object {
        private const val MODEL_ID = "grounded-mvp"
    }
}

package com.brackistar.gamemasternotes.core.ai

class GroundedMvpAiEngine : AiEngine {
    override suspend fun availableModels(): List<AiModel> =
        listOf(
            AiModel(
                id = MODEL_ID,
                displayName = "Grounded MVP responder",
                fileSizeBytes = null,
                quantization = null,
                description = "Deterministic fallback that summarizes retrieved chunks.",
                isFallback = true,
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

        val evidenceBrief = EvidenceBriefBuilder.build(request.prompt, request.context)

        return AiResponse(
            text = evidenceBrief.toReadableAnswer(),
            citationIds = evidenceBrief.citationIds,
        )
    }

    override suspend fun cancel() = Unit

    companion object {
        const val MODEL_ID = "grounded-mvp"
    }
}

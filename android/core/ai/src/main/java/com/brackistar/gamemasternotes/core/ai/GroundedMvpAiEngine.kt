package com.brackistar.gamemasternotes.core.ai

class GroundedMvpAiEngine : AiEngine {
    private val fallbackModel = AiModel(
        id = MODEL_ID,
        displayName = "Grounded MVP responder",
        fileSizeBytes = null,
        quantization = null,
        description = "Deterministic fallback that summarizes retrieved chunks.",
        isFallback = true,
    )

    override suspend fun availableModels(): List<AiModel> =
        listOf(fallbackModel)

    override suspend fun load(modelId: String): AiRuntimeStatus =
        AiRuntimeStatus(loadedModelId = MODEL_ID, isGenerating = false)

    override suspend fun unload() = Unit

    override suspend fun generate(request: AiRequest): AiResponse {
        if (request.evidence.isEmpty()) {
            return AiResponse(
                requestId = request.requestId,
                text = NO_CLEAR_ANSWER_MESSAGE,
                citationIds = emptyList(),
                provenance = AiAnswerProvenance.deterministicFallback(fallbackModel),
            )
        }

        val evidenceBrief = EvidenceBriefBuilder.build(request.originalQuestion, request.evidence)
        if (evidenceBrief.isEmpty) {
            return AiResponse(
                requestId = request.requestId,
                text = NO_CLEAR_ANSWER_MESSAGE,
                citationIds = emptyList(),
                provenance = AiAnswerProvenance.deterministicFallback(fallbackModel),
            )
        }

        return AiResponse(
            requestId = request.requestId,
            text = evidenceBrief.toReadableAnswer(),
            citationIds = evidenceBrief.sourceIds,
            provenance = AiAnswerProvenance.deterministicFallback(fallbackModel),
        )
    }

    override suspend fun cancel() = Unit

    companion object {
        const val MODEL_ID = "grounded-mvp"
        const val NO_CLEAR_ANSWER_MESSAGE =
            "I couldn't find a clear answer to that in the loaded books. Try rephrasing your question or check the Library for related sourcebooks."
    }
}

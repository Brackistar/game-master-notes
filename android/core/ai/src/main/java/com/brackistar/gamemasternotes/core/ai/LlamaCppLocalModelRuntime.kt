package com.brackistar.gamemasternotes.core.ai

import android.util.Log
import com.brackistar.gamemasternotes.core.domain.diagnostics.DiagnosticEvent
import com.brackistar.gamemasternotes.core.domain.diagnostics.DiagnosticsJournal
import com.brackistar.gamemasternotes.core.domain.diagnostics.NoOpDiagnosticsJournal
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext
import java.io.File
import kotlin.math.max

class LlamaCppLocalModelRuntime(
    private val modelsDirectory: File,
    private val deviceProfile: DeviceAiProfile,
    private val bridge: NativeModelBridge = LlamaCppBridge(),
    private val diagnosticsJournal: DiagnosticsJournal = NoOpDiagnosticsJournal,
) : LocalModelRuntime {
    private var loadedProfile: LocalModelProfile? = null
    private val mutex = Mutex()

    override suspend fun load(model: LocalModelProfile): AiRuntimeStatus = withContext(Dispatchers.Default) {
        mutex.withLock {
            loadLocked(model)
        }
    }

    private fun loadLocked(model: LocalModelProfile): AiRuntimeStatus {
        val modelFile = File(modelsDirectory, model.modelFileName)
        Log.i(
            TAG,
            "Load requested model=${model.model.id} file=${model.modelFileName} exists=${modelFile.exists()} readable=${modelFile.canRead()} sizeBytes=${modelFile.length()} directory=${modelsDirectory.path}",
        )
        require(modelFile.exists()) {
            "${model.model.displayName} is compatible with this device, but ${model.modelFileName} was not found in ${modelsDirectory.path}."
        }
        require(modelFile.canRead()) {
            "${model.modelFileName} exists but cannot be read."
        }

        if (loadedProfile?.model?.id != model.model.id) {
            val startedAt = System.currentTimeMillis()
            bridge.unload()
            val threadCount = recommendedThreadCount()
            bridge.load(
                modelPath = modelFile.absolutePath,
                threadCount = threadCount,
                contextTokens = model.generation.contextTokens,
                batchTokens = model.generation.batchTokens,
            )
            Log.d(
                TAG,
                "Loaded ${model.model.id} threads=$threadCount contextTokens=${model.generation.contextTokens} elapsedMs=${System.currentTimeMillis() - startedAt}",
            )
            loadedProfile = model
        } else {
            Log.d(TAG, "Model already loaded model=${model.model.id}")
        }
        return AiRuntimeStatus(loadedModelId = model.model.id, isGenerating = false)
    }

    override suspend fun unload() = withContext(Dispatchers.Default) {
        mutex.withLock {
            Log.i(TAG, "Runtime unload requested loadedModel=${loadedProfile?.model?.id}")
            bridge.unload()
            loadedProfile = null
        }
    }

    override suspend fun generate(model: LocalModelProfile, request: AiRequest): AiResponse = withContext(Dispatchers.Default) {
        mutex.withLock {
            if (loadedProfile?.model?.id != model.model.id) {
                loadLocked(model)
            }

            if (request.evidence.isEmpty()) {
                Log.i(TAG, "Generation skipped requestId=${request.requestId} because evidence is empty model=${model.model.id}")
                return@withLock AiResponse(
                    requestId = request.requestId,
                    text = GroundedMvpAiEngine.NO_CLEAR_ANSWER_MESSAGE,
                    citationIds = emptyList(),
                    provenance = AiAnswerProvenance.deterministicFallback(model.model),
                )
            }

            val systemMessage = request.messages.singleOrNull { it.role == AiMessageRole.System }?.content
                ?: error("A system message is required for local generation.")
            val userMessage = request.messages.singleOrNull { it.role == AiMessageRole.User }?.content
                ?: error("A user message is required for local generation.")
            val evidenceBrief = EvidenceBrief(request.evidence)
            val startedAt = System.currentTimeMillis()
            recordDiagnostic(
                requestId = request.requestId,
                stage = "native-generation",
                outcome = "started",
                fields = mapOf(
                    "modelId" to model.model.id,
                    "context" to model.generation.contextTokens.toString(),
                    "output" to model.generation.outputTokens.toString(),
                    "deadlineMs" to model.generation.deadlineMs.toString(),
                    "selectedEvidenceIds" to evidenceBrief.sourceIds.joinToString(","),
                ),
            )
            Log.d(
                TAG,
                "Generating requestId=${request.requestId} model=${model.model.id} evidenceCount=${request.evidence.size} maxTokens=${model.generation.outputTokens} timeoutMs=${model.generation.deadlineMs}",
            )
            val nativeResult = bridge.generate(
                requestId = request.requestId,
                systemMessage = systemMessage,
                userMessage = userMessage,
                profile = model.generation,
            )
            if (nativeResult.stopReason == NativeStopReason.Cancelled) {
                throw CancellationException("Local generation was cancelled.")
            }
            val rawGenerated = nativeResult.text.trim()
            val generated = rawGenerated.removePromptEchoMarkers().trim()
            val elapsedMs = System.currentTimeMillis() - startedAt
            val quality = validateGroundedAnswer(request.originalQuestion, generated, evidenceBrief)
            val shouldFallback = !nativeResult.completedNormally || !quality.usable
            recordDiagnostic(
                requestId = request.requestId,
                stage = "validation",
                outcome = if (shouldFallback) "fallback" else "accepted",
                elapsedMs = elapsedMs,
                fields = mapOf(
                    "modelId" to model.model.id,
                    "qualityReason" to (quality.reason ?: "usable"),
                    "stopReason" to nativeResult.stopReason.name,
                    "promptTokens" to nativeResult.promptTokens.toString(),
                    "generatedTokens" to nativeResult.generatedTokens.toString(),
                    "promptEvalMs" to nativeResult.promptEvalMs.toString(),
                    "firstTokenMs" to (nativeResult.firstTokenMs?.toString() ?: "none"),
                    "generationMs" to nativeResult.generationMs.toString(),
                    "templateHash" to nativeResult.templateHash,
                    "responseCitationIds" to generated.extractCitationIds().joinToString(","),
                ),
            )
            Log.i(
                TAG,
                "Generated requestId=${request.requestId} model=${model.model.id} stopReason=${nativeResult.stopReason} rawOutputChars=${rawGenerated.length} outputChars=${generated.length} fallback=$shouldFallback reason=${quality.reason} elapsedMs=$elapsedMs",
            )
            val responseText = if (shouldFallback) {
                evidenceBrief.toReadableAnswer()
            } else {
                generated
            }
            AiResponse(
                requestId = request.requestId,
                text = responseText,
                citationIds = if (shouldFallback) {
                    evidenceBrief.sourceIds
                } else {
                    generated.extractCitationIds()
                        .flatMap { evidenceId -> evidenceBrief.items.filter { it.evidenceId == evidenceId } }
                        .map { it.sourceId }
                        .distinct()
                },
                provenance = if (shouldFallback) {
                    AiAnswerProvenance.deterministicFallback(model.model)
                } else {
                    AiAnswerProvenance.localModel(model.model)
                },
            )
        }
    }

    override suspend fun cancel() {
        Log.w(TAG, "Runtime cancel requested loadedModel=${loadedProfile?.model?.id}")
        bridge.cancel()
    }

    private suspend fun recordDiagnostic(
        requestId: String,
        stage: String,
        outcome: String,
        elapsedMs: Long? = null,
        fields: Map<String, String> = emptyMap(),
    ) {
        runCatching {
            diagnosticsJournal.record(
                DiagnosticEvent(
                    requestId = requestId,
                    timestampEpochMillis = System.currentTimeMillis(),
                    stage = stage,
                    outcome = outcome,
                    elapsedMs = elapsedMs,
                    fields = fields,
                ),
            )
        }.onFailure { error ->
            Log.w(TAG, "Diagnostics event failed requestId=$requestId stage=$stage", error)
        }
    }

    private fun recommendedThreadCount(): Int {
        val cores = Runtime.getRuntime().availableProcessors()
        val ceiling = if (deviceProfile.isLowRamDevice) 1 else 2
        val threads = max(1, minOf(cores - 1, ceiling))
        Log.d(TAG, "Recommended threads=$threads cores=$cores lowRam=${deviceProfile.isLowRamDevice}")
        return threads
    }

    companion object {
        private const val TAG = "GmnLlamaRuntime"
    }
}

private fun String.removePromptEchoMarkers(): String =
    lineSequence()
        .filterNot { line ->
            val trimmed = line.trim()
            trimmed.equals("Answer:", ignoreCase = true) ||
                trimmed.startsWith("Question:", ignoreCase = true) ||
                trimmed.startsWith("Evidence:", ignoreCase = true) ||
                trimmed.startsWith("<|")
        }
        .joinToString("\n")

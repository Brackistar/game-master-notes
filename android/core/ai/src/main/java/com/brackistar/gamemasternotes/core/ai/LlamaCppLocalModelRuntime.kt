package com.brackistar.gamemasternotes.core.ai

import android.util.Log
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext
import java.io.File
import kotlin.math.max

class LlamaCppLocalModelRuntime(
    private val modelsDirectory: File,
    private val deviceProfile: DeviceAiProfile,
    private val bridge: LlamaCppBridge = LlamaCppBridge(),
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
                contextTokens = CONTEXT_TOKENS,
            )
            Log.d(
                TAG,
                "Loaded ${model.model.id} threads=$threadCount contextTokens=$CONTEXT_TOKENS elapsedMs=${System.currentTimeMillis() - startedAt}",
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

            if (request.context.isBlank()) {
                Log.i(TAG, "Generation skipped because evidence context is blank model=${model.model.id}")
                return@withLock AiResponse(
                    text = "I could not find relevant passages in the loaded books for that question.",
                    citationIds = emptyList(),
                )
            }

            val prompt = request.prompt
            val evidenceBrief = EvidenceBriefBuilder.build(question = request.prompt, context = request.context)
            val startedAt = System.currentTimeMillis()
            Log.d(
                TAG,
                "Generating model=${model.model.id} promptChars=${prompt.length} contextChars=${request.context.length} maxTokens=$MAX_RESPONSE_TOKENS timeoutMs=$MAX_GENERATION_MILLIS",
            )
            val rawGenerated = bridge.generate(
                prompt = prompt,
                maxTokens = MAX_RESPONSE_TOKENS,
                maxDurationMillis = MAX_GENERATION_MILLIS,
            ).trim()
            val generated = rawGenerated.removePromptEchoMarkers().trim()
            val elapsedMs = System.currentTimeMillis() - startedAt
            val shouldFallback = generated.shouldUseGroundedFallback(evidenceBrief.citationIds)
            Log.i(
                TAG,
                "Generated model=${model.model.id} rawOutputChars=${rawGenerated.length} outputChars=${generated.length} fallback=$shouldFallback elapsedMs=$elapsedMs",
            )
            val responseText = if (shouldFallback) {
                evidenceBrief.toReadableAnswer()
            } else {
                generated
            }
            AiResponse(
                text = responseText,
                citationIds = evidenceBrief.citationIds,
            )
        }
    }

    override suspend fun cancel() {
        Log.w(TAG, "Runtime cancel requested loadedModel=${loadedProfile?.model?.id}")
        bridge.cancel()
    }

    private fun recommendedThreadCount(): Int {
        val cores = Runtime.getRuntime().availableProcessors()
        val ceiling = if (deviceProfile.isLowRamDevice) 1 else 2
        val threads = max(1, minOf(cores - 1, ceiling))
        Log.d(TAG, "Recommended threads=$threads cores=$cores lowRam=${deviceProfile.isLowRamDevice}")
        return threads
    }

    companion object {
        private const val CONTEXT_TOKENS = 512
        private const val MAX_RESPONSE_TOKENS = 64
        private const val MAX_GENERATION_MILLIS = 12_000L
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

private fun String.shouldUseGroundedFallback(citationIds: List<String>): Boolean {
    val normalized = replace(Regex("""\s+"""), " ").trim()
    if (normalized.length < MIN_USABLE_OUTPUT_CHARS) return true
    if (citationIds.isNotEmpty() && citationIds.none { normalized.contains("[$it]") }) return true
    val letterCount = normalized.count { it.isLetter() }
    if (letterCount < MIN_USABLE_LETTERS) return true
    val asciiLetters = normalized.count { it in 'A'..'Z' || it in 'a'..'z' }
    return asciiLetters * 2 < letterCount
}

private const val MIN_USABLE_OUTPUT_CHARS = 24
private const val MIN_USABLE_LETTERS = 12

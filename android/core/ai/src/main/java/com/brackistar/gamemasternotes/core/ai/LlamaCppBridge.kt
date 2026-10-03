package com.brackistar.gamemasternotes.core.ai

import android.os.Looper
import android.util.Log

enum class NativeStopReason {
    Eog, MaxTokens, Timeout, Cancelled, ContextOverflow, UnsupportedTemplate, PromptDecodeError, NativeError,
}

data class NativeGenerationResult(
    val text: String,
    val stopReason: NativeStopReason,
    val promptTokens: Int,
    val generatedTokens: Int,
    val promptEvalMs: Long,
    val firstTokenMs: Long?,
    val generationMs: Long,
    val totalMs: Long,
    val templateHash: String,
) {
    val completedNormally: Boolean
        get() = stopReason == NativeStopReason.Eog || stopReason == NativeStopReason.MaxTokens
}

interface NativeModelBridge {
    fun load(modelPath: String, threadCount: Int, contextTokens: Int, batchTokens: Int)
    fun generate(
        requestId: String,
        systemMessage: String,
        userMessage: String,
        profile: GenerationProfile,
    ): NativeGenerationResult
    fun cancel()
    fun unload()
}

class LlamaCppBridge : NativeModelBridge {
    @Volatile
    private var nativeHandle: Long = 0

    override fun load(modelPath: String, threadCount: Int, contextTokens: Int, batchTokens: Int) {
        val startedAt = System.currentTimeMillis()
        Log.i(
            TAG,
            "JNI load started pathHash=${modelPath.hashCode()} threads=$threadCount contextTokens=$contextTokens batchTokens=$batchTokens",
        )
        nativeHandle = nativeLoad(modelPath, threadCount, contextTokens, batchTokens)
        Log.i(TAG, "JNI load finished handle=$nativeHandle elapsedMs=${System.currentTimeMillis() - startedAt}")
    }

    override fun generate(
        requestId: String,
        systemMessage: String,
        userMessage: String,
        profile: GenerationProfile,
    ): NativeGenerationResult {
        check(nativeHandle != 0L) { "No llama.cpp model is loaded." }
        check(Looper.myLooper() != Looper.getMainLooper()) {
            "llama.cpp generation must not run on the main thread."
        }
        val values = nativeGenerate(
            handle = nativeHandle,
            requestId = requestId,
            systemMessage = systemMessage,
            userMessage = userMessage,
            maxTokens = profile.outputTokens,
            maxDurationMillis = profile.deadlineMs,
            temperature = profile.temperature,
            topK = profile.topK,
            repetitionPenalty = profile.repetitionPenalty,
            seed = profile.seed,
            greedy = profile.greedy,
        )
        check(values.size == RESULT_FIELD_COUNT) { "Native generation returned ${values.size} fields." }
        return NativeGenerationResult(
            text = values[0],
            stopReason = values[1].toNativeStopReason(),
            promptTokens = values[2].toInt(),
            generatedTokens = values[3].toInt(),
            promptEvalMs = values[4].toLong(),
            firstTokenMs = values[5].toLong().takeIf { it >= 0 },
            generationMs = values[6].toLong(),
            totalMs = values[7].toLong(),
            templateHash = values[8],
        )
    }

    override fun cancel() {
        val handle = nativeHandle
        Log.w(TAG, "JNI cancel requested handle=$handle")
        if (handle != 0L) nativeCancel(handle)
    }

    override fun unload() {
        if (nativeHandle != 0L) {
            val handle = nativeHandle
            nativeHandle = 0
            nativeUnload(handle)
        }
    }

    private external fun nativeLoad(modelPath: String, threadCount: Int, contextTokens: Int, batchTokens: Int): Long
    private external fun nativeGenerate(
        handle: Long,
        requestId: String,
        systemMessage: String,
        userMessage: String,
        maxTokens: Int,
        maxDurationMillis: Long,
        temperature: Float,
        topK: Int,
        repetitionPenalty: Float,
        seed: Int,
        greedy: Boolean,
    ): Array<String>
    private external fun nativeCancel(handle: Long)
    private external fun nativeUnload(handle: Long)

    companion object {
        private const val TAG = "GmnLlamaBridge"
        private const val RESULT_FIELD_COUNT = 9

        init {
            System.loadLibrary("gmn_llama")
        }
    }
}

private fun String.toNativeStopReason(): NativeStopReason =
    when (this) {
        "eog" -> NativeStopReason.Eog
        "max_tokens" -> NativeStopReason.MaxTokens
        "timeout" -> NativeStopReason.Timeout
        "cancelled" -> NativeStopReason.Cancelled
        "context_overflow" -> NativeStopReason.ContextOverflow
        "unsupported_template" -> NativeStopReason.UnsupportedTemplate
        "prompt_decode_error" -> NativeStopReason.PromptDecodeError
        else -> NativeStopReason.NativeError
    }

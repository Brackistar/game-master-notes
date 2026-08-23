package com.brackistar.gamemasternotes.core.ai

import android.os.Looper
import android.util.Log

class LlamaCppBridge {
    @Volatile
    private var nativeHandle: Long = 0

    fun load(modelPath: String, threadCount: Int, contextTokens: Int) {
        val startedAt = System.currentTimeMillis()
        Log.i(TAG, "JNI load started pathHash=${modelPath.hashCode()} threads=$threadCount contextTokens=$contextTokens")
        nativeHandle = nativeLoad(modelPath, threadCount, contextTokens)
        Log.i(TAG, "JNI load finished handle=$nativeHandle elapsedMs=${System.currentTimeMillis() - startedAt}")
    }

    fun generate(prompt: String, maxTokens: Int, maxDurationMillis: Long): String {
        check(nativeHandle != 0L) { "No llama.cpp model is loaded." }
        check(Looper.myLooper() != Looper.getMainLooper()) {
            "llama.cpp generation must not run on the main thread."
        }
        val startedAt = System.currentTimeMillis()
        Log.i(
            TAG,
            "JNI generate started handle=$nativeHandle promptChars=${prompt.length} maxTokens=$maxTokens maxDurationMs=$maxDurationMillis thread=${Thread.currentThread().name}",
        )
        return nativeGenerate(nativeHandle, prompt, maxTokens, maxDurationMillis).also {
            Log.i(TAG, "JNI generate finished handle=$nativeHandle outputChars=${it.length} elapsedMs=${System.currentTimeMillis() - startedAt}")
        }
    }

    fun cancel() {
        val handle = nativeHandle
        Log.w(TAG, "JNI cancel requested handle=$handle")
        if (handle != 0L) nativeCancel(handle)
    }

    fun unload() {
        if (nativeHandle != 0L) {
            Log.i(TAG, "JNI unload started handle=$nativeHandle")
            val startedAt = System.currentTimeMillis()
            nativeUnload(nativeHandle)
            Log.i(TAG, "JNI unload finished elapsedMs=${System.currentTimeMillis() - startedAt}")
            nativeHandle = 0
        }
    }

    private external fun nativeLoad(modelPath: String, threadCount: Int, contextTokens: Int): Long
    private external fun nativeGenerate(handle: Long, prompt: String, maxTokens: Int, maxDurationMillis: Long): String
    private external fun nativeCancel(handle: Long)
    private external fun nativeUnload(handle: Long)

    companion object {
        private const val TAG = "GmnLlamaBridge"

        init {
            Log.i(TAG, "Loading native library gmn_llama")
            System.loadLibrary("gmn_llama")
            Log.i(TAG, "Loaded native library gmn_llama")
        }
    }
}

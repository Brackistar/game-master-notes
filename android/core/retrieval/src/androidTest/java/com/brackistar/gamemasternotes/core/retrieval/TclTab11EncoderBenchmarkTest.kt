package com.brackistar.gamemasternotes.core.retrieval

import android.os.Build
import android.os.Debug
import android.util.Log
import androidx.test.core.app.ApplicationProvider
import androidx.test.ext.junit.runners.AndroidJUnit4
import kotlinx.coroutines.runBlocking
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class TclTab11EncoderBenchmarkTest {
    @Test
    fun recordsColdAndWarmEncoderMeasurements() = runBlocking {
        assertTrue("Physical acceptance requires Android 15 or newer.", Build.VERSION.SDK_INT >= 35)
        val context = ApplicationProvider.getApplicationContext<android.content.Context>()
        val encoder = MiniLmOnnxQueryEncoder(context)
        val samples = mutableListOf<Long>()
        repeat(6) { iteration ->
            val started = System.nanoTime()
            val vector = encoder.encode(QUERIES[iteration % QUERIES.size])
            samples += (System.nanoTime() - started) / 1_000_000
            assertEquals(384, vector.size)
        }
        val memory = Debug.MemoryInfo().also(Debug::getMemoryInfo)
        Log.i(
            TAG,
            "device=${Build.MANUFACTURER}/${Build.MODEL} sdk=${Build.VERSION.SDK_INT} " +
                "coldMs=${samples.first()} warmMs=${samples.drop(1).joinToString(",")} " +
                "pssKb=${memory.totalPss} nativePssKb=${memory.nativePss}",
        )
        encoder.close()
    }

    private companion object {
        const val TAG = "GmnTabletAcceptance"
        val QUERIES = listOf(
            "Who guards the hidden library?",
            "What opens the observatory at midnight?",
            "How is the keeper connected to the silver key?",
        )
    }
}

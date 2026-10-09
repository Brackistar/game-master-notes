package com.brackistar.gamemasternotes.core.retrieval

import androidx.test.core.app.ApplicationProvider
import androidx.test.ext.junit.runners.AndroidJUnit4
import kotlinx.coroutines.runBlocking
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith
import kotlin.math.sqrt

@RunWith(AndroidJUnit4::class)
class MiniLmOnnxQueryEncoderInstrumentedTest {
    @Test
    fun androidEncoderMatchesPackBuilderGoldenVector() = runBlocking {
        val context = ApplicationProvider.getApplicationContext<android.content.Context>()
        val golden = JSONObject(context.assets.open("minilm_golden.json").bufferedReader().use { it.readText() })
        val expectedJson = golden.getJSONArray("vector")
        val expected = FloatArray(expectedJson.length()) { expectedJson.getDouble(it).toFloat() }
        MiniLmOnnxQueryEncoder(context).use { encoder ->
            val actual = encoder.encode(golden.getString("text"))
            assertEquals(384, actual.size)
            assertTrue("Cosine similarity was ${cosine(expected, actual)}", cosine(expected, actual) >= 0.995)
        }
    }

    private fun cosine(left: FloatArray, right: FloatArray): Double {
        var dot = 0.0
        var leftNorm = 0.0
        var rightNorm = 0.0
        left.indices.forEach { index ->
            dot += left[index] * right[index]
            leftNorm += left[index] * left[index]
            rightNorm += right[index] * right[index]
        }
        return dot / (sqrt(leftNorm) * sqrt(rightNorm))
    }
}

package com.brackistar.gamemasternotes.core.importpacks

import org.junit.Assert.assertEquals
import org.junit.Test
import java.util.zip.ZipInputStream

class NpyFloatMatrixTest {
    @Test fun validatesRealPackBuilderFixture() {
        val stream = checkNotNull(javaClass.classLoader?.getResourceAsStream("v1/synthetic.gmnpack"))
        val bytes = ZipInputStream(stream).use { zip ->
            generateSequence { zip.nextEntry }.first { it.name == "embeddings.npy" }
            zip.readBytes()
        }
        val info = NpyFloatMatrix.validate(bytes, expectedRows = 2, expectedColumns = 384)
        assertEquals(NpyMatrixInfo(2, 384), info)
    }

    @Test(expected = PackImportException::class)
    fun rejectsShapeMismatch() {
        val stream = checkNotNull(javaClass.classLoader?.getResourceAsStream("v1/synthetic.gmnpack"))
        val bytes = ZipInputStream(stream).use { zip ->
            generateSequence { zip.nextEntry }.first { it.name == "embeddings.npy" }
            zip.readBytes()
        }
        NpyFloatMatrix.validate(bytes, expectedRows = 3, expectedColumns = 384)
    }
}

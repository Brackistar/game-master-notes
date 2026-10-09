package com.brackistar.gamemasternotes.core.importpacks

import java.nio.ByteBuffer
import java.nio.ByteOrder
import kotlin.math.abs
import kotlin.math.sqrt

data class NpyMatrixInfo(val rows: Int, val columns: Int)

object NpyFloatMatrix {
    fun validate(bytes: ByteArray, expectedRows: Int, expectedColumns: Int): NpyMatrixInfo {
        if (bytes.size < 12 || !bytes.copyOfRange(0, 6).contentEquals(byteArrayOf(0x93.toByte(), 0x4e, 0x55, 0x4d, 0x50, 0x59))) {
            throw PackImportException("embeddings.npy has an invalid NPY header.")
        }
        val major = bytes[6].toInt()
        val headerLengthBytes = if (major == 1) 2 else 4
        val headerStart = 8 + headerLengthBytes
        if (bytes.size < headerStart) throw PackImportException("embeddings.npy header is truncated.")
        val headerLength = ByteBuffer.wrap(bytes, 8, headerLengthBytes).order(ByteOrder.LITTLE_ENDIAN).let {
            if (headerLengthBytes == 2) it.short.toInt() and 0xffff else it.int
        }
        if (headerLength <= 0 || headerStart + headerLength > bytes.size) throw PackImportException("embeddings.npy header is truncated.")
        val header = bytes.decodeToString(headerStart, headerStart + headerLength)
        if (!(header.contains("'<f4'") || header.contains("'|f4'")) || header.contains("'fortran_order': True")) {
            throw PackImportException("embeddings.npy must be a C-order float32 matrix.")
        }
        val match = Regex("'shape':\\s*\\((\\d+)\\s*,\\s*(\\d+)\\s*[,)]").find(header)
            ?: throw PackImportException("embeddings.npy must have a two-dimensional shape.")
        val rows = match.groupValues[1].toInt()
        val columns = match.groupValues[2].toInt()
        if (rows != expectedRows || columns != expectedColumns) {
            throw PackImportException("Embedding shape ($rows, $columns) does not match expected ($expectedRows, $expectedColumns).")
        }
        val dataStart = headerStart + headerLength
        val expectedBytes = rows.toLong() * columns * Float.SIZE_BYTES
        if ((bytes.size - dataStart).toLong() != expectedBytes) throw PackImportException("embeddings.npy data length does not match its shape.")
        val values = ByteBuffer.wrap(bytes, dataStart, expectedBytes.toInt()).order(ByteOrder.LITTLE_ENDIAN)
        repeat(rows) {
            var squaredNorm = 0.0
            repeat(columns) {
                val value = values.float
                if (!value.isFinite()) throw PackImportException("embeddings.npy contains a non-finite value.")
                squaredNorm += value * value
            }
            if (abs(sqrt(squaredNorm) - 1.0) > NORMALIZATION_TOLERANCE) {
                throw PackImportException("embeddings.npy rows must be normalized.")
            }
        }
        return NpyMatrixInfo(rows, columns)
    }

    private const val NORMALIZATION_TOLERANCE = 0.01
}

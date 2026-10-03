package com.brackistar.gamemasternotes.core.retrieval

import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.json.JSONObject
import java.io.File
import java.io.RandomAccessFile
import java.nio.ByteBuffer
import java.nio.ByteOrder
import java.nio.channels.FileChannel
import kotlin.math.sqrt

class NpyVectorSidecarStore(
    private val directory: File,
    override val modelId: String,
    private val modelRevision: String,
    override val dimensions: Int,
    private val dispatcher: CoroutineDispatcher = Dispatchers.Default,
) : VectorRetrievalStore {
    override suspend fun search(queryVector: FloatArray, limit: Int): List<RankedCandidate> =
        withContext(dispatcher) {
            require(queryVector.size == dimensions && queryVector.all(Float::isFinite))
            if (!directory.isDirectory) return@withContext emptyList()
            directory.listFiles { file -> file.extension == "json" }.orEmpty()
                .asSequence()
                .flatMap { metadataFile -> scan(metadataFile, queryVector).asSequence() }
                .sortedWith(compareByDescending<RankedCandidate> { it.score }.thenBy { it.sourceId })
                .take(limit)
                .toList()
        }

    private fun scan(metadataFile: File, query: FloatArray): List<RankedCandidate> {
        val metadata = runCatching { JSONObject(metadataFile.readText()) }.getOrNull() ?: return emptyList()
        if (
            metadata.optString("model_id") != modelId ||
            metadata.optString("model_revision") != modelRevision ||
            metadata.optInt("dimensions") != dimensions
        ) return emptyList()
        val chunkIdsJson = metadata.optJSONArray("chunk_ids") ?: return emptyList()
        val chunkIds = (0 until chunkIdsJson.length()).map(chunkIdsJson::getString)
        val vectorFile = File(metadataFile.parentFile, "${metadataFile.nameWithoutExtension}.npy")
        if (!vectorFile.isFile) return emptyList()
        return RandomAccessFile(vectorFile, "r").use { file ->
            val header = readHeader(file.channel)
            if (header.rows != chunkIds.size || header.columns != dimensions) return emptyList()
            val mapped = file.channel.map(FileChannel.MapMode.READ_ONLY, header.dataOffset, header.rows.toLong() * dimensions * 4)
                .order(ByteOrder.LITTLE_ENDIAN)
            buildList(header.rows) {
                repeat(header.rows) { row ->
                    var dot = 0.0
                    var rowNorm = 0.0
                    repeat(dimensions) { column ->
                        val value = mapped.float
                        dot += query[column] * value
                        rowNorm += value * value
                    }
                    val score = if (rowNorm > 0) dot / sqrt(rowNorm) else 0.0
                    add(RankedCandidate(chunkIds[row], score))
                }
            }
        }
    }

    private fun readHeader(channel: FileChannel): Header {
        val prefix = ByteBuffer.allocate(12).order(ByteOrder.LITTLE_ENDIAN)
        channel.read(prefix, 0)
        prefix.flip()
        val magic = ByteArray(6).also(prefix::get)
        require(magic.contentEquals(byteArrayOf(0x93.toByte(), 0x4e, 0x55, 0x4d, 0x50, 0x59)))
        val major = prefix.get().toInt()
        prefix.get()
        val headerLengthBytes = if (major == 1) 2 else 4
        val length = if (headerLengthBytes == 2) prefix.short.toInt() and 0xffff else prefix.int
        val dataOffset = 8L + headerLengthBytes + length
        val headerBuffer = ByteBuffer.allocate(length)
        channel.read(headerBuffer, (8 + headerLengthBytes).toLong())
        val text = headerBuffer.array().decodeToString()
        val shape = Regex("'shape':\\s*\\((\\d+)\\s*,\\s*(\\d+)\\s*[,)]").find(text) ?: error("Invalid NPY shape")
        return Header(shape.groupValues[1].toInt(), shape.groupValues[2].toInt(), dataOffset)
    }

    private data class Header(val rows: Int, val columns: Int, val dataOffset: Long)
}

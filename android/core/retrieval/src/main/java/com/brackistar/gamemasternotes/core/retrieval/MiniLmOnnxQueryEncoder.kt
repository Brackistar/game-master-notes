package com.brackistar.gamemasternotes.core.retrieval

import ai.onnxruntime.OnnxTensor
import ai.onnxruntime.OrtEnvironment
import ai.onnxruntime.OrtSession
import android.content.Context
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext
import org.json.JSONObject
import java.io.File
import java.security.MessageDigest
import java.text.Normalizer
import kotlin.math.sqrt

class MiniLmOnnxQueryEncoder(
    context: Context,
    private val dispatcher: CoroutineDispatcher = Dispatchers.Default,
) : QueryEmbeddingEncoder, AutoCloseable {
    private val appContext = context.applicationContext
    private val installDirectory = appContext.filesDir.resolve("encoders/minilm-l6-v2")
    private val mutex = Mutex()
    private var resources: Resources? = null

    override val modelId: String = MODEL_ID
    override val dimensions: Int = DIMENSIONS

    override suspend fun encode(text: String): FloatArray = withContext(dispatcher) {
        require(text.isNotBlank()) { "Embedding query must not be blank." }
        val loaded = mutex.withLock { resources ?: loadResources().also { resources = it } }
        val encoded = loaded.tokenizer.encode(text, MAX_WORD_PIECES)
        val shape = longArrayOf(1, MAX_WORD_PIECES.toLong())
        OnnxTensor.createTensor(loaded.environment, arrayOf(encoded.inputIds)).use { inputIds ->
            OnnxTensor.createTensor(loaded.environment, arrayOf(encoded.attentionMask)).use { attentionMask ->
                OnnxTensor.createTensor(loaded.environment, arrayOf(encoded.tokenTypeIds)).use { tokenTypes ->
                    val inputs = buildMap {
                        if ("input_ids" in loaded.session.inputNames) put("input_ids", inputIds)
                        if ("attention_mask" in loaded.session.inputNames) put("attention_mask", attentionMask)
                        if ("token_type_ids" in loaded.session.inputNames) put("token_type_ids", tokenTypes)
                    }
                    loaded.session.run(inputs).use { result ->
                        normalize(readSentenceEmbedding(result, encoded.attentionMask))
                    }
                }
            }
        }
    }

    override fun close() {
        resources?.session?.close()
        resources = null
    }

    private fun loadResources(): Resources {
        installDirectory.mkdirs()
        val contract = JSONObject(appContext.assets.open(CONTRACT_ASSET).bufferedReader().use { it.readText() })
        require(contract.getString("model_id") == MODEL_ID)
        require(contract.getInt("dimensions") == DIMENSIONS)
        require(contract.getInt("max_word_pieces") == MAX_WORD_PIECES)
        val modelFile = installAsset(MODEL_ASSET, "model.onnx", contract.getString("model_sha256"))
        val vocabFile = installAsset(VOCAB_ASSET, "vocab.txt", contract.getString("vocab_sha256"))
        val environment = OrtEnvironment.getEnvironment("gmn-minilm")
        val options = OrtSession.SessionOptions().apply {
            setIntraOpNumThreads(2)
            setInterOpNumThreads(1)
            setOptimizationLevel(OrtSession.SessionOptions.OptLevel.ALL_OPT)
        }
        return Resources(
            environment = environment,
            session = environment.createSession(modelFile.absolutePath, options),
            tokenizer = WordPieceTokenizer(vocabFile.readLines()),
        )
    }

    private fun installAsset(assetPath: String, fileName: String, expectedSha256: String): File {
        val destination = installDirectory.resolve(fileName)
        if (destination.isFile && destination.sha256() == expectedSha256) return destination
        val staged = installDirectory.resolve(".$fileName.tmp")
        appContext.assets.open(assetPath).use { input ->
            staged.outputStream().buffered().use { output -> input.copyTo(output, 32 * 1024) }
        }
        require(staged.sha256() == expectedSha256) { "$fileName failed checksum verification." }
        if (!staged.renameTo(destination)) {
            staged.copyTo(destination, overwrite = true)
            staged.delete()
        }
        return destination
    }

    @Suppress("UNCHECKED_CAST")
    private fun readSentenceEmbedding(result: OrtSession.Result, attentionMask: LongArray): FloatArray {
        val sentence = result.get("sentence_embedding").orElse(null)?.value
        if (sentence is Array<*> && sentence.firstOrNull() is FloatArray) {
            return (sentence as Array<FloatArray>)[0]
        }
        val tokens = result.get("token_embeddings").orElse(null)?.value
            ?: result.firstOrNull()?.value
            ?: error("MiniLM output did not contain sentence or token embeddings.")
        val tokenRows = ((tokens as Array<*>)[0] as Array<FloatArray>)
        val pooled = FloatArray(DIMENSIONS)
        var active = 0
        tokenRows.forEachIndexed { index, row ->
            if (index < attentionMask.size && attentionMask[index] == 1L) {
                for (dimension in pooled.indices) pooled[dimension] += row[dimension]
                active++
            }
        }
        require(active > 0)
        for (dimension in pooled.indices) pooled[dimension] /= active
        return pooled
    }

    private fun normalize(values: FloatArray): FloatArray {
        require(values.size == DIMENSIONS && values.all(Float::isFinite))
        val norm = sqrt(values.sumOf { it.toDouble() * it })
        require(norm > 0.0)
        return FloatArray(values.size) { index -> (values[index] / norm).toFloat() }
    }

    private data class Resources(
        val environment: OrtEnvironment,
        val session: OrtSession,
        val tokenizer: WordPieceTokenizer,
    )

    companion object {
        const val MODEL_ID = "sentence-transformers/all-MiniLM-L6-v2"
        const val MODEL_REVISION = "1110a243fdf4706b3f48f1d95db1a4f5529b4d41"
        const val DIMENSIONS = 384
        const val MAX_WORD_PIECES = 256
        private const val CONTRACT_ASSET = "minilm/contract.json"
        private const val MODEL_ASSET = "minilm/model.onnx"
        private const val VOCAB_ASSET = "minilm/vocab.txt"
    }
}

internal class WordPieceTokenizer(vocabulary: List<String>) {
    private val tokenIds = vocabulary.withIndex().associate { it.value to it.index.toLong() }
    private val unknownId = tokenIds.getValue("[UNK]")
    private val clsId = tokenIds.getValue("[CLS]")
    private val sepId = tokenIds.getValue("[SEP]")
    private val padId = tokenIds.getValue("[PAD]")

    fun encode(text: String, maxTokens: Int): EncodedTokens {
        require(maxTokens >= 2)
        val pieces = basicTokens(text).flatMap(::wordPieces).take(maxTokens - 2)
        val ids = LongArray(maxTokens) { padId }
        ids[0] = clsId
        pieces.forEachIndexed { index, piece -> ids[index + 1] = tokenIds.getValue(piece) }
        ids[pieces.size + 1] = sepId
        val active = pieces.size + 2
        return EncodedTokens(
            inputIds = ids,
            attentionMask = LongArray(maxTokens) { if (it < active) 1L else 0L },
            tokenTypeIds = LongArray(maxTokens),
        )
    }

    private fun basicTokens(text: String): List<String> {
        val normalized = Normalizer.normalize(text.lowercase(), Normalizer.Form.NFD)
            .filterNot { Character.getType(it) == Character.NON_SPACING_MARK.toInt() }
        return buildList {
            val current = StringBuilder()
            fun flush() { if (current.isNotEmpty()) { add(current.toString()); current.clear() } }
            normalized.forEach { character ->
                when {
                    character.isWhitespace() || Character.isISOControl(character) -> flush()
                    character.isLetterOrDigit() -> current.append(character)
                    else -> { flush(); add(character.toString()) }
                }
            }
            flush()
        }
    }

    private fun wordPieces(token: String): List<String> {
        if (token.length > 100) return listOf("[UNK]")
        val pieces = mutableListOf<String>()
        var start = 0
        while (start < token.length) {
            var end = token.length
            var match: String? = null
            while (start < end) {
                val candidate = (if (start == 0) "" else "##") + token.substring(start, end)
                if (candidate in tokenIds) { match = candidate; break }
                end--
            }
            if (match == null) return listOf("[UNK]")
            pieces += match
            start = end
        }
        return pieces
    }
}

internal data class EncodedTokens(
    val inputIds: LongArray,
    val attentionMask: LongArray,
    val tokenTypeIds: LongArray,
)

private fun File.sha256(): String = inputStream().buffered().use { input ->
    val digest = MessageDigest.getInstance("SHA-256")
    val buffer = ByteArray(32 * 1024)
    while (true) {
        val count = input.read(buffer)
        if (count < 0) break
        digest.update(buffer, 0, count)
    }
    digest.digest().joinToString("") { "%02x".format(it) }
}

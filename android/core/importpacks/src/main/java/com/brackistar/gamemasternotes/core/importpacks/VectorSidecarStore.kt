package com.brackistar.gamemasternotes.core.importpacks

import org.json.JSONArray
import org.json.JSONObject
import java.io.File
import java.nio.file.Files
import java.nio.file.StandardCopyOption

class VectorSidecarStore(private val directory: File) {
    fun isCurrent(packId: String, fingerprint: String, modelId: String, modelRevision: String): Boolean {
        val metadata = metadataFile(packId)
        val vectors = vectorFile(packId)
        if (!metadata.isFile || !vectors.isFile) return false
        return runCatching {
            val json = JSONObject(metadata.readText())
            json.getString("fingerprint") == fingerprint &&
                json.getString("model_id") == modelId &&
                json.optString("model_revision") == modelRevision
        }.getOrDefault(false)
    }

    fun installAtomic(
        packId: String,
        fingerprint: String,
        modelId: String,
        modelRevision: String,
        dimensions: Int,
        chunkIds: List<String>,
        npyBytes: ByteArray,
    ) {
        directory.mkdirs()
        val safeId = packId.safeFileName()
        val stagedVectors = File(directory, ".$safeId-vectors.tmp")
        val stagedMetadata = File(directory, ".$safeId-metadata.tmp")
        stagedVectors.outputStream().buffered().use { it.write(npyBytes) }
        stagedMetadata.writeText(
            JSONObject()
                .put("format_version", 1)
                .put("pack_id", packId)
                .put("fingerprint", fingerprint)
                .put("model_id", modelId)
                .put("model_revision", modelRevision)
                .put("dimensions", dimensions)
                .put("chunk_ids", JSONArray(chunkIds))
                .toString(),
        )
        replace(stagedVectors, vectorFile(packId))
        replace(stagedMetadata, metadataFile(packId))
    }

    fun pruneToPackIds(packIds: Set<String>) {
        if (!directory.isDirectory) return
        val safeIds = packIds.mapTo(hashSetOf()) { it.safeFileName() }
        directory.listFiles().orEmpty().forEach { file ->
            val base = file.name.substringBeforeLast('.')
            if (!file.name.startsWith('.') && base !in safeIds) file.delete()
        }
    }

    private fun vectorFile(packId: String) = File(directory, "${packId.safeFileName()}.npy")
    private fun metadataFile(packId: String) = File(directory, "${packId.safeFileName()}.json")

    private fun replace(source: File, destination: File) {
        runCatching {
            Files.move(
                source.toPath(),
                destination.toPath(),
                StandardCopyOption.REPLACE_EXISTING,
                StandardCopyOption.ATOMIC_MOVE,
            )
        }.getOrElse {
            Files.move(source.toPath(), destination.toPath(), StandardCopyOption.REPLACE_EXISTING)
        }
    }
}

private fun String.safeFileName(): String =
    replace(Regex("[^A-Za-z0-9._-]"), "_").take(120)

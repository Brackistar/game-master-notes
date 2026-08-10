package com.brackistar.gamemasternotes.core.importpacks

interface PackImporter {
    suspend fun inspectPack(path: String): PackInspection
}

data class PackInspection(
    val packId: String,
    val title: String,
    val schemaVersion: Int,
    val chunkCount: Int,
    val embeddingCount: Int,
)

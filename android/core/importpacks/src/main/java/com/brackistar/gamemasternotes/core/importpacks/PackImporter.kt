package com.brackistar.gamemasternotes.core.importpacks

import android.content.ContentResolver
import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.documentfile.provider.DocumentFile
import com.brackistar.gamemasternotes.core.data.ImportedPack
import com.brackistar.gamemasternotes.core.data.SourceChunkEntity
import com.brackistar.gamemasternotes.core.data.SourceChunkFtsEntity
import com.brackistar.gamemasternotes.core.data.SourceDocumentEntity
import com.brackistar.gamemasternotes.core.data.SourcebookPackEntity
import com.brackistar.gamemasternotes.core.data.SourcebookRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.withContext
import org.json.JSONArray
import org.json.JSONObject
import java.io.BufferedReader
import java.io.ByteArrayOutputStream
import java.io.InputStreamReader
import java.util.zip.ZipEntry
import java.util.zip.ZipInputStream

interface PackImporter {
    suspend fun inspectPack(uri: Uri): PackInspection
    suspend fun importFolder(treeUri: Uri): PackImportSummary
}

data class PackInspection(
    val packId: String,
    val title: String,
    val schemaVersion: String,
    val chunkCount: Int,
    val embeddingCount: Int,
)

data class PackImportSummary(
    val scannedCount: Int,
    val importedCount: Int,
    val skippedUnchangedCount: Int,
    val removedCount: Int,
    val errors: List<String>,
)

class ContentResolverPackImporter(
    private val context: Context,
    private val repository: SourcebookRepository,
    private val vectorSidecarStore: VectorSidecarStore,
) : PackImporter {
    private val resolver: ContentResolver = context.contentResolver

    override suspend fun inspectPack(uri: Uri): PackInspection = withContext(Dispatchers.IO) {
        val members = readPackMembers(uri, readChunks = false)
        val manifest = members.manifest
        PackInspection(
            packId = manifest.requiredString("pack_id"),
            title = manifest.requiredString("title"),
            schemaVersion = manifest.requiredSupportedSchemaVersion(),
            chunkCount = manifest.requiredInt("chunk_count"),
            embeddingCount = manifest.optInt("chunk_count"),
        )
    }

    override suspend fun importFolder(treeUri: Uri): PackImportSummary = withContext(Dispatchers.IO) {
        val folder = DocumentFile.fromTreeUri(context, treeUri)
            ?: return@withContext PackImportSummary(0, 0, 0, 0, listOf("Selected folder is not available."))
        val packFiles = folder.findPackFiles()

        var imported = 0
        var skipped = 0
        val seenPackIds = mutableListOf<String>()
        val errors = mutableListOf<String>()

        for (packFile in packFiles) {
            try {
                val parsed = parsePack(treeUri, packFile)
                seenPackIds += parsed.importedPack.pack.packId
                val pack = parsed.importedPack.pack
                val vectorsCurrent = vectorSidecarStore.isCurrent(
                    pack.packId,
                    pack.archiveFingerprint,
                    parsed.embeddingModelId,
                    parsed.embeddingModelRevision,
                )
                if (repository.existingFingerprint(pack.packId) == pack.archiveFingerprint && vectorsCurrent) {
                    skipped += 1
                } else {
                    repository.replaceImportedPack(parsed.importedPack)
                    vectorSidecarStore.installAtomic(
                        packId = pack.packId,
                        fingerprint = pack.archiveFingerprint,
                        modelId = parsed.embeddingModelId,
                        modelRevision = parsed.embeddingModelRevision,
                        dimensions = parsed.embeddingDimensions,
                        chunkIds = parsed.importedPack.chunks
                            .sortedBy { it.embeddingRowIndex }
                            .map { it.chunkId },
                        npyBytes = parsed.embeddings,
                    )
                    imported += 1
                }
            } catch (error: Exception) {
                if (error is CancellationException) throw error
                errors += "${packFile.name ?: "Unknown pack"}: ${error.message}"
            }
        }

        val removed = if (shouldPruneAfterScan(errors)) {
            val beforePrune = repository.packIds()
            repository.pruneToAvailablePacks(seenPackIds)
            vectorSidecarStore.pruneToPackIds(seenPackIds.toSet())
            (beforePrune - seenPackIds.toSet()).size
        } else {
            0
        }

        PackImportSummary(
            scannedCount = packFiles.size,
            importedCount = imported,
            skippedUnchangedCount = skipped,
            removedCount = removed,
            errors = errors,
        )
    }

    private fun DocumentFile.findPackFiles(): List<DocumentFile> {
        val packFiles = mutableListOf<DocumentFile>()

        fun visit(folder: DocumentFile, depth: Int) {
            if (depth > MAX_FOLDER_SCAN_DEPTH) return

            folder.listFiles().forEach { child ->
                val childName = child.name.orEmpty()
                when {
                    child.isDirectory -> visit(child, depth + 1)
                    childName.endsWith(PACK_FILE_EXTENSION, ignoreCase = true) -> packFiles += child
                }
            }
        }

        visit(this, depth = 0)
        return packFiles
    }

    private fun parsePack(folderUri: Uri, packFile: DocumentFile): ParsedPack {
        val uri = packFile.uri
        val members = readPackMembers(uri, readChunks = true)
        val manifest = members.manifest
        manifest.requiredSupportedSchemaVersion()
        val packId = manifest.requiredString("pack_id")
        val title = manifest.requiredString("title")
        val system = manifest.requiredString("system")
        val importedAt = System.currentTimeMillis()
        val fingerprint = listOf(
            uri.toString(),
            packFile.name.orEmpty(),
            packFile.length().toString(),
            packFile.lastModified().toString(),
        ).joinToString("|")

        val documents = members.documents
            .requiredArray("documents")
            .mapObjects { document ->
                SourceDocumentEntity(
                    documentId = document.requiredString("document_id"),
                    packId = packId,
                    sourceFilename = document.requiredString("source_filename"),
                    sourceChecksum = document.requiredString("source_checksum"),
                    pageCount = document.requiredInt("page_count"),
                )
            }

        val chunks = members.chunks.map { chunk ->
            SourceChunkEntity(
                chunkId = chunk.requiredString("chunk_id"),
                packId = packId,
                documentId = chunk.requiredString("document_id"),
                pageStart = chunk.requiredInt("page_start"),
                pageEnd = chunk.requiredInt("page_end"),
                citationLabel = chunk.requiredString("citation_label"),
                text = chunk.requiredString("text"),
                charCount = chunk.requiredInt("char_count"),
                embeddingRowIndex = chunk.requiredInt("embedding_row_index"),
            )
        }
        if (chunks.size != manifest.requiredInt("chunk_count")) {
            throw PackImportException("Manifest chunk count does not match chunks.jsonl.")
        }
        val embeddingDimensions = manifest.optIntOrNull("embedding_dimensions")
            ?: throw PackImportException("Missing required field embedding_dimensions.")
        val embeddingModelId = manifest.optStringOrNull("embedding_model_id")
            ?: throw PackImportException("Missing required field embedding_model_id.")
        val embeddingModelRevision = manifest.optStringOrNull("embedding_model_revision") ?: "unversioned"
        NpyFloatMatrix.validate(members.embeddings, chunks.size, embeddingDimensions)

        return ParsedPack(
            importedPack = ImportedPack(
                pack = SourcebookPackEntity(
                packId = packId,
                title = title,
                system = system,
                edition = manifest.requiredString("edition"),
                language = manifest.requiredString("language"),
                schemaVersion = SUPPORTED_PACK_SCHEMA_STORAGE_VERSION,
                generatorVersion = manifest.requiredString("generator_version"),
                embeddingModelId = embeddingModelId,
                embeddingDimensions = embeddingDimensions,
                chunkCount = chunks.size,
                sourceFolderUri = folderUri.toString(),
                sourceDisplayName = packFile.name ?: title,
                archiveDocumentUri = uri.toString(),
                archiveFingerprint = fingerprint,
                importedAtEpochMillis = importedAt,
            ),
                documents = documents,
                chunks = chunks,
                ftsRows = chunks.map { chunk ->
                    SourceChunkFtsEntity(
                        chunkId = chunk.chunkId,
                        packId = packId,
                        title = title,
                        system = system,
                        text = chunk.text,
                    )
                },
            ),
            embeddings = members.embeddings,
            embeddingModelId = embeddingModelId,
            embeddingModelRevision = embeddingModelRevision,
            embeddingDimensions = embeddingDimensions,
        )
    }

    private fun readPackMembers(uri: Uri, readChunks: Boolean): PackMembers {
        val entries = mutableMapOf<String, ByteArray>()
        val seenMembers = mutableSetOf<String>()
        resolver.openInputStream(uri)?.use { input ->
            ZipInputStream(input).use { zip ->
                var entry = zip.nextEntry
                while (entry != null) {
                    if (!entry.isDirectory) {
                        val name = entry.name
                        if (name in requiredMembers) {
                            seenMembers += name
                        }
                        if (name in readableMembers || (readChunks && name in setOf("chunks.jsonl", "embeddings.npy"))) {
                            entries[name] = zip.readEntryBytes(entry, name)
                        }
                    }
                    entry = zip.nextEntry
                }
            }
        } ?: throw PackImportException("Could not open pack archive.")

        for (member in requiredMembers) {
            if (member !in seenMembers) throw PackImportException("Missing required archive member $member.")
        }

        return PackMembers(
            manifest = JSONObject(entries.requiredBytes("manifest.json").decodeToString()),
            documents = JSONObject(entries.requiredBytes("documents.json").decodeToString()),
            chunks = if (readChunks) {
                BufferedReader(InputStreamReader(entries.requiredBytes("chunks.jsonl").inputStream()))
                    .lineSequence()
                    .filter { it.isNotBlank() }
                    .map { JSONObject(it) }
                    .toList()
            } else {
                emptyList()
            },
            embeddings = if (readChunks) entries.requiredBytes("embeddings.npy") else byteArrayOf(),
        )
    }
}

class PackFolderStore(context: Context) {
    private val preferences = context.getSharedPreferences("pack-folders", Context.MODE_PRIVATE)

    fun selectedFolderUri(): Uri? = preferences.getString(KEY_FOLDER_URI, null)?.let(Uri::parse)

    fun saveSelectedFolder(uri: Uri) {
        preferences.edit().putString(KEY_FOLDER_URI, uri.toString()).apply()
    }

    companion object {
        private const val KEY_FOLDER_URI = "selected_folder_uri"

        fun persistReadPermission(context: Context, uri: Uri) {
            context.contentResolver.takePersistableUriPermission(
                uri,
                Intent.FLAG_GRANT_READ_URI_PERMISSION,
            )
        }
    }
}

private data class PackMembers(
    val manifest: JSONObject,
    val documents: JSONObject,
    val chunks: List<JSONObject>,
    val embeddings: ByteArray,
)

private data class ParsedPack(
    val importedPack: ImportedPack,
    val embeddings: ByteArray,
    val embeddingModelId: String,
    val embeddingModelRevision: String,
    val embeddingDimensions: Int,
)

class PackImportException(message: String) : Exception(message)

private val requiredMembers = setOf(
    "manifest.json",
    "documents.json",
    "chunks.jsonl",
    "embeddings.npy",
    "extraction-report.json",
)

private val readableMembers = requiredMembers - setOf("embeddings.npy", "chunks.jsonl")

private const val PACK_FILE_EXTENSION = ".gmnpack"
private const val MAX_FOLDER_SCAN_DEPTH = 3
private const val MAX_MANIFEST_BYTES = 256 * 1024
private const val MAX_DOCUMENTS_BYTES = 2 * 1024 * 1024
private const val MAX_CHUNKS_BYTES = 24 * 1024 * 1024
private const val MAX_REPORT_BYTES = 4 * 1024 * 1024
private const val MAX_EMBEDDINGS_BYTES = 64 * 1024 * 1024
private const val READ_BUFFER_BYTES = 8 * 1024
private const val SUPPORTED_PACK_SCHEMA_VERSION = "1.0"
private const val SUPPORTED_PACK_SCHEMA_STORAGE_VERSION = 1

private fun ZipInputStream.readEntryBytes(entry: ZipEntry, name: String): ByteArray {
    val maxBytes = maxBytesForMember(name)
    if (entry.size > maxBytes) {
        throw PackImportException("Archive member $name is too large.")
    }

    val buffer = ByteArray(READ_BUFFER_BYTES)
    val output = ByteArrayOutputStream()
    var totalBytes = 0L
    while (true) {
        val read = read(buffer)
        if (read < 0) break
        totalBytes += read
        if (totalBytes > maxBytes) {
            throw PackImportException("Archive member $name is too large.")
        }
        output.write(buffer, 0, read)
    }
    return output.toByteArray()
}

private fun maxBytesForMember(name: String): Int =
    when (name) {
        "manifest.json" -> MAX_MANIFEST_BYTES
        "documents.json" -> MAX_DOCUMENTS_BYTES
        "chunks.jsonl" -> MAX_CHUNKS_BYTES
        "extraction-report.json" -> MAX_REPORT_BYTES
        "embeddings.npy" -> MAX_EMBEDDINGS_BYTES
        else -> MAX_CHUNKS_BYTES
    }

private fun Map<String, ByteArray>.requiredBytes(name: String): ByteArray =
    this[name] ?: throw PackImportException("Missing required archive member $name.")

private fun JSONObject.requiredString(name: String): String =
    optString(name).takeIf { it.isNotBlank() }
        ?: throw PackImportException("Missing required field $name.")

private fun JSONObject.requiredSupportedSchemaVersion(): String =
    validatePackSchemaVersion(if (has("schema_version") && !isNull("schema_version")) get("schema_version") else null)

internal fun validatePackSchemaVersion(value: Any?): String {
    if (value !is String || value != SUPPORTED_PACK_SCHEMA_VERSION) {
        throw PackImportException(
            "Unsupported schema_version ${value ?: "missing"}; expected \"$SUPPORTED_PACK_SCHEMA_VERSION\".",
        )
    }
    return value
}

internal fun shouldPruneAfterScan(errors: List<String>): Boolean = errors.isEmpty()

private fun JSONObject.requiredInt(name: String): Int {
    if (!has(name)) throw PackImportException("Missing required field $name.")
    return optInt(name)
}

private fun JSONObject.optStringOrNull(name: String): String? =
    if (has(name) && !isNull(name)) optString(name) else null

private fun JSONObject.optIntOrNull(name: String): Int? =
    if (has(name) && !isNull(name)) optInt(name) else null

private fun JSONObject.requiredArray(name: String): JSONArray =
    optJSONArray(name) ?: throw PackImportException("Missing required field $name.")

private fun <T> JSONArray.mapObjects(transform: (JSONObject) -> T): List<T> =
    (0 until length()).map { index -> transform(getJSONObject(index)) }

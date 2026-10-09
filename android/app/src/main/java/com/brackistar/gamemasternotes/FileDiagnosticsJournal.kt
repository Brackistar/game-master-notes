package com.brackistar.gamemasternotes

import com.brackistar.gamemasternotes.core.domain.diagnostics.DiagnosticEvent
import com.brackistar.gamemasternotes.core.domain.diagnostics.DiagnosticsJournal
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext
import java.io.ByteArrayOutputStream
import java.io.File
import java.nio.file.AtomicMoveNotSupportedException
import java.nio.file.Files
import java.nio.file.StandardCopyOption
import java.util.zip.ZipEntry
import java.util.zip.ZipOutputStream

class FileDiagnosticsJournal(
    private val directory: File,
    private val buildMetadata: Map<String, String>,
    private val clock: () -> Long = System::currentTimeMillis,
) : DiagnosticsJournal {
    private val mutex = Mutex()

    override suspend fun record(event: DiagnosticEvent) = withContext(Dispatchers.IO) {
        mutex.withLock {
            directory.mkdirs()
            val encoded = encodeEvent(event)
            val line = if (encoded.toByteArray().size <= MAX_FILE_BYTES) {
                encoded
            } else {
                encodeEvent(
                    event.copy(
                        outcome = "event-too-large",
                        fields = mapOf("originalFieldCount" to event.fields.size.toString()),
                    ),
                )
            }
            rotateIfNeeded(line.toByteArray().size + 1)
            activeFile().appendText(line + "\n", Charsets.UTF_8)
            enforceRequestLimit()
        }
    }

    override suspend fun exportArchive(): ByteArray = withContext(Dispatchers.IO) {
        mutex.withLock {
            directory.mkdirs()
            ByteArrayOutputStream().use { output ->
                ZipOutputStream(output).use { zip ->
                    journalFilesOldestFirst().filter(File::isFile).forEach { file ->
                        zip.putNextEntry(ZipEntry(file.name))
                        file.inputStream().use { it.copyTo(zip, bufferSize = COPY_BUFFER_BYTES) }
                        zip.closeEntry()
                    }
                    zip.putNextEntry(ZipEntry("metadata.json"))
                    zip.write(encodeMetadata().toByteArray(Charsets.UTF_8))
                    zip.closeEntry()
                }
                output.toByteArray()
            }
        }
    }

    override suspend fun clear() = withContext(Dispatchers.IO) {
        mutex.withLock {
            journalFilesOldestFirst().forEach { file ->
                if (file.exists() && !file.delete()) {
                    error("Could not clear diagnostics file ${file.name}.")
                }
            }
        }
    }

    private fun rotateIfNeeded(incomingBytes: Int) {
        val active = activeFile()
        if (!active.exists() || active.length() + incomingBytes <= MAX_FILE_BYTES) return

        journalFile(MAX_FILES - 1).delete()
        for (index in MAX_FILES - 2 downTo 0) {
            val source = journalFile(index)
            if (source.exists()) moveReplacing(source, journalFile(index + 1))
        }
    }

    private fun enforceRequestLimit() {
        val files = journalFilesOldestFirst().filter(File::isFile)
        val requestIds = linkedSetOf<String>()
        files.forEach { file ->
            file.useLines { lines ->
                lines.forEach { line -> extractRequestId(line)?.let(requestIds::add) }
            }
        }
        val removeIds = requestIds.take(maxOf(0, requestIds.size - MAX_REQUESTS)).toSet()
        if (removeIds.isEmpty()) return

        files.forEach { file ->
            val retained = file.readLines(Charsets.UTF_8)
                .filterNot { line -> extractRequestId(line) in removeIds }
            replaceText(file, retained.joinToString(separator = "\n", postfix = if (retained.isEmpty()) "" else "\n"))
        }
    }

    private fun replaceText(file: File, text: String) {
        val temporary = File(directory, "${file.name}.tmp")
        temporary.writeText(text, Charsets.UTF_8)
        moveReplacing(temporary, file)
    }

    private fun moveReplacing(source: File, target: File) {
        try {
            Files.move(
                source.toPath(),
                target.toPath(),
                StandardCopyOption.ATOMIC_MOVE,
                StandardCopyOption.REPLACE_EXISTING,
            )
        } catch (_: AtomicMoveNotSupportedException) {
            Files.move(source.toPath(), target.toPath(), StandardCopyOption.REPLACE_EXISTING)
        }
    }

    private fun encodeEvent(event: DiagnosticEvent): String = buildString {
        append('{')
        appendJsonField("schemaVersion", DIAGNOSTICS_SCHEMA_VERSION)
        append(',')
        appendJsonField("requestId", event.requestId)
        append(',')
        appendJsonField("timestamp", event.timestampEpochMillis.toString(), quoted = false)
        append(',')
        appendJsonField("stage", event.stage)
        append(',')
        appendJsonField("outcome", event.outcome)
        event.elapsedMs?.let {
            append(',')
            appendJsonField("elapsedMs", it.toString(), quoted = false)
        }
        event.fields.toSortedMap().forEach { (key, value) ->
            append(',')
            appendJsonField(key, value)
        }
        append('}')
    }

    private fun encodeMetadata(): String = buildString {
        append('{')
        appendJsonField("schemaVersion", DIAGNOSTICS_SCHEMA_VERSION)
        append(',')
        appendJsonField("exportedAt", clock().toString(), quoted = false)
        append(',')
        appendJsonField("maxFiles", MAX_FILES.toString(), quoted = false)
        append(',')
        appendJsonField("maxFileBytes", MAX_FILE_BYTES.toString(), quoted = false)
        append(',')
        appendJsonField("maxRequests", MAX_REQUESTS.toString(), quoted = false)
        buildMetadata.toSortedMap().forEach { (key, value) ->
            append(',')
            appendJsonField(key, value)
        }
        append('}')
    }

    private fun StringBuilder.appendJsonField(key: String, value: String, quoted: Boolean = true) {
        append('"').append(key.escapeJson()).append("\":")
        if (quoted) append('"')
        append(value.escapeJson())
        if (quoted) append('"')
    }

    private fun String.escapeJson(): String = buildString(length) {
        for (character in this@escapeJson) {
            when (character) {
                '"' -> append("\\\"")
                '\\' -> append("\\\\")
                '\b' -> append("\\b")
                '\u000C' -> append("\\f")
                '\n' -> append("\\n")
                '\r' -> append("\\r")
                '\t' -> append("\\t")
                else -> if (character.code < 0x20) {
                    append("\\u").append(character.code.toString(16).padStart(4, '0'))
                } else {
                    append(character)
                }
            }
        }
    }

    private fun extractRequestId(line: String): String? = REQUEST_ID_PATTERN.find(line)?.groupValues?.get(1)

    private fun activeFile(): File = journalFile(0)
    private fun journalFile(index: Int): File = File(directory, "events-$index.jsonl")
    private fun journalFilesOldestFirst(): List<File> = (MAX_FILES - 1 downTo 0).map(::journalFile)

    companion object {
        private const val DIAGNOSTICS_SCHEMA_VERSION = "1"
        internal const val MAX_FILES = 3
        internal const val MAX_FILE_BYTES = 1024 * 1024
        internal const val MAX_REQUESTS = 100
        private const val COPY_BUFFER_BYTES = 8 * 1024
        private val REQUEST_ID_PATTERN = Regex("""\"requestId\":\"([^\"]+)\"""")
    }
}

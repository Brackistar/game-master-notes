package com.brackistar.gamemasternotes

import com.brackistar.gamemasternotes.core.domain.diagnostics.DiagnosticEvent
import kotlinx.coroutines.runBlocking
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.ByteArrayInputStream
import java.nio.file.Files
import java.util.zip.ZipInputStream

class FileDiagnosticsJournalTest {
    @Test
    fun journalKeepsAtMostOneHundredRequests() = runBlocking {
        val directory = Files.createTempDirectory("gmn-diagnostics").toFile()
        try {
            val journal = FileDiagnosticsJournal(directory, mapOf("buildId" to "test"), clock = { 5L })
            repeat(101) { index -> journal.record(event("request-$index")) }

            val archive = unzip(journal.exportArchive())
            val events = archive.filterKeys { it.endsWith(".jsonl") }.values.joinToString("\n")

            assertFalse(events.contains("\"requestId\":\"request-0\""))
            assertTrue(events.contains("\"requestId\":\"request-100\""))
            assertEquals(100, Regex("\"requestId\"").findAll(events).count())
        } finally {
            directory.deleteRecursively()
        }
    }

    @Test
    fun journalRotatesAtBoundAndExportsMetadataWithoutRawContent() = runBlocking {
        val directory = Files.createTempDirectory("gmn-diagnostics").toFile()
        try {
            val journal = FileDiagnosticsJournal(
                directory,
                mapOf("buildId" to "build-1", "deviceModel" to "tablet"),
                clock = { 9L },
            )
            val boundedPayload = "x".repeat(600_000)
            repeat(5) { index ->
                journal.record(event("request-$index", mapOf("boundedMetric" to boundedPayload)))
            }

            val archive = unzip(journal.exportArchive())

            assertTrue(archive.keys.count { it.endsWith(".jsonl") } <= FileDiagnosticsJournal.MAX_FILES)
            assertTrue(archive.getValue("metadata.json").contains("\"buildId\":\"build-1\""))
            assertFalse(archive.values.any { it.contains("secret question") })
        } finally {
            directory.deleteRecursively()
        }
    }

    @Test
    fun clearRemovesAllJournalEvents() = runBlocking {
        val directory = Files.createTempDirectory("gmn-diagnostics").toFile()
        try {
            val journal = FileDiagnosticsJournal(directory, emptyMap())
            journal.record(event("request-1"))

            journal.clear()

            val archive = unzip(journal.exportArchive())
            assertTrue(archive.keys.none { it.endsWith(".jsonl") })
        } finally {
            directory.deleteRecursively()
        }
    }

    private fun event(requestId: String, fields: Map<String, String> = emptyMap()) = DiagnosticEvent(
        requestId = requestId,
        timestampEpochMillis = 1L,
        stage = "test",
        outcome = "completed",
        fields = fields,
    )

    private fun unzip(archive: ByteArray): Map<String, String> {
        val entries = mutableMapOf<String, String>()
        ZipInputStream(ByteArrayInputStream(archive)).use { zip ->
            var entry = zip.nextEntry
            while (entry != null) {
                entries[entry.name] = zip.readBytes().decodeToString()
                entry = zip.nextEntry
            }
        }
        return entries
    }
}

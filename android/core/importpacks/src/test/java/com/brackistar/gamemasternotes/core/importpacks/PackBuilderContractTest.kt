package com.brackistar.gamemasternotes.core.importpacks

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import java.util.zip.ZipInputStream

class PackBuilderContractTest {
    @Test
    fun realBuilderFixtureMatchesAndroidV1Contract() {
        val input = requireNotNull(javaClass.classLoader?.getResourceAsStream("v1/synthetic.gmnpack"))
        val entries = mutableMapOf<String, ByteArray>()
        ZipInputStream(input).use { zip ->
            var entry = zip.nextEntry
            while (entry != null) {
                if (!entry.isDirectory) entries[entry.name] = zip.readBytes()
                entry = zip.nextEntry
            }
        }

        assertEquals(
            setOf("manifest.json", "documents.json", "chunks.jsonl", "embeddings.npy", "extraction-report.json"),
            entries.keys,
        )
        val manifest = entries.getValue("manifest.json").decodeToString()
        assertTrue(manifest.contains("\"schema_version\": \"1.0\""))
        assertTrue(manifest.contains("\"pack_id\": \"fixture-pack\""))
        assertTrue(manifest.contains("\"embedding_dimensions\": 384"))
        assertTrue(manifest.contains("\"chunk_count\": 2"))

        val documents = entries.getValue("documents.json").decodeToString()
        assertTrue(documents.contains("\"document_id\": \"fixture-doc\""))
        val chunks = entries.getValue("chunks.jsonl").decodeToString().lineSequence().filter(String::isNotBlank).toList()
        assertEquals(2, chunks.size)
        assertTrue(chunks[0].contains("\"citation_label\": \"Fixture Codex p. 1\""))
        assertTrue(chunks[1].contains("\"embedding_row_index\": 1"))

        val npy = entries.getValue("embeddings.npy")
        assertTrue(npy.copyOfRange(0, 6).contentEquals(byteArrayOf(0x93.toByte(), 0x4e, 0x55, 0x4d, 0x50, 0x59)))
    }
}

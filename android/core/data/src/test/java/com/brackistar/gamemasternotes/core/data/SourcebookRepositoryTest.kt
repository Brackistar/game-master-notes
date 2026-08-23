package com.brackistar.gamemasternotes.core.data

import androidx.test.core.app.ApplicationProvider
import com.brackistar.gamemasternotes.core.retrieval.RetrievalQuery
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

@RunWith(RobolectricTestRunner::class)
class SourcebookRepositoryTest {
    private lateinit var database: AppDatabase
    private lateinit var repository: SourcebookRepository

    @Before
    fun setUp() {
        database = AppDatabase.createInMemory(ApplicationProvider.getApplicationContext())
        repository = SourcebookRepository(database)
    }

    @After
    fun tearDown() {
        database.close()
    }

    @Test
    fun importedPackAppearsInLibraryAndSearch() = runTest {
        repository.replaceImportedPack(testPack(text = "The Silver Ladder guards the hidden library."))

        val packs = repository.observePacks().first()
        val results = repository.search(RetrievalQuery("Silver Ladder"))

        assertEquals(1, packs.size)
        assertEquals("Synthetic Book", packs.single().title)
        assertEquals(1, results.size)
        assertTrue(results.single().snippet.contains("Silver Ladder"))
        assertEquals("Synthetic Book, pp. 4-5", results.single().citationLabel)
    }

    @Test
    fun replacingPackDoesNotDuplicateRows() = runTest {
        repository.replaceImportedPack(testPack(text = "First text."))
        repository.replaceImportedPack(testPack(text = "Second text about Atlantis."))

        val packs = repository.observePacks().first()
        val results = repository.search(RetrievalQuery("Atlantis"))

        assertEquals(1, packs.size)
        assertEquals(1, results.size)
        assertTrue(results.single().snippet.contains("Second text"))
    }

    @Test
    fun pruningRemovedPacksRemovesSearchRows() = runTest {
        repository.replaceImportedPack(testPack(text = "A vanished citadel."))
        repository.pruneToAvailablePacks(emptyList())

        assertEquals(0, repository.observePacks().first().size)
        assertEquals(emptyList<Any>(), repository.search(RetrievalQuery("citadel")))
    }

    @Test
    fun searchDoesNotReturnWeakPartialMatchesForMultiTermQuestions() = runTest {
        repository.replaceImportedPack(testPack(text = "The ladder is stored near a mundane shed."))

        val results = repository.search(RetrievalQuery("Silver Ladder"))

        assertEquals(emptyList<Any>(), results)
    }

    @Test
    fun searchReturnsRelevantParagraphInsteadOfChunkBeginning() = runTest {
        repository.replaceImportedPack(
            testPack(
                text = """
                    Opening fiction with unrelated imagery.

                    The Silver Ladder guards the hidden library. Its members preserve the rites and laws of awakened society.

                    Closing text.
                """.trimIndent(),
            ),
        )

        val results = repository.search(RetrievalQuery("Silver Ladder"))

        assertEquals(1, results.size)
        assertEquals(
            "The Silver Ladder guards the hidden library. Its members preserve the rites and laws of awakened society.",
            results.single().snippet,
        )
    }

    private fun testPack(text: String): ImportedPack {
        val pack = SourcebookPackEntity(
            packId = "pack-1",
            title = "Synthetic Book",
            system = "Test System",
            edition = "1e",
            language = "en",
            schemaVersion = 1,
            generatorVersion = "test",
            embeddingModelId = "deterministic",
            embeddingDimensions = 384,
            chunkCount = 1,
            sourceFolderUri = "content://folder",
            sourceDisplayName = "synthetic.gmnpack",
            archiveDocumentUri = "content://folder/synthetic.gmnpack",
            archiveFingerprint = "fingerprint",
            importedAtEpochMillis = 1L,
        )
        val document = SourceDocumentEntity(
            documentId = "doc-1",
            packId = "pack-1",
            sourceFilename = "synthetic.pdf",
            sourceChecksum = "checksum",
            pageCount = 10,
        )
        val chunk = SourceChunkEntity(
            chunkId = "chunk-1",
            packId = "pack-1",
            documentId = "doc-1",
            pageStart = 4,
            pageEnd = 5,
            citationLabel = "Synthetic Book pp. 4-5",
            text = text,
            charCount = text.length,
            embeddingRowIndex = 0,
        )
        return ImportedPack(
            pack = pack,
            documents = listOf(document),
            chunks = listOf(chunk),
            ftsRows = listOf(
                SourceChunkFtsEntity(
                    chunkId = chunk.chunkId,
                    packId = pack.packId,
                    title = pack.title,
                    system = pack.system,
                    text = chunk.text,
                ),
            ),
        )
    }
}

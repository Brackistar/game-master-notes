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
        assertEquals("Synthetic Book pp. 4-5", results.single().citationLabel)
        assertEquals("pack-1", results.single().packId)
        assertEquals("doc-1", results.single().documentId)
        assertEquals("chunk-1", results.single().sourceId)
        assertEquals(4, results.single().pageStart)
        assertEquals(5, results.single().pageEnd)
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
        assertFtsMatchesActiveChunks()
    }

    @Test
    fun replacingPackRemovesOldFtsTerms() = runTest {
        repository.replaceImportedPack(testPack(text = "The old citadel sank below the marsh."))
        repository.replaceImportedPack(testPack(text = "The new observatory studies clear stars."))

        assertEquals(emptyList<Any>(), repository.search(RetrievalQuery("citadel marsh")))
        assertEquals(1, repository.search(RetrievalQuery("observatory stars")).size)
        assertFtsMatchesActiveChunks()
    }

    @Test
    fun pruningRemovedPacksRemovesSearchRows() = runTest {
        repository.replaceImportedPack(testPack(text = "A vanished citadel."))
        repository.pruneToAvailablePacks(emptyList())

        assertEquals(0, repository.observePacks().first().size)
        assertEquals(emptyList<Any>(), repository.search(RetrievalQuery("citadel")))
        assertFtsMatchesActiveChunks()
    }

    @Test
    fun pruningSubsetKeepsOnlyAvailablePackFtsRows() = runTest {
        repository.replaceImportedPack(
            testPack(packId = "pack-1", chunkId = "chunk-1", text = "A silver bridge."),
        )
        repository.replaceImportedPack(
            testPack(packId = "pack-2", chunkId = "chunk-2", text = "A golden library."),
        )

        repository.pruneToAvailablePacks(listOf("pack-2"))

        assertEquals(emptyList<Any>(), repository.search(RetrievalQuery("silver bridge")))
        assertEquals(1, repository.search(RetrievalQuery("golden library")).size)
        assertFtsMatchesActiveChunks()
    }

    @Test
    fun replacingPackRejectsDuplicateIncomingChunkIds() = runTest {
        val pack = testPack(text = "One chunk.").let { imported ->
            imported.copy(
                chunks = imported.chunks + imported.chunks.single().copy(text = "Duplicate chunk."),
                ftsRows = imported.ftsRows + imported.ftsRows.single().copy(text = "Duplicate chunk."),
            )
        }

        val error = runCatching { repository.replaceImportedPack(pack) }.exceptionOrNull()

        assertTrue(error is IllegalArgumentException)
        assertTrue(error?.message.orEmpty().contains("duplicate chunk ID"))
        assertFtsMatchesActiveChunks()
    }

    @Test
    fun replacingPackRejectsChunkIdOwnedByAnotherPack() = runTest {
        repository.replaceImportedPack(
            testPack(packId = "pack-1", chunkId = "shared-chunk", text = "First owner text."),
        )

        val error = runCatching {
            repository.replaceImportedPack(
                testPack(packId = "pack-2", chunkId = "shared-chunk", text = "Second owner text."),
            )
        }.exceptionOrNull()

        assertTrue(error is IllegalArgumentException)
        assertTrue(error?.message.orEmpty().contains("owned by pack pack-1"))
        assertEquals(1, repository.search(RetrievalQuery("First owner")).size)
        assertEquals(emptyList<Any>(), repository.search(RetrievalQuery("Second owner")))
        assertFtsMatchesActiveChunks()
    }

    @Test
    fun searchDoesNotReturnWeakPartialMatchesForMultiTermQuestions() = runTest {
        repository.replaceImportedPack(testPack(text = "The ladder is stored near a mundane shed."))

        val results = repository.search(RetrievalQuery("Silver Ladder"))

        assertEquals(emptyList<Any>(), results)
    }

    @Test
    fun searchRelaxesNaturalLanguageQuestionsWithoutAcceptingSingleTermMatches() = runTest {
        repository.replaceImportedPack(
            testPack(text = "The Silver Ladder protects the hidden library."),
        )

        val results = repository.search(RetrievalQuery("How does the Silver Ladder protect lore?"))

        assertEquals(1, results.size)
        assertTrue(results.single().snippet.contains("Silver Ladder"))
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

    @Test
    fun searchReturnsMultipleRelevantParagraphsFromOneChunk() = runTest {
        repository.replaceImportedPack(
            testPack(
                text = """
                    The Silver Ladder guards the hidden library.

                    The Silver Ladder members preserve the rites and laws of awakened society.

                    The Silver Ladder also trains archivists to protect dangerous lore.
                """.trimIndent(),
            ),
        )

        val result = repository.search(RetrievalQuery("Silver Ladder"))

        assertEquals(1, result.size)
        assertTrue(result.single().snippet.contains("The Silver Ladder guards the hidden library."))
        assertTrue(result.single().snippet.contains("The Silver Ladder also trains archivists"))
    }

    @Test
    fun rankingRecoversStrongHitAfterOldRowLimit() = runTest {
        val weak = "silver filler filler ladder filler filler protects filler filler moon filler filler archive"
        val texts = List(40) { index -> "$weak weak-$index" } +
            "silver ladder protects moon archive with an exact compact rule"
        repository.replaceImportedPack(testPack(texts))

        val results = repository.search(RetrievalQuery("silver ladder protects moon archive", limit = 4))

        assertEquals("chunk-40", results.first().sourceId)
    }

    private suspend fun assertFtsMatchesActiveChunks() {
        val dao = database.sourcebookDao()
        assertEquals(dao.activeChunkCount(), dao.chunkFtsRowCount())
        assertEquals(0, dao.orphanChunkFtsRowCount())
        assertEquals(emptyList<FtsDuplicateRow>(), dao.duplicateChunkFtsRows())
    }

    private fun testPack(
        packId: String = "pack-1",
        chunkId: String = "chunk-1",
        text: String,
    ): ImportedPack {
        val numericSuffix = packId.substringAfterLast("-").toIntOrNull() ?: 1
        val pack = SourcebookPackEntity(
            packId = packId,
            title = if (packId == "pack-1") "Synthetic Book" else "Synthetic Book $numericSuffix",
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
            archiveFingerprint = "fingerprint-$packId-${text.hashCode()}",
            importedAtEpochMillis = 1L,
        )
        val document = SourceDocumentEntity(
            documentId = "doc-1",
            packId = packId,
            sourceFilename = "synthetic.pdf",
            sourceChecksum = "checksum",
            pageCount = 10,
        )
        val chunk = SourceChunkEntity(
            chunkId = chunkId,
            packId = packId,
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

    private fun testPack(texts: List<String>): ImportedPack {
        val first = testPack(text = texts.first())
        val chunks = texts.mapIndexed { index, text ->
            first.chunks.single().copy(
                chunkId = "chunk-$index",
                pageStart = index + 1,
                pageEnd = index + 1,
                citationLabel = "Synthetic Book p. ${index + 1}",
                text = text,
                charCount = text.length,
                embeddingRowIndex = index,
            )
        }
        return first.copy(
            pack = first.pack.copy(chunkCount = chunks.size),
            chunks = chunks,
            ftsRows = chunks.map { chunk ->
                first.ftsRows.single().copy(
                    chunkId = chunk.chunkId,
                    text = chunk.text,
                )
            },
        )
    }
}

package com.brackistar.gamemasternotes.core.data

import com.brackistar.gamemasternotes.core.retrieval.RetrievalQuery
import com.brackistar.gamemasternotes.core.retrieval.RetrievalRepository
import com.brackistar.gamemasternotes.core.retrieval.RetrievalResult
import kotlinx.coroutines.flow.Flow

class SourcebookRepository(
    private val database: AppDatabase,
) : RetrievalRepository {
    private val dao = database.sourcebookDao()

    fun observePackCount(): Flow<Int> = dao.observePackCount()

    fun observePacks(): Flow<List<SourcebookPackSummary>> = dao.observePacks()

    suspend fun replaceImportedPack(pack: ImportedPack) {
        database.replaceImportedPack(pack)
    }

    suspend fun existingFingerprint(packId: String): String? = dao.archiveFingerprint(packId)

    suspend fun packIds(): Set<String> = dao.packIds().toSet()

    suspend fun pruneToAvailablePacks(packIds: List<String>) {
        if (packIds.isEmpty()) {
            dao.deleteAllPacks()
        } else {
            dao.deletePacksNotIn(packIds)
        }
    }

    override suspend fun search(query: RetrievalQuery): List<RetrievalResult> {
        val ftsQuery = query.text
            .split(Regex("\\s+"))
            .map { it.trim().replace("\"", "") }
            .filter { it.length >= 2 }
            .joinToString(" OR ") { "\"$it\"" }
        if (ftsQuery.isBlank()) return emptyList()

        return dao.searchChunks(ftsQuery, limit = 8).map { row ->
            RetrievalResult(
                sourceId = row.chunkId,
                title = row.packTitle,
                snippet = row.text.take(700),
                citationLabel = "${row.packTitle}, pp. ${row.pageStart}-${row.pageEnd}",
                score = -row.rank,
            )
        }
    }
}

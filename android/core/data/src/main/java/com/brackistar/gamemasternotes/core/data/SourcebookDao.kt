package com.brackistar.gamemasternotes.core.data

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import kotlinx.coroutines.flow.Flow

@Dao
interface SourcebookDao {
    @Query("SELECT COUNT(*) FROM sourcebook_packs")
    fun observePackCount(): Flow<Int>

    @Query(
        """
        SELECT packId, title, system, edition, chunkCount, sourceDisplayName
        FROM sourcebook_packs
        ORDER BY title COLLATE NOCASE
        """,
    )
    fun observePacks(): Flow<List<SourcebookPackSummary>>

    @Query("SELECT packId FROM sourcebook_packs")
    suspend fun packIds(): List<String>

    @Query("SELECT archiveFingerprint FROM sourcebook_packs WHERE packId = :packId")
    suspend fun archiveFingerprint(packId: String): String?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertPack(pack: SourcebookPackEntity)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertDocuments(documents: List<SourceDocumentEntity>)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertChunks(chunks: List<SourceChunkEntity>)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertChunkFts(rows: List<SourceChunkFtsEntity>)

    @Query("DELETE FROM sourcebook_packs WHERE packId = :packId")
    suspend fun deletePackById(packId: String)

    @Query("DELETE FROM sourcebook_packs WHERE packId NOT IN (:packIds)")
    suspend fun deletePacksNotIn(packIds: List<String>)

    @Query("DELETE FROM sourcebook_packs")
    suspend fun deleteAllPacks()

    @Query("DELETE FROM source_chunks_fts WHERE packId = :packId")
    suspend fun deleteChunkFtsByPackId(packId: String)

    @Query("DELETE FROM source_chunks_fts WHERE packId IN (:packIds)")
    suspend fun deleteChunkFtsByPackIds(packIds: List<String>)

    @Query("DELETE FROM source_chunks_fts")
    suspend fun deleteAllChunkFts()

    @Query(
        """
        SELECT chunkId, packId
        FROM source_chunks
        WHERE chunkId IN (:chunkIds) AND packId != :packId
        """,
    )
    suspend fun chunkOwnersOutsidePack(chunkIds: List<String>, packId: String): List<ChunkOwnerRow>

    @Query("SELECT COUNT(*) FROM source_chunks")
    suspend fun activeChunkCount(): Int

    @Query("SELECT COUNT(*) FROM source_chunks_fts")
    suspend fun chunkFtsRowCount(): Int

    @Query(
        """
        SELECT COUNT(*)
        FROM source_chunks_fts f
        LEFT JOIN source_chunks c
            ON f.chunkId = c.chunkId AND f.packId = c.packId
        WHERE c.chunkId IS NULL
        """,
    )
    suspend fun orphanChunkFtsRowCount(): Int

    @Query(
        """
        SELECT packId, chunkId, COUNT(*) AS rowCount
        FROM source_chunks_fts
        GROUP BY packId, chunkId
        HAVING COUNT(*) > 1
        """,
    )
    suspend fun duplicateChunkFtsRows(): List<FtsDuplicateRow>

    @Query(
        """
        SELECT c.chunkId, c.packId, c.documentId, p.title AS packTitle, p.system, c.pageStart, c.pageEnd,
               c.citationLabel, c.text, 0.0 AS rank
        FROM source_chunks_fts
        JOIN source_chunks c
            ON source_chunks_fts.chunkId = c.chunkId
            AND source_chunks_fts.packId = c.packId
        JOIN sourcebook_packs p ON c.packId = p.packId
        WHERE source_chunks_fts MATCH :query
        ORDER BY source_chunks_fts.rowid
        LIMIT :limit
        """,
    )
    suspend fun searchChunks(query: String, limit: Int): List<SourceChunkSearchRow>

    @Query(
        """
        SELECT c.chunkId, c.packId, c.documentId, p.title AS packTitle, p.system, c.pageStart, c.pageEnd,
               c.citationLabel, c.text, 0.0 AS rank
        FROM source_chunks c
        JOIN sourcebook_packs p ON c.packId = p.packId
        WHERE c.chunkId IN (:chunkIds)
        """,
    )
    suspend fun chunksByIds(chunkIds: List<String>): List<SourceChunkSearchRow>

    @Query(
        """
        SELECT DISTINCT c.chunkId, c.packId, c.documentId, p.title AS packTitle, p.system,
               c.pageStart, c.pageEnd, c.citationLabel, c.text, 0.0 AS rank
        FROM source_chunks seed
        JOIN source_chunks c ON c.documentId = seed.documentId AND c.chunkId != seed.chunkId
        JOIN sourcebook_packs p ON c.packId = p.packId
        WHERE seed.chunkId IN (:seedIds)
          AND (
            ABS(c.embeddingRowIndex - seed.embeddingRowIndex) = 1
            OR (c.pageStart <= seed.pageEnd AND seed.pageStart <= c.pageEnd)
          )
        ORDER BY c.documentId, c.embeddingRowIndex
        LIMIT :limit
        """,
    )
    suspend fun relatedChunks(seedIds: List<String>, limit: Int): List<SourceChunkSearchRow>
}

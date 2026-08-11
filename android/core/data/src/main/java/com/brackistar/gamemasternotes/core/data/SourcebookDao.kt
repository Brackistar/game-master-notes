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

    @Query(
        """
        SELECT c.chunkId, c.packId, p.title AS packTitle, p.system, c.pageStart, c.pageEnd,
               c.citationLabel, c.text, 0.0 AS rank
        FROM source_chunks_fts
        JOIN source_chunks c ON source_chunks_fts.chunkId = c.chunkId
        JOIN sourcebook_packs p ON c.packId = p.packId
        WHERE source_chunks_fts MATCH :query
        ORDER BY source_chunks_fts.rowid
        LIMIT :limit
        """,
    )
    suspend fun searchChunks(query: String, limit: Int): List<SourceChunkSearchRow>
}

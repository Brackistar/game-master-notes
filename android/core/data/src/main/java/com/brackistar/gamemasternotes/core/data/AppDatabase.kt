package com.brackistar.gamemasternotes.core.data

import android.content.Context
import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import androidx.room.withTransaction

@Database(
    entities = [
        SourcebookPackEntity::class,
        SourceDocumentEntity::class,
        SourceChunkEntity::class,
        SourceChunkFtsEntity::class,
    ],
    version = 1,
    exportSchema = false,
)
abstract class AppDatabase : RoomDatabase() {
    abstract fun sourcebookDao(): SourcebookDao

    companion object {
        fun create(context: Context): AppDatabase =
            Room.databaseBuilder(
                context.applicationContext,
                AppDatabase::class.java,
                "game-master-notes.db",
            ).build()

        fun createInMemory(context: Context): AppDatabase =
            Room.inMemoryDatabaseBuilder(
                context.applicationContext,
                AppDatabase::class.java,
            ).build()
    }
}

suspend fun AppDatabase.replaceImportedPack(pack: ImportedPack) {
    val dao = sourcebookDao()
    val incomingChunkIds = pack.chunks.map { it.chunkId }
    val duplicateIncomingChunkId = incomingChunkIds
        .groupingBy { it }
        .eachCount()
        .filterValues { count -> count > 1 }
        .keys
        .firstOrNull()
    require(duplicateIncomingChunkId == null) {
        "Imported pack ${pack.pack.packId} contains duplicate chunk ID $duplicateIncomingChunkId."
    }

    withTransaction {
        val conflictingOwner = if (incomingChunkIds.isEmpty()) {
            null
        } else {
            dao.chunkOwnersOutsidePack(incomingChunkIds, pack.pack.packId).firstOrNull()
        }
        require(conflictingOwner == null) {
            "Imported pack ${pack.pack.packId} reuses chunk ID ${conflictingOwner?.chunkId} " +
                "owned by pack ${conflictingOwner?.packId}."
        }

        dao.deleteChunkFtsByPackId(pack.pack.packId)
        dao.deletePackById(pack.pack.packId)
        dao.insertPack(pack.pack)
        dao.insertDocuments(pack.documents)
        dao.insertChunks(pack.chunks)
        dao.insertChunkFts(pack.ftsRows)
    }
}

suspend fun AppDatabase.pruneToAvailablePacks(packIds: List<String>) {
    val dao = sourcebookDao()
    withTransaction {
        if (packIds.isEmpty()) {
            dao.deleteAllChunkFts()
            dao.deleteAllPacks()
        } else {
            val removedPackIds = dao.packIds() - packIds.toSet()
            if (removedPackIds.isNotEmpty()) {
                dao.deleteChunkFtsByPackIds(removedPackIds)
            }
            dao.deletePacksNotIn(packIds)
        }
    }
}

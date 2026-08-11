package com.brackistar.gamemasternotes.core.data

import android.content.Context
import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import androidx.room.Transaction

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

@Transaction
suspend fun AppDatabase.replaceImportedPack(pack: ImportedPack) {
    sourcebookDao().deletePackById(pack.pack.packId)
    sourcebookDao().insertPack(pack.pack)
    sourcebookDao().insertDocuments(pack.documents)
    sourcebookDao().insertChunks(pack.chunks)
    sourcebookDao().insertChunkFts(pack.ftsRows)
}

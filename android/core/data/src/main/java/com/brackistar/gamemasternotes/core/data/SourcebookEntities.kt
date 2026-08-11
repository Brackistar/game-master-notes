package com.brackistar.gamemasternotes.core.data

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.ForeignKey
import androidx.room.Fts4
import androidx.room.Index
import androidx.room.PrimaryKey

@Entity(tableName = "sourcebook_packs")
data class SourcebookPackEntity(
    @PrimaryKey val packId: String,
    val title: String,
    val system: String,
    val edition: String,
    val language: String,
    val schemaVersion: Int,
    val generatorVersion: String,
    val embeddingModelId: String?,
    val embeddingDimensions: Int?,
    val chunkCount: Int,
    val sourceFolderUri: String,
    val sourceDisplayName: String,
    val archiveDocumentUri: String,
    val archiveFingerprint: String,
    val importedAtEpochMillis: Long,
)

@Entity(
    tableName = "source_documents",
    primaryKeys = ["packId", "documentId"],
    foreignKeys = [
        ForeignKey(
            entity = SourcebookPackEntity::class,
            parentColumns = ["packId"],
            childColumns = ["packId"],
            onDelete = ForeignKey.CASCADE,
        ),
    ],
    indices = [Index("packId")],
)
data class SourceDocumentEntity(
    val documentId: String,
    val packId: String,
    val sourceFilename: String,
    val sourceChecksum: String,
    val pageCount: Int,
)

@Entity(
    tableName = "source_chunks",
    foreignKeys = [
        ForeignKey(
            entity = SourcebookPackEntity::class,
            parentColumns = ["packId"],
            childColumns = ["packId"],
            onDelete = ForeignKey.CASCADE,
        ),
    ],
    indices = [Index("packId"), Index("documentId")],
)
data class SourceChunkEntity(
    @PrimaryKey val chunkId: String,
    val packId: String,
    val documentId: String,
    val pageStart: Int,
    val pageEnd: Int,
    val citationLabel: String,
    val text: String,
    val charCount: Int,
    val embeddingRowIndex: Int,
)

@Fts4
@Entity(tableName = "source_chunks_fts")
data class SourceChunkFtsEntity(
    @PrimaryKey(autoGenerate = true)
    @ColumnInfo(name = "rowid")
    val rowId: Int = 0,
    val chunkId: String,
    val packId: String,
    val title: String,
    val system: String,
    val text: String,
)

data class ImportedPack(
    val pack: SourcebookPackEntity,
    val documents: List<SourceDocumentEntity>,
    val chunks: List<SourceChunkEntity>,
    val ftsRows: List<SourceChunkFtsEntity>,
)

data class SourcebookPackSummary(
    val packId: String,
    val title: String,
    val system: String,
    val edition: String,
    val chunkCount: Int,
    val sourceDisplayName: String,
)

data class SourceChunkSearchRow(
    val chunkId: String,
    val packId: String,
    val packTitle: String,
    val system: String,
    val pageStart: Int,
    val pageEnd: Int,
    val citationLabel: String,
    val text: String,
    val rank: Double,
)

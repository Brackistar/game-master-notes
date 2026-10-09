package com.brackistar.gamemasternotes.core.domain.diagnostics

data class DiagnosticEvent(
    val requestId: String,
    val timestampEpochMillis: Long,
    val stage: String,
    val outcome: String,
    val elapsedMs: Long? = null,
    val fields: Map<String, String> = emptyMap(),
)

interface DiagnosticsJournal {
    suspend fun record(event: DiagnosticEvent)
    suspend fun exportArchive(): ByteArray
    suspend fun clear()
}

object NoOpDiagnosticsJournal : DiagnosticsJournal {
    override suspend fun record(event: DiagnosticEvent) = Unit
    override suspend fun exportArchive(): ByteArray = ByteArray(0)
    override suspend fun clear() = Unit
}

package com.brackistar.gamemasternotes.feature.assistant

import android.net.Uri

fun interface DiagnosticsArchiveWriter {
    suspend fun write(uri: Uri, archive: ByteArray)
}

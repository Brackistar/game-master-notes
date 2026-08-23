package com.brackistar.gamemasternotes.core.ai

import android.content.Context
import android.net.Uri
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.File

class ModelFileInstaller(
    context: Context,
    private val modelsDirectory: File = context.applicationContext.filesDir.resolve("models"),
) {
    private val contentResolver = context.applicationContext.contentResolver

    fun expectedFileName(modelId: String): String? =
        LocalModelProfiles.All.firstOrNull { it.model.id == modelId }?.modelFileName

    fun primaryModelId(): String = LocalModelProfiles.Lfm25TinyQ2.model.id

    suspend fun importModel(modelId: String, sourceUri: Uri): InstalledModel = withContext(Dispatchers.IO) {
        val fileName = requireNotNull(expectedFileName(modelId)) {
            "Unknown model id: $modelId"
        }
        require(fileName.endsWith(".gguf")) {
            "Expected a GGUF model file."
        }

        modelsDirectory.mkdirs()
        val destination = modelsDirectory.resolve(fileName)
        val temporaryDestination = modelsDirectory.resolve("$fileName.tmp")
        if (temporaryDestination.exists()) temporaryDestination.delete()
        contentResolver.openInputStream(sourceUri).use { input ->
            requireNotNull(input) { "Could not open selected model file." }
            temporaryDestination.outputStream().use { output ->
                input.copyTo(output)
            }
        }
        if (temporaryDestination.length() == 0L) {
            temporaryDestination.delete()
            error("Selected model file was empty.")
        }
        if (destination.exists()) destination.delete()
        require(temporaryDestination.renameTo(destination)) {
            "Could not finish importing the model file."
        }

        InstalledModel(
            modelId = modelId,
            fileName = fileName,
            sizeBytes = destination.length(),
        )
    }
}

data class InstalledModel(
    val modelId: String,
    val fileName: String,
    val sizeBytes: Long,
)

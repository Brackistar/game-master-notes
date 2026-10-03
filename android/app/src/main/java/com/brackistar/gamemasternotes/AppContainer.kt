package com.brackistar.gamemasternotes

import android.content.Context
import android.os.Build
import com.brackistar.gamemasternotes.core.ai.AndroidDeviceAiProfileReader
import com.brackistar.gamemasternotes.core.ai.LlamaCppLocalModelRuntime
import com.brackistar.gamemasternotes.core.ai.ModelSelectingAiEngine
import com.brackistar.gamemasternotes.core.ai.ModelFileInstaller
import com.brackistar.gamemasternotes.core.data.AppDatabase
import com.brackistar.gamemasternotes.core.data.SourcebookRepository
import com.brackistar.gamemasternotes.core.importpacks.ContentResolverPackImporter
import com.brackistar.gamemasternotes.core.importpacks.PackFolderStore
import com.brackistar.gamemasternotes.core.importpacks.VectorSidecarStore
import com.brackistar.gamemasternotes.core.retrieval.HybridRetrievalRepository
import com.brackistar.gamemasternotes.core.retrieval.MiniLmOnnxQueryEncoder
import com.brackistar.gamemasternotes.core.retrieval.NpyVectorSidecarStore
import com.brackistar.gamemasternotes.core.retrieval.RetrievalResultResolver
import com.brackistar.gamemasternotes.core.retrieval.RelatedResultResolver
import com.brackistar.gamemasternotes.feature.assistant.DiagnosticsArchiveWriter
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

class AppContainer(context: Context) {
    private val appContext = context.applicationContext
    private val database = AppDatabase.create(appContext)
    val deviceAiProfile = AndroidDeviceAiProfileReader.read(appContext)
    val diagnosticsJournal = FileDiagnosticsJournal(
        directory = appContext.filesDir.resolve("diagnostics"),
        buildMetadata = mapOf(
            "buildId" to Build.ID,
            "deviceModel" to Build.MODEL,
            "androidVersion" to Build.VERSION.RELEASE,
            "abi" to Build.SUPPORTED_ABIS.joinToString(","),
            "lowRam" to deviceAiProfile.isLowRamDevice.toString(),
            "memoryClassMb" to deviceAiProfile.totalRamMb.toString(),
            "appVersion" to appContext.packageManager
                .getPackageInfo(appContext.packageName, 0)
                .versionName
                .orEmpty(),
        ),
    )
    val diagnosticsArchiveWriter = DiagnosticsArchiveWriter { uri, archive ->
        withContext(Dispatchers.IO) {
            appContext.contentResolver.openOutputStream(uri, "wt")?.use { output ->
                output.write(archive)
            } ?: error("Could not open the diagnostics export destination.")
        }
    }
    val sourcebookRepository = SourcebookRepository(database, diagnosticsJournal = diagnosticsJournal)
    val packFolderStore = PackFolderStore(appContext)
    private val vectorSidecarDirectory = appContext.filesDir.resolve("vector-index")
    private val vectorSidecarStore = VectorSidecarStore(vectorSidecarDirectory)
    val packImporter = ContentResolverPackImporter(appContext, sourcebookRepository, vectorSidecarStore)
    private val queryEncoder = MiniLmOnnxQueryEncoder(appContext)
    val assistantRetrievalRepository = HybridRetrievalRepository(
        lexicalRepository = sourcebookRepository,
        encoder = queryEncoder,
        vectorStore = NpyVectorSidecarStore(
            directory = vectorSidecarDirectory,
            modelId = MiniLmOnnxQueryEncoder.MODEL_ID,
            modelRevision = MiniLmOnnxQueryEncoder.MODEL_REVISION,
            dimensions = MiniLmOnnxQueryEncoder.DIMENSIONS,
        ),
        resultResolver = RetrievalResultResolver(sourcebookRepository::resultsBySourceIds),
        relatedResultResolver = RelatedResultResolver(sourcebookRepository::relatedResults),
        diagnosticsJournal = diagnosticsJournal,
    )
    private val modelsDirectory = appContext.filesDir.resolve("models")
    val modelFileInstaller = ModelFileInstaller(appContext, modelsDirectory)
    val aiEngine = ModelSelectingAiEngine(
        deviceProfile = deviceAiProfile,
        runtime = LlamaCppLocalModelRuntime(
            modelsDirectory = modelsDirectory,
            deviceProfile = deviceAiProfile,
            diagnosticsJournal = diagnosticsJournal,
        ),
        isModelFileInstalled = { profile ->
            modelsDirectory.resolve(profile.modelFileName).canRead()
        },
    )
}

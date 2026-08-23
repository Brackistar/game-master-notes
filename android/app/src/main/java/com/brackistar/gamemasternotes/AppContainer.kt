package com.brackistar.gamemasternotes

import android.content.Context
import com.brackistar.gamemasternotes.core.ai.AndroidDeviceAiProfileReader
import com.brackistar.gamemasternotes.core.ai.LlamaCppLocalModelRuntime
import com.brackistar.gamemasternotes.core.ai.ModelSelectingAiEngine
import com.brackistar.gamemasternotes.core.ai.ModelFileInstaller
import com.brackistar.gamemasternotes.core.data.AppDatabase
import com.brackistar.gamemasternotes.core.data.SourcebookRepository
import com.brackistar.gamemasternotes.core.importpacks.ContentResolverPackImporter
import com.brackistar.gamemasternotes.core.importpacks.PackFolderStore

class AppContainer(context: Context) {
    private val appContext = context.applicationContext
    private val database = AppDatabase.create(appContext)

    val sourcebookRepository = SourcebookRepository(database)
    val packFolderStore = PackFolderStore(appContext)
    val packImporter = ContentResolverPackImporter(appContext, sourcebookRepository)
    val deviceAiProfile = AndroidDeviceAiProfileReader.read(appContext)
    private val modelsDirectory = appContext.filesDir.resolve("models")
    val modelFileInstaller = ModelFileInstaller(appContext, modelsDirectory)
    val aiEngine = ModelSelectingAiEngine(
        deviceProfile = deviceAiProfile,
        runtime = LlamaCppLocalModelRuntime(
            modelsDirectory = modelsDirectory,
            deviceProfile = deviceAiProfile,
        ),
        isModelFileInstalled = { profile ->
            modelsDirectory.resolve(profile.modelFileName).canRead()
        },
    )
}

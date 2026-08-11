package com.brackistar.gamemasternotes

import android.content.Context
import com.brackistar.gamemasternotes.core.ai.GroundedMvpAiEngine
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
    val aiEngine = GroundedMvpAiEngine()
}

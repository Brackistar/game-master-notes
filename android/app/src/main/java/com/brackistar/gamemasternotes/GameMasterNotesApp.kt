package com.brackistar.gamemasternotes

import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.brackistar.gamemasternotes.feature.assistant.AssistantScreen
import com.brackistar.gamemasternotes.feature.home.HomeScreen
import com.brackistar.gamemasternotes.feature.importpacks.ImportPacksScreen
import com.brackistar.gamemasternotes.feature.library.LibraryScreen

@Composable
fun GameMasterNotesApp(container: AppContainer) {
    val navController = rememberNavController()

    LaunchedEffect(Unit) {
        container.packFolderStore.selectedFolderUri()?.let { uri ->
            runCatching { container.packImporter.importFolder(uri) }
        }
    }

    Scaffold { paddingValues ->
        NavHost(
            navController = navController,
            startDestination = AppRoute.Home.path,
        ) {
            composable(AppRoute.Home.path) {
                HomeScreen(
                    paddingValues = paddingValues,
                    destinations = AppRoute.entries
                        .filterNot { it == AppRoute.Home }
                        .map { it.label to it.path },
                    onNavigate = navController::navigate,
                    repository = container.sourcebookRepository,
                )
            }
            composable(AppRoute.Library.path) {
                LibraryScreen(
                    paddingValues = paddingValues,
                    repository = container.sourcebookRepository,
                )
            }
            composable(AppRoute.Import.path) {
                ImportPacksScreen(
                    paddingValues = paddingValues,
                    folderStore = container.packFolderStore,
                    importer = container.packImporter,
                )
            }
            composable(AppRoute.Assistant.path) {
                AssistantScreen(
                    paddingValues = paddingValues,
                    repository = container.sourcebookRepository,
                    aiEngine = container.aiEngine,
                    modelFileInstaller = container.modelFileInstaller,
                )
            }
        }
    }
}

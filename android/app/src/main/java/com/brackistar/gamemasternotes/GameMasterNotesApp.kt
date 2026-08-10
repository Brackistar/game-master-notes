package com.brackistar.gamemasternotes

import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.brackistar.gamemasternotes.feature.assistant.AssistantScreen
import com.brackistar.gamemasternotes.feature.home.HomeScreen
import com.brackistar.gamemasternotes.feature.importpacks.ImportPacksScreen
import com.brackistar.gamemasternotes.feature.library.LibraryScreen
import com.brackistar.gamemasternotes.feature.search.SearchScreen
import com.brackistar.gamemasternotes.feature.session.SessionScreen
import com.brackistar.gamemasternotes.feature.settings.SettingsScreen

@Composable
fun GameMasterNotesApp() {
    val navController = rememberNavController()

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
                )
            }
            composable(AppRoute.Library.path) { LibraryScreen(paddingValues = paddingValues) }
            composable(AppRoute.Session.path) { SessionScreen(paddingValues = paddingValues) }
            composable(AppRoute.Import.path) { ImportPacksScreen(paddingValues = paddingValues) }
            composable(AppRoute.Search.path) { SearchScreen(paddingValues = paddingValues) }
            composable(AppRoute.Assistant.path) { AssistantScreen(paddingValues = paddingValues) }
            composable(AppRoute.Settings.path) { SettingsScreen(paddingValues = paddingValues) }
        }
    }
}

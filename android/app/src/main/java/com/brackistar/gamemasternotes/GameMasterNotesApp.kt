package com.brackistar.gamemasternotes

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.NavigationBarItemDefaults
import androidx.compose.material3.NavigationRail
import androidx.compose.material3.NavigationRailItem
import androidx.compose.material3.NavigationRailItemDefaults
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.navigation.NavDestination.Companion.hierarchy
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navOptions
import com.brackistar.gamemasternotes.feature.assistant.AssistantScreen
import com.brackistar.gamemasternotes.core.design.WayfinderTheme
import com.brackistar.gamemasternotes.feature.importpacks.ImportPacksScreen
import com.brackistar.gamemasternotes.feature.library.LibraryScreen
import com.brackistar.gamemasternotes.feature.settings.SettingsScreen

private val topLevelRoutes = listOf(AppRoute.Assistant, AppRoute.Library, AppRoute.Settings)

@Composable
fun GameMasterNotesApp(container: AppContainer) {
    val navController = rememberNavController()
    val backStackEntry by navController.currentBackStackEntryAsState()
    val currentDestination = backStackEntry?.destination

    LaunchedEffect(Unit) {
        container.packFolderStore.selectedFolderUri()?.let { uri ->
            runCatching { container.packImporter.importFolder(uri) }
        }
    }

    BoxWithConstraints(
        Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background),
    ) {
        val useRail = maxWidth >= 600.dp
        val wayfinderColors = WayfinderTheme.colors
        val navigateTopLevel: (AppRoute) -> Unit = { route ->
            navController.navigate(route.path, navOptions {
                popUpTo(AppRoute.Assistant.path) { saveState = true }
                launchSingleTop = true
                restoreState = true
            })
        }
        val content: @Composable () -> Unit = {
            NavHost(navController, AppRoute.Assistant.path, Modifier.fillMaxSize()) {
                composable(AppRoute.Assistant.path) {
                    AssistantScreen(
                        PaddingValues(0.dp), container.assistantRetrievalRepository,
                        container.aiEngine, container.modelFileInstaller,
                        container.diagnosticsJournal, container.diagnosticsArchiveWriter,
                    )
                }
                composable(AppRoute.Library.path) {
                    LibraryScreen(PaddingValues(0.dp), container.sourcebookRepository) {
                        navController.navigate(AppRoute.Import.path)
                    }
                }
                composable(AppRoute.Settings.path) { SettingsScreen(PaddingValues(0.dp)) }
                composable(AppRoute.Import.path) {
                    ImportPacksScreen(PaddingValues(0.dp), container.packFolderStore, container.packImporter)
                }
            }
        }
        if (useRail) {
            Row(Modifier.fillMaxSize()) {
                NavigationRail(
                    modifier = Modifier.width(112.dp),
                    containerColor = wayfinderColors.navigationContainer,
                    contentColor = wayfinderColors.onNavigationContainer,
                    header = {
                        Text(
                            text = "GMN",
                            modifier = Modifier.padding(vertical = 12.dp),
                            style = MaterialTheme.typography.titleMedium,
                        )
                    },
                ) {
                    topLevelRoutes.forEach { route ->
                        val selected = currentDestination?.hierarchy?.any { it.route == route.path } == true
                        NavigationRailItem(
                            selected = selected,
                            onClick = { navigateTopLevel(route) },
                            icon = { DestinationMark(route.label) },
                            label = { Text(route.label) },
                            colors = NavigationRailItemDefaults.colors(
                                selectedIconColor = wayfinderColors.onNavigationSelected,
                                selectedTextColor = wayfinderColors.onNavigationContainer,
                                indicatorColor = wayfinderColors.navigationSelected,
                                unselectedIconColor = wayfinderColors.onNavigationContainer,
                                unselectedTextColor = wayfinderColors.onNavigationContainer,
                            ),
                        )
                    }
                }
                Box(
                    Modifier
                        .weight(1f)
                        .fillMaxSize()
                        .background(MaterialTheme.colorScheme.background),
                ) { content() }
            }
        } else {
            Scaffold(bottomBar = {
                NavigationBar(containerColor = wayfinderColors.navigationContainer) {
                    topLevelRoutes.forEach { route ->
                        val selected = currentDestination?.hierarchy?.any { it.route == route.path } == true
                        NavigationBarItem(
                            selected = selected,
                            onClick = { navigateTopLevel(route) },
                            icon = { DestinationMark(route.label) },
                            label = { Text(route.label) },
                            colors = NavigationBarItemDefaults.colors(
                                selectedIconColor = wayfinderColors.onNavigationSelected,
                                selectedTextColor = wayfinderColors.onNavigationContainer,
                                indicatorColor = wayfinderColors.navigationSelected,
                                unselectedIconColor = wayfinderColors.onNavigationContainer,
                                unselectedTextColor = wayfinderColors.onNavigationContainer,
                            ),
                        )
                    }
                }
            }, containerColor = MaterialTheme.colorScheme.background) { innerPadding ->
                Box(
                    Modifier
                        .fillMaxSize()
                        .padding(innerPadding)
                        .background(MaterialTheme.colorScheme.background),
                ) { content() }
            }
        }
    }
}

@Composable
private fun DestinationMark(label: String) {
    Box(
        Modifier
            .padding(horizontal = 10.dp, vertical = 3.dp),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            label.take(1),
            style = MaterialTheme.typography.labelLarge,
        )
    }
}

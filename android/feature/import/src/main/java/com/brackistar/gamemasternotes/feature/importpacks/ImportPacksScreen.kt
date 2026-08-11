package com.brackistar.gamemasternotes.feature.importpacks

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.brackistar.gamemasternotes.core.importpacks.PackFolderStore
import com.brackistar.gamemasternotes.core.importpacks.PackImportSummary
import com.brackistar.gamemasternotes.core.importpacks.PackImporter
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

@Composable
fun ImportPacksScreen(
    paddingValues: PaddingValues,
    folderStore: PackFolderStore,
    importer: PackImporter,
) {
    val context = LocalContext.current
    val viewModel: ImportPacksViewModel = viewModel(
        factory = ImportPacksViewModel.factory(folderStore, importer),
    )
    val state by viewModel.state.collectAsState()
    val folderPicker = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.OpenDocumentTree(),
    ) { uri ->
        if (uri != null) {
            PackFolderStore.persistReadPermission(context, uri)
            viewModel.selectFolder(uri)
        }
    }

    LaunchedEffect(Unit) {
        viewModel.scanSelectedFolder()
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(paddingValues)
            .padding(24.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text(text = "Sourcebook Packs", style = MaterialTheme.typography.headlineSmall)
        Text(text = state.selectedFolder ?: "No pack folder selected.")
        Button(onClick = { folderPicker.launch(null) }) {
            Text(text = "Choose pack folder")
        }
        Button(
            onClick = viewModel::scanSelectedFolder,
            enabled = state.selectedFolder != null && !state.isScanning,
        ) {
            Text(text = if (state.isScanning) "Scanning..." else "Rescan folder")
        }
        state.summary?.let { summary ->
            Text(
                text = "Scanned ${summary.scannedCount}, imported ${summary.importedCount}, skipped ${summary.skippedUnchangedCount}, removed ${summary.removedCount}.",
            )
            summary.errors.forEach { error ->
                Text(text = error, color = MaterialTheme.colorScheme.error)
            }
        }
        state.message?.let { Text(text = it) }
    }
}

data class ImportPacksUiState(
    val selectedFolder: String? = null,
    val isScanning: Boolean = false,
    val summary: PackImportSummary? = null,
    val message: String? = null,
)

class ImportPacksViewModel(
    private val folderStore: PackFolderStore,
    private val importer: PackImporter,
) : ViewModel() {
    private val _state = MutableStateFlow(
        ImportPacksUiState(selectedFolder = folderStore.selectedFolderUri()?.toString()),
    )
    val state: StateFlow<ImportPacksUiState> = _state.asStateFlow()

    fun selectFolder(uri: Uri) {
        folderStore.saveSelectedFolder(uri)
        _state.update { it.copy(selectedFolder = uri.toString(), summary = null, message = null) }
        scanSelectedFolder()
    }

    fun scanSelectedFolder() {
        val uri = folderStore.selectedFolderUri()
        if (uri == null) {
            _state.update { it.copy(message = "Choose a folder that contains .gmnpack files.") }
            return
        }
        viewModelScope.launch {
            _state.update { it.copy(isScanning = true, message = null) }
            runCatching { importer.importFolder(uri) }
                .onSuccess { summary ->
                    _state.update {
                        it.copy(isScanning = false, summary = summary)
                    }
                }
                .onFailure { error ->
                    _state.update {
                        it.copy(isScanning = false, message = error.message ?: "Scan failed.")
                    }
                }
        }
    }

    companion object {
        fun factory(
            folderStore: PackFolderStore,
            importer: PackImporter,
        ): ViewModelProvider.Factory =
            object : ViewModelProvider.Factory {
                @Suppress("UNCHECKED_CAST")
                override fun <T : ViewModel> create(modelClass: Class<T>): T =
                    ImportPacksViewModel(folderStore, importer) as T
            }
    }
}

package com.brackistar.gamemasternotes.feature.library

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.width
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.brackistar.gamemasternotes.core.data.SourcebookPackSummary
import com.brackistar.gamemasternotes.core.data.SourcebookRepository
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn

@Composable
fun LibraryScreen(
    paddingValues: PaddingValues,
    repository: SourcebookRepository,
    onManageSources: () -> Unit = {},
) {
    val viewModel: LibraryViewModel = viewModel(factory = LibraryViewModel.factory(repository))
    val packs by viewModel.packs.collectAsState()

    var selectedPack by remember(packs) { mutableStateOf(packs.firstOrNull()) }
    BoxWithConstraints(
        modifier = Modifier
            .fillMaxSize()
            .padding(paddingValues)
    ) {
        val expanded = maxWidth >= 840.dp
        Column(Modifier.fillMaxSize().padding(if (expanded) 32.dp else 16.dp)) {
            Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Column {
                    Text(text = "Library", style = MaterialTheme.typography.headlineMedium)
                    Text("${packs.size} indexed sourcebook ${if (packs.size == 1) "pack" else "packs"}")
                }
                Button(onClick = onManageSources) { Text("Manage sources") }
            }
            HorizontalDivider(Modifier.padding(vertical = 16.dp))
            if (packs.isEmpty()) {
                Card(Modifier.fillMaxWidth()) {
                    Column(Modifier.padding(24.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                        Text("Add sourcebooks to begin", style = MaterialTheme.typography.titleLarge)
                        Text("Choose the folder that contains your .gmnpack files. The library stays on this device.")
                        Button(onClick = onManageSources) { Text("Choose source folder") }
                    }
                }
            } else if (expanded) {
                Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.spacedBy(24.dp)) {
                    PackList(packs, selectedPack, { selectedPack = it }, Modifier.width(340.dp))
                    PackDetail(selectedPack, Modifier.weight(1f))
                }
            } else {
                PackList(packs, selectedPack, { selectedPack = it }, Modifier.fillMaxWidth())
            }
        }
    }
}

@Composable
private fun PackList(
    packs: List<SourcebookPackSummary>,
    selected: SourcebookPackSummary?,
    onSelect: (SourcebookPackSummary) -> Unit,
    modifier: Modifier,
) {
    Column(modifier, verticalArrangement = Arrangement.spacedBy(8.dp)) {
        packs.forEach { pack ->
            Card(Modifier.fillMaxWidth().clickable { onSelect(pack) }) {
                Column(Modifier.padding(16.dp)) {
                    Text(pack.title, style = MaterialTheme.typography.titleMedium)
                    Text("${pack.system} · ${pack.edition}", style = MaterialTheme.typography.bodyMedium)
                    Text("${pack.chunkCount} indexed passages${if (selected == pack) " · Selected" else ""}", style = MaterialTheme.typography.labelMedium)
                }
            }
        }
    }
}

@Composable
private fun PackDetail(pack: SourcebookPackSummary?, modifier: Modifier) {
    Card(modifier.fillMaxSize()) {
        Column(Modifier.padding(24.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Text(pack?.title ?: "Select a sourcebook", style = MaterialTheme.typography.titleLarge)
            pack?.let {
                Text("Indexed and ready", color = MaterialTheme.colorScheme.primary, style = MaterialTheme.typography.labelLarge)
                Text("System: ${it.system}")
                Text("Edition: ${it.edition}")
                Text("Source: ${it.sourceDisplayName}")
                Text("${it.chunkCount} searchable passages")
            }
        }
    }
}

class LibraryViewModel(repository: SourcebookRepository) : ViewModel() {
    val packs: StateFlow<List<SourcebookPackSummary>> = repository.observePacks()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5_000), emptyList())

    companion object {
        fun factory(repository: SourcebookRepository): ViewModelProvider.Factory =
            object : ViewModelProvider.Factory {
                @Suppress("UNCHECKED_CAST")
                override fun <T : ViewModel> create(modelClass: Class<T>): T =
                    LibraryViewModel(repository) as T
            }
    }
}

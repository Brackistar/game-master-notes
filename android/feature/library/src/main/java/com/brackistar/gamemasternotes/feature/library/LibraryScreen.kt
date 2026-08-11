package com.brackistar.gamemasternotes.feature.library

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
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
) {
    val viewModel: LibraryViewModel = viewModel(factory = LibraryViewModel.factory(repository))
    val packs by viewModel.packs.collectAsState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(paddingValues)
            .padding(24.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text(text = "Library", style = MaterialTheme.typography.headlineSmall)
        if (packs.isEmpty()) {
            Text(text = "No sourcebook packs indexed yet.")
        } else {
            packs.forEach { pack ->
                Text(
                    text = "${pack.title} (${pack.system} ${pack.edition}) - ${pack.chunkCount} chunks",
                    style = MaterialTheme.typography.bodyLarge,
                )
                Text(text = pack.sourceDisplayName, style = MaterialTheme.typography.bodySmall)
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

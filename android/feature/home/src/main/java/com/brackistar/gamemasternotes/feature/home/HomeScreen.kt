package com.brackistar.gamemasternotes.feature.home

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewmodel.compose.viewModel
import com.brackistar.gamemasternotes.core.data.SourcebookRepository
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import androidx.lifecycle.viewModelScope

@Composable
fun HomeScreen(
    paddingValues: PaddingValues,
    destinations: List<Pair<String, String>>,
    onNavigate: (String) -> Unit,
    repository: SourcebookRepository,
) {
    val viewModel: HomeViewModel = viewModel(
        factory = HomeViewModel.factory(repository),
    )
    val packCount by viewModel.packCount.collectAsState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(paddingValues)
            .padding(24.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text(text = "Game Master Notes", style = MaterialTheme.typography.headlineSmall)
        Text(text = "$packCount sourcebook pack${if (packCount == 1) "" else "s"} indexed")
        destinations.forEach { (label, path) ->
            Button(onClick = { onNavigate(path) }) {
                Text(text = label)
            }
        }
    }
}

class HomeViewModel(repository: SourcebookRepository) : ViewModel() {
    val packCount: StateFlow<Int> = repository.observePackCount()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5_000), 0)

    companion object {
        fun factory(repository: SourcebookRepository): ViewModelProvider.Factory =
            object : ViewModelProvider.Factory {
                @Suppress("UNCHECKED_CAST")
                override fun <T : ViewModel> create(modelClass: Class<T>): T =
                    HomeViewModel(repository) as T
            }
    }
}

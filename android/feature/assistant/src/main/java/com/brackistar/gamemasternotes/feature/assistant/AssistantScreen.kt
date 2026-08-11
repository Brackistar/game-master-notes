package com.brackistar.gamemasternotes.feature.assistant

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
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
import com.brackistar.gamemasternotes.core.ai.AiEngine
import com.brackistar.gamemasternotes.core.ai.AiRequest
import com.brackistar.gamemasternotes.core.data.SourcebookRepository
import com.brackistar.gamemasternotes.core.retrieval.RetrievalQuery
import com.brackistar.gamemasternotes.core.retrieval.RetrievalResult
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

@Composable
fun AssistantScreen(
    paddingValues: PaddingValues,
    repository: SourcebookRepository,
    aiEngine: AiEngine,
) {
    val viewModel: AssistantViewModel = viewModel(
        factory = AssistantViewModel.factory(repository, aiEngine),
    )
    val state by viewModel.state.collectAsState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(paddingValues)
            .padding(24.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text(text = "Ask the Books", style = MaterialTheme.typography.headlineSmall)
        Column(
            modifier = Modifier
                .weight(1f)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            if (state.messages.isEmpty()) {
                Text(text = "Ask a question and I will answer only from indexed sourcebook chunks.")
            }
            state.messages.forEach { message ->
                Text(
                    text = if (message.isUser) "You: ${message.text}" else "Books: ${message.text}",
                    style = if (message.isUser) MaterialTheme.typography.bodyLarge else MaterialTheme.typography.bodyMedium,
                )
                if (!message.isUser) {
                    message.citations.forEach { citation ->
                        Text(
                            text = citation.citationLabel ?: citation.title,
                            style = MaterialTheme.typography.labelMedium,
                        )
                    }
                }
            }
            state.error?.let { Text(text = it, color = MaterialTheme.colorScheme.error) }
        }
        OutlinedTextField(
            value = state.question,
            onValueChange = viewModel::updateQuestion,
            modifier = Modifier.fillMaxWidth(),
            label = { Text(text = "Question") },
            enabled = !state.isGenerating,
        )
        Button(
            onClick = viewModel::ask,
            enabled = state.question.isNotBlank() && !state.isGenerating,
        ) {
            Text(text = if (state.isGenerating) "Asking..." else "Ask")
        }
    }
}

data class ChatMessage(
    val text: String,
    val isUser: Boolean,
    val citations: List<RetrievalResult> = emptyList(),
)

data class AssistantUiState(
    val question: String = "",
    val isGenerating: Boolean = false,
    val messages: List<ChatMessage> = emptyList(),
    val error: String? = null,
)

class AssistantViewModel(
    private val repository: SourcebookRepository,
    private val aiEngine: AiEngine,
) : ViewModel() {
    private val _state = MutableStateFlow(AssistantUiState())
    val state: StateFlow<AssistantUiState> = _state.asStateFlow()

    fun updateQuestion(question: String) {
        _state.update { it.copy(question = question) }
    }

    fun ask() {
        val question = state.value.question.trim()
        if (question.isBlank()) return

        viewModelScope.launch {
            _state.update {
                it.copy(
                    question = "",
                    isGenerating = true,
                    error = null,
                    messages = it.messages + ChatMessage(question, isUser = true),
                )
            }

            runCatching {
                val results = repository.search(RetrievalQuery(text = question))
                val context = results.joinToString("\n\n") { result ->
                    "[${result.citationLabel ?: result.sourceId}]\n${result.snippet}"
                }
                val response = aiEngine.generate(AiRequest(prompt = question, context = context))
                ChatMessage(response.text, isUser = false, citations = results)
            }.onSuccess { message ->
                _state.update {
                    it.copy(isGenerating = false, messages = it.messages + message)
                }
            }.onFailure { error ->
                _state.update {
                    it.copy(isGenerating = false, error = error.message ?: "Could not ask the books.")
                }
            }
        }
    }

    companion object {
        fun factory(
            repository: SourcebookRepository,
            aiEngine: AiEngine,
        ): ViewModelProvider.Factory =
            object : ViewModelProvider.Factory {
                @Suppress("UNCHECKED_CAST")
                override fun <T : ViewModel> create(modelClass: Class<T>): T =
                    AssistantViewModel(repository, aiEngine) as T
            }
    }
}

package com.brackistar.gamemasternotes.feature.assistant

import android.net.Uri
import android.util.Log
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.MenuAnchorType
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.brackistar.gamemasternotes.core.ai.AiEngine
import com.brackistar.gamemasternotes.core.ai.AiModelAvailability
import com.brackistar.gamemasternotes.core.ai.AiModel
import com.brackistar.gamemasternotes.core.ai.AiRequest
import com.brackistar.gamemasternotes.core.ai.AiResponse
import com.brackistar.gamemasternotes.core.ai.EvidenceBriefBuilder
import com.brackistar.gamemasternotes.core.ai.ModelFileInstaller
import com.brackistar.gamemasternotes.core.retrieval.RetrievalQuery
import com.brackistar.gamemasternotes.core.retrieval.RetrievalRepository
import com.brackistar.gamemasternotes.core.retrieval.RetrievalResult
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeoutOrNull

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AssistantScreen(
    paddingValues: PaddingValues,
    repository: RetrievalRepository,
    aiEngine: AiEngine,
    modelFileInstaller: ModelFileInstaller,
) {
    val viewModel: AssistantViewModel = viewModel(
        factory = AssistantViewModel.factory(repository, aiEngine, modelFileInstaller),
    )
    val state by viewModel.state.collectAsStateWithLifecycle()
    val modelPicker = rememberLauncherForActivityResult(ActivityResultContracts.OpenDocument()) { uri ->
        viewModel.importPendingModel(uri)
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(paddingValues)
            .padding(24.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text(text = "Ask the Books", style = MaterialTheme.typography.headlineSmall)
        ModelSelector(
            models = state.availableModels,
            selectedModelId = state.selectedModelId,
            isEnabled = !state.isGenerating,
            onSelectModel = viewModel::selectModel,
        )
        Button(
            onClick = {
                viewModel.preparePrimaryModelImport()
                modelPicker.launch(arrayOf("application/octet-stream", "*/*"))
            },
            enabled = !state.isGenerating,
        ) {
            Text(
                text = if (state.availableModels.any { !it.isFallback }) {
                    "Replace LFM2.5 350M Q2 GGUF"
                } else {
                    "Import LFM2.5 350M Q2 GGUF"
                },
            )
        }
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

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun ModelSelector(
    models: List<AiModel>,
    selectedModelId: String?,
    isEnabled: Boolean,
    onSelectModel: (String) -> Unit,
) {
    val selectedModel = models.firstOrNull { it.id == selectedModelId } ?: models.firstOrNull()
    var expanded by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(false) }

    ExposedDropdownMenuBox(
        expanded = expanded,
        onExpandedChange = { if (isEnabled) expanded = !expanded },
        modifier = Modifier.fillMaxWidth(),
    ) {
        OutlinedTextField(
            value = selectedModel?.displayName ?: "No compatible model",
            onValueChange = {},
            readOnly = true,
            enabled = isEnabled && models.isNotEmpty(),
            modifier = Modifier
                .menuAnchor(MenuAnchorType.PrimaryNotEditable, enabled = isEnabled)
                .fillMaxWidth(),
            label = { Text(text = "Answer model") },
            supportingText = {
                selectedModel?.let { model ->
                    Text(text = model.description.ifBlank { model.quantization ?: model.id })
                }
            },
            trailingIcon = {
                ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded)
            },
        )
        ExposedDropdownMenu(
            expanded = expanded,
            onDismissRequest = { expanded = false },
        ) {
            models.forEach { model ->
                DropdownMenuItem(
                    text = {
                        Column {
                            Text(text = model.displayName)
                            Text(
                                text = model.statusLabel(),
                                style = MaterialTheme.typography.labelSmall,
                            )
                        }
                    },
                    onClick = {
                        expanded = false
                        onSelectModel(model.id)
                    },
                )
            }
        }
    }
}

private fun AiModel.statusLabel(): String =
    when (availability) {
        AiModelAvailability.Ready ->
            quantization?.let { "${minimumRamMb} MB RAM minimum, $it" } ?: description
        AiModelAvailability.MissingModelFile ->
            "Not installed"
        AiModelAvailability.UnsupportedDevice ->
            "Unsupported device"
    }

data class ChatMessage(
    val text: String,
    val isUser: Boolean,
    val citations: List<RetrievalResult> = emptyList(),
)

data class AssistantUiState(
    val question: String = "",
    val isGenerating: Boolean = false,
    val availableModels: List<AiModel> = emptyList(),
    val selectedModelId: String? = null,
    val pendingImportModelId: String? = null,
    val messages: List<ChatMessage> = emptyList(),
    val error: String? = null,
)

class AssistantViewModel(
    private val repository: RetrievalRepository,
    private val aiEngine: AiEngine,
    private val modelFileInstaller: ModelFileInstaller,
) : ViewModel() {
    private val _state = MutableStateFlow(AssistantUiState())
    val state: StateFlow<AssistantUiState> = _state.asStateFlow()

    init {
        viewModelScope.launch {
            refreshModels(loadDefault = false)
        }
    }

    fun updateQuestion(question: String) {
        _state.update { it.copy(question = question) }
    }

    fun selectModel(modelId: String) {
        if (modelId == state.value.selectedModelId) return
        val model = state.value.availableModels.firstOrNull { it.id == modelId }
        if (model?.availability != AiModelAvailability.Ready) {
            Log.w(TAG, "Model select rejected modelId=$modelId availability=${model?.availability}")
            _state.update {
                it.copy(error = "${model?.displayName ?: modelId} is not ready: ${model?.availability}.")
            }
            return
        }

        viewModelScope.launch {
            val startedAt = System.currentTimeMillis()
            Log.i(TAG, "Loading selected model modelId=$modelId displayName=${model.displayName}")
            runCatching {
                aiEngine.load(modelId)
            }.onSuccess {
                Log.i(TAG, "Loaded selected model modelId=$modelId elapsedMs=${System.currentTimeMillis() - startedAt}")
                _state.update { state ->
                    state.copy(selectedModelId = modelId, error = null)
                }
            }.onFailure { error ->
                Log.e(TAG, "Could not load selected model modelId=$modelId elapsedMs=${System.currentTimeMillis() - startedAt}", error)
                _state.update { state ->
                    state.copy(error = error.message ?: "Could not load selected model.")
                }
            }
        }
    }

    fun prepareModelImport(modelId: String) {
        _state.update {
            it.copy(
                pendingImportModelId = modelId,
                error = "Select ${modelFileInstaller.expectedFileName(modelId) ?: "the GGUF model file"}.",
            )
        }
    }

    fun preparePrimaryModelImport() {
        prepareModelImport(modelFileInstaller.primaryModelId())
    }

    fun importPendingModel(uri: Uri?) {
        val modelId = state.value.pendingImportModelId
        if (uri == null || modelId == null) {
            _state.update { it.copy(pendingImportModelId = null) }
            return
        }

        viewModelScope.launch {
            val startedAt = System.currentTimeMillis()
            Log.i(TAG, "Importing model file modelId=$modelId uriScheme=${uri.scheme}")
            _state.update { it.copy(error = "Importing model file...") }
            runCatching {
                modelFileInstaller.importModel(modelId, uri)
            }.onSuccess { installed ->
                Log.i(
                    TAG,
                    "Imported model file modelId=${installed.modelId} fileName=${installed.fileName} elapsedMs=${System.currentTimeMillis() - startedAt}",
                )
                refreshModels(loadDefault = false)
                _state.update {
                    it.copy(
                        pendingImportModelId = null,
                        error = "Installed ${installed.fileName}.",
                    )
                }
                selectModel(installed.modelId)
            }.onFailure { error ->
                Log.e(TAG, "Could not import model file modelId=$modelId elapsedMs=${System.currentTimeMillis() - startedAt}", error)
                _state.update {
                    it.copy(
                        pendingImportModelId = null,
                        error = error.message ?: "Could not import model file.",
                    )
                }
            }
        }
    }

    fun ask() {
        val question = state.value.question.trim()
        if (question.isBlank()) return

        viewModelScope.launch {
            val askStartedAt = System.currentTimeMillis()
            val selectedModelId = state.value.selectedModelId ?: "none"
            val questionHash = question.hashCode()
            Log.i(TAG, "Ask started questionHash=$questionHash modelId=$selectedModelId questionChars=${question.length}")
            _state.update {
                it.copy(
                    question = "",
                    isGenerating = true,
                    error = null,
                    messages = it.messages + ChatMessage(question, isUser = true),
                )
            }

            runCatching {
                val retrievalStartedAt = System.currentTimeMillis()
                val results = repository.search(RetrievalQuery(text = question, limit = ASSISTANT_RETRIEVAL_LIMIT))
                Log.i(
                    TAG,
                    "Retrieval finished questionHash=$questionHash resultCount=${results.size} elapsedMs=${System.currentTimeMillis() - retrievalStartedAt} citations=${results.citationSummary()}",
                )
                val context = results.joinToString("\n\n") { result ->
                    "[${result.citationLabel ?: result.sourceId}]\n${result.snippet}"
                }
                val evidenceBrief = EvidenceBriefBuilder.build(question, context)
                Log.i(
                    TAG,
                    "Evidence brief built questionHash=$questionHash contextChars=${context.length} evidenceChars=${evidenceBrief.text.length} citationCount=${evidenceBrief.citationIds.size}",
                )
                val generationStartedAt = System.currentTimeMillis()
                Log.i(
                    TAG,
                    "AI generation started questionHash=$questionHash modelId=$selectedModelId timeoutMs=$ASK_TIMEOUT_MILLIS",
                )
                val response = withTimeoutOrNull(ASK_TIMEOUT_MILLIS) {
                    state.value.selectedModelId?.let { aiEngine.load(it) }
                    aiEngine.generate(AiRequest(prompt = question, context = evidenceBrief.text))
                } ?: run {
                    Log.w(
                        TAG,
                        "AI generation timed out questionHash=$questionHash modelId=$selectedModelId elapsedMs=${System.currentTimeMillis() - generationStartedAt}",
                    )
                    aiEngine.cancel()
                    AiResponse(
                        text = "The local model took too long to answer. Try a narrower question or switch to the grounded fallback model.",
                        citationIds = evidenceBrief.citationIds,
                    )
                }
                Log.i(
                    TAG,
                    "AI generation finished questionHash=$questionHash modelId=$selectedModelId outputChars=${response.text.length} responseCitationCount=${response.citationIds.size} elapsedMs=${System.currentTimeMillis() - generationStartedAt}",
                )
                ChatMessage(response.text, isUser = false, citations = results)
            }.onSuccess { message ->
                Log.i(
                    TAG,
                    "Ask finished questionHash=$questionHash modelId=$selectedModelId totalElapsedMs=${System.currentTimeMillis() - askStartedAt}",
                )
                _state.update {
                    it.copy(isGenerating = false, messages = it.messages + message)
                }
            }.onFailure { error ->
                Log.e(
                    TAG,
                    "Ask failed questionHash=$questionHash modelId=$selectedModelId totalElapsedMs=${System.currentTimeMillis() - askStartedAt}",
                    error,
                )
                _state.update {
                    it.copy(isGenerating = false, error = error.message ?: "Could not ask the books.")
                }
            }
        }
    }

    companion object {
        private const val ASK_TIMEOUT_MILLIS = 25_000L
        private const val ASSISTANT_RETRIEVAL_LIMIT = 2
        private const val TAG = "GmnAssistant"

        fun factory(
            repository: RetrievalRepository,
            aiEngine: AiEngine,
            modelFileInstaller: ModelFileInstaller,
        ): ViewModelProvider.Factory =
            object : ViewModelProvider.Factory {
                @Suppress("UNCHECKED_CAST")
                override fun <T : ViewModel> create(modelClass: Class<T>): T =
                    AssistantViewModel(repository, aiEngine, modelFileInstaller) as T
            }
    }

    private suspend fun refreshModels(loadDefault: Boolean) {
        val startedAt = System.currentTimeMillis()
        val models = aiEngine.availableModels()
        val currentSelectedId = state.value.selectedModelId
        val selectedModel = models.firstOrNull { it.id == currentSelectedId && it.availability == AiModelAvailability.Ready }
            ?: models.firstOrNull { it.availability == AiModelAvailability.Ready && !it.isFallback }
            ?: models.firstOrNull { it.availability == AiModelAvailability.Ready }
        if (loadDefault && selectedModel != null) {
            aiEngine.load(selectedModel.id)
        }
        Log.i(
            TAG,
            "Model list refreshed count=${models.size} local=${models.count { !it.isFallback }} selected=${selectedModel?.id} elapsedMs=${System.currentTimeMillis() - startedAt}",
        )
        _state.update {
            it.copy(
                availableModels = models,
                selectedModelId = selectedModel?.id,
            )
        }
    }
}

private fun List<RetrievalResult>.citationSummary(): String =
    take(MAX_LOGGED_CITATIONS)
        .joinToString(separator = " | ") { it.citationLabel ?: it.sourceId }
        .ifBlank { "none" }

private const val MAX_LOGGED_CITATIONS = 5

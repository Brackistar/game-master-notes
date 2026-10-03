package com.brackistar.gamemasternotes.feature.assistant

import android.net.Uri
import android.util.Log
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MenuAnchorType
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.Surface
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.brackistar.gamemasternotes.core.ai.AiEngine
import com.brackistar.gamemasternotes.core.ai.AiAnswerProvenance
import com.brackistar.gamemasternotes.core.ai.AiEvidence
import com.brackistar.gamemasternotes.core.ai.AnswerMode
import com.brackistar.gamemasternotes.core.ai.AiModelAvailability
import com.brackistar.gamemasternotes.core.ai.AiModel
import com.brackistar.gamemasternotes.core.ai.AiRequest
import com.brackistar.gamemasternotes.core.ai.AiResponse
import com.brackistar.gamemasternotes.core.ai.AiResponseMode
import com.brackistar.gamemasternotes.core.ai.EvidenceBriefBuilder
import com.brackistar.gamemasternotes.core.ai.ModelFileInstaller
import com.brackistar.gamemasternotes.core.domain.diagnostics.DiagnosticEvent
import com.brackistar.gamemasternotes.core.domain.diagnostics.DiagnosticsJournal
import com.brackistar.gamemasternotes.core.retrieval.RetrievalQuery
import com.brackistar.gamemasternotes.core.retrieval.RetrievalRepository
import com.brackistar.gamemasternotes.core.retrieval.RetrievalResult
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeoutOrNull
import java.security.MessageDigest
import java.util.UUID

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AssistantScreen(
    paddingValues: PaddingValues,
    repository: RetrievalRepository,
    aiEngine: AiEngine,
    modelFileInstaller: ModelFileInstaller,
    diagnosticsJournal: DiagnosticsJournal,
    diagnosticsArchiveWriter: DiagnosticsArchiveWriter,
) {
    val viewModel: AssistantViewModel = viewModel(
        factory = AssistantViewModel.factory(
            repository,
            aiEngine,
            modelFileInstaller,
            diagnosticsJournal,
            diagnosticsArchiveWriter,
        ),
    )
    val state by viewModel.state.collectAsStateWithLifecycle()
    val modelPicker = rememberLauncherForActivityResult(ActivityResultContracts.OpenDocument()) { uri ->
        viewModel.importPendingModel(uri)
    }
    val diagnosticsExporter = rememberLauncherForActivityResult(
        ActivityResultContracts.CreateDocument("application/zip"),
    ) { uri ->
        viewModel.exportDiagnostics(uri)
    }

    BoxWithConstraints(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
            .padding(paddingValues)
    ) {
        val outerPadding = when {
            maxWidth >= 840.dp -> 32.dp
            maxWidth >= 600.dp -> 24.dp
            else -> 16.dp
        }
        val useHorizontalControls = maxWidth >= 600.dp

        Column(
            modifier = Modifier
                .widthIn(max = 720.dp)
                .fillMaxWidth()
                .fillMaxHeight()
                .align(Alignment.TopCenter)
                .padding(horizontal = outerPadding, vertical = 20.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.Top,
            ) {
                Column(
                    modifier = Modifier.weight(1f),
                    verticalArrangement = Arrangement.spacedBy(4.dp),
                ) {
                    Text(text = "Ask the Books", style = MaterialTheme.typography.headlineMedium)
                    Text(
                        text = "Rules and lore, grounded in your indexed sourcebooks.",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
                Surface(
                    color = MaterialTheme.colorScheme.primaryContainer,
                    shape = RoundedCornerShape(999.dp),
                ) {
                    Text(
                        text = "Offline · ready",
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 7.dp),
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.onPrimaryContainer,
                    )
                }
            }

            if (useHorizontalControls) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    AnswerModeSelector(
                        mode = state.answerMode,
                        isEnabled = !state.isGenerating,
                        onSelectMode = viewModel::selectAnswerMode,
                        modifier = Modifier.weight(1f),
                        showSupportingText = false,
                    )
                    ModelSelector(
                        models = state.availableModels,
                        selectedModelId = state.selectedModelId,
                        isEnabled = !state.isGenerating,
                        onSelectModel = viewModel::selectModel,
                        modifier = Modifier.weight(1.4f),
                        showSupportingText = false,
                    )
                }
            } else {
                Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    AnswerModeSelector(
                        mode = state.answerMode,
                        isEnabled = !state.isGenerating,
                        onSelectMode = viewModel::selectAnswerMode,
                        showSupportingText = false,
                    )
                    ModelSelector(
                        models = state.availableModels,
                        selectedModelId = state.selectedModelId,
                        isEnabled = !state.isGenerating,
                        onSelectModel = viewModel::selectModel,
                        showSupportingText = false,
                    )
                }
            }

            if (useHorizontalControls) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    TextButton(
                        onClick = {
                            viewModel.preparePrimaryModelImport()
                            modelPicker.launch(arrayOf("application/octet-stream", "*/*"))
                        },
                        enabled = !state.isGenerating,
                    ) {
                        Text(if (state.availableModels.any { !it.isFallback }) "Replace model file" else "Import model file")
                    }
                    TextButton(onClick = { diagnosticsExporter.launch(DIAGNOSTICS_EXPORT_NAME) }) {
                        Text(text = "Export diagnostics")
                    }
                    TextButton(onClick = viewModel::clearDiagnostics) {
                        Text(text = "Clear diagnostics")
                    }
                }
            } else {
                Column {
                    TextButton(
                        onClick = {
                            viewModel.preparePrimaryModelImport()
                            modelPicker.launch(arrayOf("application/octet-stream", "*/*"))
                        },
                        enabled = !state.isGenerating,
                    ) {
                        Text(if (state.availableModels.any { !it.isFallback }) "Replace model file" else "Import model file")
                    }
                    Row {
                        TextButton(onClick = { diagnosticsExporter.launch(DIAGNOSTICS_EXPORT_NAME) }) {
                            Text(text = "Export diagnostics")
                        }
                        TextButton(onClick = viewModel::clearDiagnostics) {
                            Text(text = "Clear diagnostics")
                        }
                    }
                }
            }

            HorizontalDivider()
            Column(
                modifier = Modifier
                    .weight(1f)
                    .verticalScroll(rememberScrollState()),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                if (state.messages.isEmpty()) {
                    Surface(
                        modifier = Modifier.fillMaxWidth(),
                        color = MaterialTheme.colorScheme.surfaceVariant,
                        shape = RoundedCornerShape(14.dp),
                    ) {
                        Column(
                            Modifier.padding(20.dp),
                            verticalArrangement = Arrangement.spacedBy(8.dp),
                        ) {
                            Text("What do you need at the table?", style = MaterialTheme.typography.titleLarge)
                            Text(
                                "Choose an example or write your own question. Examples fill the draft without submitting.",
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                            )
                            TextButton(onClick = { viewModel.updateQuestion("How does hidden movement work?") }) {
                                Text("How does hidden movement work?")
                            }
                            TextButton(onClick = { viewModel.updateQuestion("What is known about the Ashen Court?") }) {
                                Text("What is known about the Ashen Court?")
                            }
                        }
                    }
                }
                state.messages.forEach { message ->
                    Surface(
                        modifier = Modifier.fillMaxWidth(),
                        color = if (message.isUser) MaterialTheme.colorScheme.primaryContainer else MaterialTheme.colorScheme.surface,
                        shape = RoundedCornerShape(10.dp),
                        tonalElevation = if (message.isUser) 0.dp else 1.dp,
                    ) {
                        Column(Modifier.padding(16.dp)) {
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.SpaceBetween,
                                verticalAlignment = Alignment.CenterVertically,
                            ) {
                                Text(if (message.isUser) "Your question" else "From the books", style = MaterialTheme.typography.labelLarge)
                                if (!message.isUser && message.answerProvenance != null) {
                                    AnswerProvenanceBadge(message.answerProvenance)
                                }
                            }
                            Text(message.text, style = MaterialTheme.typography.bodyLarge)
                        }
                    }
                    if (!message.isUser) {
                        if (message.citations.isNotEmpty()) {
                            Text(text = "Cited sources", style = MaterialTheme.typography.labelLarge)
                            SourceCards(message.requestId, message.citations)
                        }
                        if (message.retrievedEvidence.isNotEmpty()) {
                            Text(text = "Retrieved evidence", style = MaterialTheme.typography.labelLarge)
                            SourceCards(message.requestId, message.retrievedEvidence)
                        }
                    }
                }
                if (state.isGenerating) {
                    Text(text = "Searching the loaded books and composing an answer...")
                    TextButton(onClick = viewModel::cancel) { Text(text = "Cancel") }
                }
                state.error?.let { Text(text = it, color = MaterialTheme.colorScheme.error) }
                state.notice?.let { Text(text = it, color = MaterialTheme.colorScheme.primary) }
            }

            Surface(
                modifier = Modifier.fillMaxWidth(),
                color = MaterialTheme.colorScheme.surface,
                shape = RoundedCornerShape(14.dp),
                tonalElevation = 1.dp,
            ) {
                Column(
                    modifier = Modifier.padding(12.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    OutlinedTextField(
                        value = state.question,
                        onValueChange = viewModel::updateQuestion,
                        modifier = Modifier
                            .fillMaxWidth()
                            .heightIn(min = 72.dp, max = 136.dp),
                        label = { Text(text = "Question") },
                        placeholder = { Text("Ask a rules or lore question…") },
                        enabled = !state.isGenerating,
                        minLines = 2,
                        maxLines = 5,
                    )
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.End,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Button(
                            onClick = viewModel::ask,
                            enabled = state.question.isNotBlank() && !state.isGenerating,
                        ) {
                            Text(text = if (state.isGenerating) "Asking…" else "Ask")
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun SourceCards(requestId: String?, sources: List<RetrievalResult>) {
    sources.distinctBy { it.sourceId }.forEach { source ->
        var expanded by androidx.compose.runtime.remember(requestId, source.sourceId) {
            androidx.compose.runtime.mutableStateOf(false)
        }
        Card(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(10.dp)) {
                TextButton(onClick = { expanded = !expanded }) {
                    Text(text = if (expanded) "Hide source" else "Show source: ${source.citationLabel}")
                }
                if (expanded) {
                    Text(text = source.snippet, style = MaterialTheme.typography.bodySmall)
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun AnswerModeSelector(
    mode: AnswerMode,
    isEnabled: Boolean,
    onSelectMode: (AnswerMode) -> Unit,
    modifier: Modifier = Modifier,
    showSupportingText: Boolean = true,
) {
    var expanded by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(false) }
    ExposedDropdownMenuBox(
        expanded = expanded,
        onExpandedChange = { if (isEnabled) expanded = !expanded },
        modifier = modifier,
    ) {
        OutlinedTextField(
            value = mode.displayName,
            onValueChange = {},
            readOnly = true,
            enabled = isEnabled,
            modifier = Modifier.menuAnchor(MenuAnchorType.PrimaryNotEditable, enabled = isEnabled).fillMaxWidth(),
            label = { Text(text = "Answer style") },
            supportingText = if (showSupportingText) {
                { Text(text = mode.description) }
            } else {
                null
            },
            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded) },
        )
        ExposedDropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
            AnswerMode.entries.forEach { option ->
                DropdownMenuItem(
                    text = {
                        Column {
                            Text(text = option.displayName)
                            Text(text = option.description, style = MaterialTheme.typography.labelSmall)
                        }
                    },
                    onClick = {
                        expanded = false
                        onSelectMode(option)
                    },
                )
            }
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
    modifier: Modifier = Modifier,
    showSupportingText: Boolean = true,
) {
    val selectedModel = models.firstOrNull { it.id == selectedModelId } ?: models.firstOrNull()
    var expanded by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(false) }

    ExposedDropdownMenuBox(
        expanded = expanded,
        onExpandedChange = { if (isEnabled) expanded = !expanded },
        modifier = modifier,
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
            supportingText = if (showSupportingText) {
                {
                    selectedModel?.let { model ->
                        Text(text = model.description.ifBlank { model.quantization ?: model.id })
                    }
                }
            } else {
                null
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

@Composable
private fun AnswerProvenanceBadge(provenance: AiAnswerProvenance) {
    val (label, containerColor, contentColor) = when (provenance.mode) {
        AiResponseMode.DeterministicFallback -> Triple(
            "Deterministic lookup (no AI model loaded)",
            MaterialTheme.colorScheme.tertiaryContainer,
            MaterialTheme.colorScheme.onTertiaryContainer,
        )
        AiResponseMode.LocalModel -> Triple(
            provenance.modelDisplayName?.let { "Local AI model: $it" } ?: "Local AI model",
            MaterialTheme.colorScheme.secondaryContainer,
            MaterialTheme.colorScheme.onSecondaryContainer,
        )
        AiResponseMode.SystemMessage -> Triple(
            provenance.modelDisplayName ?: "Assistant status",
            MaterialTheme.colorScheme.surfaceVariant,
            MaterialTheme.colorScheme.onSurfaceVariant,
        )
    }
    Surface(
        color = containerColor,
        contentColor = contentColor,
        shape = RoundedCornerShape(999.dp),
    ) {
        Text(
            text = label,
            modifier = Modifier.padding(horizontal = 10.dp, vertical = 4.dp),
            style = MaterialTheme.typography.labelSmall,
        )
    }
}

data class ChatMessage(
    val text: String,
    val isUser: Boolean,
    val requestId: String? = null,
    val citations: List<RetrievalResult> = emptyList(),
    val retrievedEvidence: List<RetrievalResult> = emptyList(),
    val answerProvenance: AiAnswerProvenance? = null,
)

data class AssistantUiState(
    val question: String = "",
    val isGenerating: Boolean = false,
    val availableModels: List<AiModel> = emptyList(),
    val selectedModelId: String? = null,
    val pendingImportModelId: String? = null,
    val messages: List<ChatMessage> = emptyList(),
    val error: String? = null,
    val notice: String? = null,
    val answerMode: AnswerMode = AnswerMode.Explain,
)

class AssistantViewModel(
    private val repository: RetrievalRepository,
    private val aiEngine: AiEngine,
    private val modelFileInstaller: ModelFileInstaller,
    private val diagnosticsJournal: DiagnosticsJournal,
    private val diagnosticsArchiveWriter: DiagnosticsArchiveWriter,
) : ViewModel() {
    private val _state = MutableStateFlow(AssistantUiState())
    val state: StateFlow<AssistantUiState> = _state.asStateFlow()
    private var activeRequestId: String? = null
    private var cancellationRequestedFor: String? = null

    init {
        viewModelScope.launch {
            refreshModels(loadDefault = false)
        }
    }

    fun updateQuestion(question: String) {
        _state.update { it.copy(question = question, notice = null) }
    }

    fun selectAnswerMode(mode: AnswerMode) {
        _state.update { it.copy(answerMode = mode) }
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

    fun exportDiagnostics(uri: Uri?) {
        if (uri == null) return
        viewModelScope.launch {
            runCatching {
                diagnosticsArchiveWriter.write(uri, diagnosticsJournal.exportArchive())
            }.onSuccess {
                _state.update { it.copy(notice = "Diagnostics exported.", error = null) }
            }.onFailure { error ->
                _state.update { it.copy(error = error.message ?: "Could not export diagnostics.", notice = null) }
            }
        }
    }

    fun clearDiagnostics() {
        viewModelScope.launch {
            runCatching { diagnosticsJournal.clear() }
                .onSuccess { _state.update { it.copy(notice = "Diagnostics cleared.", error = null) } }
                .onFailure { error ->
                    _state.update { it.copy(error = error.message ?: "Could not clear diagnostics.", notice = null) }
                }
        }
    }

    fun ask() {
        val question = state.value.question.trim()
        if (question.isBlank()) return

        viewModelScope.launch {
            val requestId = UUID.randomUUID().toString()
            activeRequestId = requestId
            cancellationRequestedFor = null
            val askStartedAt = System.currentTimeMillis()
            val selectedModelId = state.value.selectedModelId ?: "none"
            val questionHash = question.sha256()
            Log.i(TAG, "requestId=$requestId stage=ask outcome=started queryHash=$questionHash modelId=$selectedModelId questionChars=${question.length}")
            recordDiagnostic(
                requestId = requestId,
                stage = "ask",
                outcome = "started",
                fields = mapOf(
                    "queryHash" to questionHash,
                    "modelId" to selectedModelId,
                    "questionChars" to question.length.toString(),
                ),
            )
            _state.update {
                it.copy(
                    question = "",
                    isGenerating = true,
                    error = null,
                    messages = it.messages + ChatMessage(question, isUser = true, requestId = requestId),
                )
            }

            runCatching {
                val retrievalStartedAt = System.currentTimeMillis()
                val previousQuestion = state.value.messages.dropLast(1).lastOrNull { it.isUser }?.text
                val queryPlan = AssistantQueryPlanner.plan(question, previousQuestion)
                recordDiagnostic(
                    requestId = requestId,
                    stage = "query-plan",
                    outcome = "completed",
                    fields = mapOf(
                        "queryHash" to questionHash,
                        "followUp" to queryPlan.isFollowUp.toString(),
                        "retrievalMode" to "lexical",
                    ),
                )
                Log.i(TAG, "requestId=$requestId stage=query-plan outcome=completed queryHash=$questionHash followUp=${queryPlan.isFollowUp} retrievalChars=${queryPlan.retrievalText.length}")
                val results = repository.search(
                    RetrievalQuery(
                        text = queryPlan.retrievalText,
                        requestId = requestId,
                        limit = ASSISTANT_RETRIEVAL_LIMIT,
                    ),
                )
                Log.i(
                    TAG,
                    "requestId=$requestId stage=retrieval outcome=completed resultCount=${results.size} elapsedMs=${System.currentTimeMillis() - retrievalStartedAt} sourceIds=${results.sourceIdSummary()}",
                )
                recordDiagnostic(
                    requestId = requestId,
                    stage = "retrieval",
                    outcome = "completed",
                    elapsedMs = System.currentTimeMillis() - retrievalStartedAt,
                    fields = mapOf(
                        "candidateCount" to results.size.toString(),
                        "candidateIds" to results.joinToString(",") { it.sourceId },
                    ),
                )
                val evidence = results.map { result ->
                    AiEvidence(
                        sourceId = result.sourceId,
                        citationLabel = result.citationLabel,
                        text = result.snippet,
                    )
                }
                val evidenceBrief = EvidenceBriefBuilder.build(question, evidence)
                Log.i(
                    TAG,
                    "requestId=$requestId stage=evidence-selection outcome=completed candidateCount=${results.size} selectedCount=${evidenceBrief.items.size} evidenceChars=${evidenceBrief.text.length} selectedSourceIds=${evidenceBrief.sourceIds.joinToString(",")}",
                )
                recordDiagnostic(
                    requestId = requestId,
                    stage = "evidence-selection",
                    outcome = if (evidenceBrief.isEmpty) "empty" else "completed",
                    fields = mapOf(
                        "candidateCount" to results.size.toString(),
                        "selectedEvidenceIds" to evidenceBrief.sourceIds.joinToString(","),
                        "evidenceChars" to evidenceBrief.text.length.toString(),
                    ),
                )
                if (results.isEmpty() || evidenceBrief.isEmpty) {
                    return@runCatching ChatMessage(
                        text = "I couldn't find enough support in the loaded books for this question.\n\nTry naming a specific rule, character, place, or book term, or ask a broader question.",
                        isUser = false,
                        requestId = requestId,
                    )
                }
                val generationStartedAt = System.currentTimeMillis()
                Log.i(
                    TAG,
                    "requestId=$requestId stage=generation outcome=started modelId=$selectedModelId timeoutMs=$ASK_TIMEOUT_MILLIS",
                )
                val response = withTimeoutOrNull(ASK_TIMEOUT_MILLIS) {
                    state.value.selectedModelId?.let { modelId ->
                        val loadStartedAt = System.currentTimeMillis()
                        recordDiagnostic(requestId, "model-load", "started", fields = mapOf("modelId" to modelId))
                        aiEngine.load(modelId)
                        recordDiagnostic(
                            requestId,
                            "model-load",
                            "completed",
                            elapsedMs = System.currentTimeMillis() - loadStartedAt,
                            fields = mapOf("modelId" to modelId),
                        )
                    }
                    aiEngine.generate(
                        AiRequest(
                            requestId = requestId,
                            originalQuestion = question,
                            evidence = evidenceBrief.items,
                            answerMode = state.value.answerMode,
                        ),
                    )
                } ?: run {
                    Log.w(
                        TAG,
                        "requestId=$requestId stage=generation outcome=timeout modelId=$selectedModelId elapsedMs=${System.currentTimeMillis() - generationStartedAt}",
                    )
                    aiEngine.cancel()
                    recordDiagnostic(
                        requestId = requestId,
                        stage = "generation",
                        outcome = "timeout",
                        elapsedMs = System.currentTimeMillis() - generationStartedAt,
                        fields = mapOf("modelId" to selectedModelId, "deadlineMs" to ASK_TIMEOUT_MILLIS.toString()),
                    )
                    AiResponse(
                        requestId = requestId,
                        text = "The local model took too long to answer. Try a narrower question or switch to the grounded fallback model.",
                        citationIds = evidenceBrief.sourceIds,
                        provenance = AiAnswerProvenance.systemMessage(
                            engineId = "assistant-timeout",
                            label = "Assistant timeout",
                        ),
                    )
                }
                if (cancellationRequestedFor == requestId) {
                    return@runCatching ChatMessage(
                        text = "Answer cancelled.",
                        isUser = false,
                        requestId = requestId,
                    )
                }
                require(response.requestId == requestId) { "AI response request ID did not match the active request." }
                Log.i(
                    TAG,
                    "requestId=$requestId stage=generation outcome=completed modelId=$selectedModelId outputChars=${response.text.length} responseCitationCount=${response.citationIds.size} elapsedMs=${System.currentTimeMillis() - generationStartedAt}",
                )
                val (citations, retrievedEvidence) = resolveAnswerSources(response, results)
                recordDiagnostic(
                    requestId = requestId,
                    stage = "display",
                    outcome = "completed",
                    elapsedMs = System.currentTimeMillis() - generationStartedAt,
                    fields = mapOf(
                        "responseCitationIds" to response.citationIds.joinToString(","),
                        "displayedSourceIds" to citations.joinToString(",") { it.sourceId },
                        "retrievedEvidenceIds" to retrievedEvidence.joinToString(",") { it.sourceId },
                    ),
                )
                ChatMessage(
                    text = response.text,
                    isUser = false,
                    requestId = requestId,
                    citations = citations,
                    retrievedEvidence = retrievedEvidence,
                    answerProvenance = response.provenance,
                )
            }.onSuccess { message ->
                recordDiagnostic(
                    requestId = requestId,
                    stage = "ask",
                    outcome = "completed",
                    elapsedMs = System.currentTimeMillis() - askStartedAt,
                    fields = mapOf("modelId" to selectedModelId),
                )
                Log.i(
                    TAG,
                    "requestId=$requestId stage=ask outcome=completed modelId=$selectedModelId totalElapsedMs=${System.currentTimeMillis() - askStartedAt}",
                )
                _state.update {
                    it.copy(isGenerating = false, messages = it.messages + message)
                }
                activeRequestId = null
            }.onFailure { error ->
                recordDiagnostic(
                    requestId = requestId,
                    stage = "ask",
                    outcome = "failed",
                    elapsedMs = System.currentTimeMillis() - askStartedAt,
                    fields = mapOf(
                        "modelId" to selectedModelId,
                        "errorType" to error::class.java.simpleName,
                    ),
                )
                Log.e(
                    TAG,
                    "requestId=$requestId stage=ask outcome=failed modelId=$selectedModelId totalElapsedMs=${System.currentTimeMillis() - askStartedAt}",
                    error,
                )
                _state.update {
                    if (cancellationRequestedFor == requestId) {
                        it.copy(
                            isGenerating = false,
                            notice = "Answer cancelled.",
                            error = null,
                            messages = it.messages + ChatMessage("Answer cancelled.", false, requestId),
                        )
                    } else {
                        it.copy(isGenerating = false, error = error.message ?: "Could not ask the books.")
                    }
                }
                activeRequestId = null
            }
        }
    }

    fun cancel() {
        val requestId = activeRequestId ?: return
        cancellationRequestedFor = requestId
        _state.update { it.copy(notice = "Cancelling answer...") }
        viewModelScope.launch {
            aiEngine.cancel()
            recordDiagnostic(requestId, "generation", "cancel-requested")
        }
    }

    companion object {
        private const val ASK_TIMEOUT_MILLIS = 40_000L
        private const val ASSISTANT_RETRIEVAL_LIMIT = 4
        private const val TAG = "GmnAssistant"

        fun factory(
            repository: RetrievalRepository,
            aiEngine: AiEngine,
            modelFileInstaller: ModelFileInstaller,
            diagnosticsJournal: DiagnosticsJournal,
            diagnosticsArchiveWriter: DiagnosticsArchiveWriter,
        ): ViewModelProvider.Factory =
            object : ViewModelProvider.Factory {
                @Suppress("UNCHECKED_CAST")
                override fun <T : ViewModel> create(modelClass: Class<T>): T =
                    AssistantViewModel(
                        repository,
                        aiEngine,
                        modelFileInstaller,
                        diagnosticsJournal,
                        diagnosticsArchiveWriter,
                    ) as T
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

    private suspend fun recordDiagnostic(
        requestId: String,
        stage: String,
        outcome: String,
        elapsedMs: Long? = null,
        fields: Map<String, String> = emptyMap(),
    ) {
        runCatching {
            diagnosticsJournal.record(
                DiagnosticEvent(
                    requestId = requestId,
                    timestampEpochMillis = System.currentTimeMillis(),
                    stage = stage,
                    outcome = outcome,
                    elapsedMs = elapsedMs,
                    fields = fields,
                ),
            )
        }.onFailure { error ->
            Log.w(TAG, "Diagnostics event failed requestId=$requestId stage=$stage", error)
        }
    }
}

internal fun resolveAnswerSources(
    response: AiResponse,
    results: List<RetrievalResult>,
): Pair<List<RetrievalResult>, List<RetrievalResult>> {
    val citationIds = response.citationIds.toSet()
    val cited = results.filter { it.sourceId in citationIds }.distinctBy { it.sourceId }
    val retrieved = results.filterNot { it.sourceId in citationIds }.distinctBy { it.sourceId }
    return cited to retrieved
}

private fun List<RetrievalResult>.sourceIdSummary(): String =
    take(MAX_LOGGED_CITATIONS)
        .joinToString(separator = ",") { it.sourceId }
        .ifBlank { "none" }

private const val MAX_LOGGED_CITATIONS = 5
private const val DIAGNOSTICS_EXPORT_NAME = "game-master-notes-diagnostics.zip"

private fun String.sha256(): String =
    MessageDigest.getInstance("SHA-256")
        .digest(toByteArray(Charsets.UTF_8))
        .joinToString("") { byte -> "%02x".format(byte) }

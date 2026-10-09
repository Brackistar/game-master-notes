package com.brackistar.gamemasternotes.core.ai

object EvidenceBriefBuilder {
    fun build(question: String, evidence: List<AiEvidence>): EvidenceBrief {
        if (evidence.isEmpty()) return EvidenceBrief(emptyList())
        val intent = DeterministicQuestionIntentParser.parse(question)
        val evidenceItems = evidence
            .asSequence()
            .mapIndexedNotNull { index, item -> item.toEvidenceItem(intent, index) }
            .sortedWith(compareByDescending<EvidenceItem> { it.score }.thenBy { it.index })
            .take(MAX_EVIDENCE_ITEMS)
            .toList()

        if (evidenceItems.isEmpty()) return EvidenceBrief(emptyList())

        val excerpts = evidenceItems
            .mapIndexed { index, item ->
                AiEvidence(
                    sourceId = item.sourceId,
                    citationLabel = item.citationLabel,
                    text = item.text,
                    evidenceId = "E${index + 1}",
                )
            }
            .takeWithinCharacterBudget(MAX_TOTAL_EVIDENCE_CHARS)
        return EvidenceBrief(excerpts)
    }

    internal fun analyze(question: String): DeterministicQuestionIntent =
        DeterministicQuestionIntentParser.parse(question)

    private fun AiEvidence.toEvidenceItem(intent: DeterministicQuestionIntent, index: Int): EvidenceItem? {
        val paragraphs = text
            .lineSequence()
            .map { it.normalizeWhitespace() }
            .filter { it.isNotBlank() }
            .toList()
        if (paragraphs.isEmpty()) return null

        val selectedText = paragraphs
            .map { paragraph -> paragraph to intent.score(paragraph) }
            .filter { (_, match) -> match.isRelevant }
            .sortedByDescending { (_, match) -> match.score }
            .map { (paragraph, _) -> paragraph }
            .take(MAX_PARAGRAPHS_PER_SOURCE)
            .joinToString("\n\n")
            .takeCleanly(MAX_EVIDENCE_CHARS)
        if (selectedText.isBlank()) return null

        val match = intent.score(selectedText)
        if (!match.isRelevant) return null

        return EvidenceItem(
            sourceId = sourceId,
            citationLabel = citationLabel,
            text = selectedText,
            score = match.score,
            index = index,
        )
    }

    private fun String.normalizeWhitespace(): String =
        replace(Regex("""\s+"""), " ").trim()

    private fun String.takeCleanly(maxChars: Int): String {
        if (length <= maxChars) return this
        val clipped = take(maxChars)
        return clipped.substringBeforeLast(" ").ifBlank { clipped }.trimEnd() + "..."
    }

    private fun List<AiEvidence>.takeWithinCharacterBudget(maxChars: Int): List<AiEvidence> {
        var usedChars = 0
        val bounded = mutableListOf<AiEvidence>()
        for (item in this) {
            val separatorChars = if (usedChars == 0) 0 else 2
            val availableChars = maxChars - usedChars - separatorChars
            if (availableChars <= 0) break
            val boundedText = item.text.takeCleanly(availableChars)
            if (boundedText.isBlank()) break
            usedChars += separatorChars + boundedText.length
            bounded += item.copy(text = boundedText)
        }
        return bounded
    }

    private data class EvidenceItem(
        val sourceId: String,
        val citationLabel: String,
        val text: String,
        val score: Int,
        val index: Int,
    )

    private const val MAX_EVIDENCE_ITEMS = 4
    private const val MAX_PARAGRAPHS_PER_SOURCE = 3
    private const val MAX_EVIDENCE_CHARS = 1_400
    private const val MAX_TOTAL_EVIDENCE_CHARS = 1_800
}

data class EvidenceBrief(
    val items: List<AiEvidence>,
) {
    val sourceIds: List<String> = items.map { it.sourceId }.distinct()
    val citationLabels: List<String> = items.map { it.citationLabel }.distinct()
    val evidenceIds: List<String> = items.mapNotNull { it.evidenceId }.distinct()

    val isEmpty: Boolean
        get() = items.isEmpty()

    val text: String = toPromptText()

    fun toPromptText(): String {
        if (items.isEmpty()) return ""
        return buildString {
            appendLine("Evidence:")
            items.forEachIndexed { index, item ->
                if (index > 0) appendLine()
                append(index + 1)
                append(". [")
                append(requireNotNull(item.evidenceId) { "Prompt evidence requires a stable evidence ID." })
                append("] ")
                appendLine(item.text)
            }
        }.trim()
    }

    fun toReadableAnswer(): String {
        if (items.isEmpty()) {
            return "I could not find relevant passages in the loaded books for that question."
        }
        return buildString {
            appendLine("I found these relevant passages in the loaded books:")
            items.forEach { item ->
                appendLine()
                appendLine("[${item.citationLabel}]")
                appendLine(item.text)
            }
        }.trim()
    }
}

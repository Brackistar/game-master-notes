package com.brackistar.gamemasternotes.core.ai

object EvidenceBriefBuilder {
    fun build(question: String, context: String): EvidenceBrief {
        if (context.isBlank()) return EvidenceBrief(items = emptyList(), citationIds = emptyList())
        parseExistingBrief(context)?.let { return it }

        val terms = question.significantTerms()
        val evidenceItems = context
            .split("\n\n")
            .asSequence()
            .mapIndexedNotNull { index, section -> section.toEvidenceItem(terms, index) }
            .sortedWith(compareByDescending<EvidenceItem> { it.score }.thenBy { it.index })
            .take(MAX_EVIDENCE_ITEMS)
            .toList()

        if (evidenceItems.isEmpty()) return EvidenceBrief(items = emptyList(), citationIds = emptyList())

        return EvidenceBrief(
            items = evidenceItems.map { EvidenceExcerpt(citation = it.citation, text = it.text) },
            citationIds = evidenceItems.map { it.citation }.distinct(),
        )
    }

    private fun String.toEvidenceItem(terms: Set<String>, index: Int): EvidenceItem? {
        val lines = lineSequence().filter { it.isNotBlank() }.toList()
        if (lines.isEmpty()) return null

        val citation = lines.first()
            .removePrefix("[")
            .substringBefore("]")
            .ifBlank { "source" }
        val body = lines.drop(1).joinToString(" ").normalizeWhitespace()
        if (body.isBlank()) return null

        val selectedText = body
            .split(Regex("""(?<=[.!?])\s+"""))
            .asSequence()
            .map { sentence -> sentence.trim() }
            .filter { it.isNotBlank() }
            .map { sentence -> sentence to sentence.scoreAgainst(terms) }
            .sortedByDescending { it.second }
            .take(MAX_SENTENCES_PER_SOURCE)
            .map { it.first }
            .joinToString(" ")
            .ifBlank { body }
            .take(MAX_EVIDENCE_CHARS)

        return EvidenceItem(
            citation = citation,
            text = selectedText,
            score = selectedText.scoreAgainst(terms),
            index = index,
        )
    }

    private fun String.significantTerms(): Set<String> =
        lowercase()
            .split(Regex("""[^a-z0-9]+"""))
            .filter { it.length >= MIN_TERM_LENGTH && it !in STOP_WORDS }
            .toSet()

    private fun String.scoreAgainst(terms: Set<String>): Int {
        if (terms.isEmpty()) return 0
        val lower = lowercase()
        return terms.count { it in lower }
    }

    private fun String.normalizeWhitespace(): String =
        replace(Regex("""\s+"""), " ").trim()

    private fun parseExistingBrief(context: String): EvidenceBrief? {
        val items = context.lineSequence()
            .map { it.trim() }
            .filter { it.isNotBlank() && !it.equals("Evidence:", ignoreCase = true) }
            .mapNotNull { line ->
                val match = NUMBERED_CITATION.find(line) ?: return@mapNotNull null
                EvidenceExcerpt(
                    citation = match.groupValues[1],
                    text = match.groupValues[2].normalizeWhitespace(),
                )
            }
            .toList()
        if (items.isEmpty()) return null
        return EvidenceBrief(
            items = items,
            citationIds = items.map { it.citation }.distinct(),
        )
    }

    private data class EvidenceItem(
        val citation: String,
        val text: String,
        val score: Int,
        val index: Int,
    )

    private const val MAX_EVIDENCE_ITEMS = 2
    private const val MAX_SENTENCES_PER_SOURCE = 1
    private const val MAX_EVIDENCE_CHARS = 180
    private const val MIN_TERM_LENGTH = 3
    private val NUMBERED_CITATION = Regex("""^\d+\.\s+\[([^\]]+)]\s*(.*)$""")

    private val STOP_WORDS = setOf(
        "the",
        "and",
        "for",
        "from",
        "what",
        "when",
        "where",
        "which",
        "with",
        "about",
        "that",
        "this",
        "does",
        "have",
        "how",
        "are",
        "was",
        "were",
    )
}

data class EvidenceBrief(
    val items: List<EvidenceExcerpt>,
    val citationIds: List<String>,
) {
    val text: String = toPromptText()

    fun toPromptText(): String {
        if (items.isEmpty()) return ""
        return buildString {
            appendLine("Evidence:")
            items.forEachIndexed { index, item ->
                if (index > 0) appendLine()
                append(index + 1)
                append(". [")
                append(item.citation)
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
                appendLine("[${item.citation}]")
                appendLine(item.text)
            }
        }.trim()
    }
}

data class EvidenceExcerpt(
    val citation: String,
    val text: String,
)

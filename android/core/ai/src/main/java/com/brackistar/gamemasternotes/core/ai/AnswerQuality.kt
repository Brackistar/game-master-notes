package com.brackistar.gamemasternotes.core.ai

data class AnswerQuality(
    val usable: Boolean,
    val reason: String? = null,
)

fun validateGroundedAnswer(question: String, answer: String, evidence: EvidenceBrief): AnswerQuality {
    val normalized = answer.replace(Regex("\\s+"), " ").trim()
    if (normalized.length < 24) return AnswerQuality(false, "answer-too-short")
    if (normalized.contains("Question:", ignoreCase = true) || normalized.contains("Evidence:", ignoreCase = true)) {
        return AnswerQuality(false, "prompt-echo")
    }
    val citedIds = normalized.extractCitationIds()
    if (citedIds.any { it !in evidence.citationIds }) return AnswerQuality(false, "unsupported-citation")
    if (evidence.citationIds.isNotEmpty() && citedIds.none { it in evidence.citationIds }) {
        return AnswerQuality(false, "missing-supported-citation")
    }
    if (normalized.count { it.isLetter() } < 12) return AnswerQuality(false, "too-few-letters")

    val terms = question.significantTerms()
    val coveredTerms = terms.count { normalized.contains(it, ignoreCase = true) }
    if (terms.size >= 2 && coveredTerms == 0) return AnswerQuality(false, "question-not-covered")
    return AnswerQuality(true)
}

private fun String.significantTerms(): Set<String> =
    lowercase()
        .split(Regex("[^a-z0-9]+"))
        .filter { it.length >= 4 && it !in STOP_WORDS }
        .toSet()

private val STOP_WORDS = setOf("what", "when", "where", "which", "who", "whom", "that", "this", "does", "have", "with", "from", "into", "about", "explain", "tell", "give")

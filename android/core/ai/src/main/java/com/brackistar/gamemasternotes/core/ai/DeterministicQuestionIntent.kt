package com.brackistar.gamemasternotes.core.ai

internal enum class QuestionIntentType {
    Procedure,
    Definition,
    Permission,
    Explanation,
    Comparison,
    Unknown,
}

internal data class DeterministicQuestionIntent(
    val type: QuestionIntentType,
    val subjectTerms: Set<String>,
    val supportingTerms: Set<String>,
    val subjectPhrases: Set<String>,
) {
    fun score(text: String): EvidenceMatchScore {
        if (subjectTerms.isEmpty() && supportingTerms.isEmpty()) {
            return EvidenceMatchScore()
        }

        val normalizedText = text.normalizeQuestionText()
        val textTerms = normalizedText
            .split(' ')
            .filter(String::isNotBlank)
            .map { token -> token.normalizeKeywordTerm() }
            .toSet()

        val matchedSubjectTerms = subjectTerms.filterTo(linkedSetOf()) { it in textTerms }
        val matchedSupportingTerms = supportingTerms
            .filterTo(linkedSetOf()) { it in textTerms && it !in matchedSubjectTerms }
        val phraseMatches = subjectPhrases.count { phrase -> phrase in normalizedText }
        val score = (matchedSubjectTerms.size * SUBJECT_TERM_WEIGHT) +
            (matchedSupportingTerms.size * SUPPORTING_TERM_WEIGHT) +
            (phraseMatches * SUBJECT_PHRASE_WEIGHT)
        return EvidenceMatchScore(
            score = score,
            matchedSubjectTerms = matchedSubjectTerms,
            matchedSupportingTerms = matchedSupportingTerms,
            phraseMatches = phraseMatches,
        )
    }

    companion object {
        internal const val SUBJECT_TERM_WEIGHT = 5
        internal const val SUPPORTING_TERM_WEIGHT = 2
        internal const val SUBJECT_PHRASE_WEIGHT = 7
        internal const val MIN_RELEVANCE_SCORE = 5
    }
}

internal data class EvidenceMatchScore(
    val score: Int = 0,
    val matchedSubjectTerms: Set<String> = emptySet(),
    val matchedSupportingTerms: Set<String> = emptySet(),
    val phraseMatches: Int = 0,
) {
    val isRelevant: Boolean
        get() = matchedSubjectTerms.isNotEmpty() && score >= DeterministicQuestionIntent.MIN_RELEVANCE_SCORE
}

internal object DeterministicQuestionIntentParser {
    fun parse(question: String): DeterministicQuestionIntent {
        val normalizedQuestion = question.normalizeQuestionText()
        if (normalizedQuestion.isBlank()) {
            return DeterministicQuestionIntent(
                type = QuestionIntentType.Unknown,
                subjectTerms = emptySet(),
                supportingTerms = emptySet(),
                subjectPhrases = emptySet(),
            )
        }

        val (intentType, subjectText) = QUESTION_PREFIXES.firstNotNullOfOrNull { (regex, type) ->
            regex.find(normalizedQuestion)?.let { type to normalizedQuestion.removeRange(it.range).trim() }
        } ?: (QuestionIntentType.Unknown to normalizedQuestion)

        val subjectTokens = subjectText
            .split(' ')
            .map { token -> token.normalizeKeywordTerm() }
            .filter { token -> token.length >= MIN_TERM_LENGTH && token !in STOP_WORDS }

        if (subjectTokens.isEmpty()) {
            return DeterministicQuestionIntent(
                type = intentType,
                subjectTerms = emptySet(),
                supportingTerms = emptySet(),
                subjectPhrases = emptySet(),
            )
        }

        val prioritizedSubjectTerms = subjectTokens
            .filterNot { token -> token in GENERIC_ACTION_TERMS }
            .toLinkedSet()
            .ifEmpty { subjectTokens.toLinkedSet() }
        val supportingTerms = subjectTokens
            .filter { token -> token !in prioritizedSubjectTerms }
            .toLinkedSet()
        val phrases = buildPhrases(subjectTokens)

        return DeterministicQuestionIntent(
            type = intentType,
            subjectTerms = prioritizedSubjectTerms,
            supportingTerms = supportingTerms,
            subjectPhrases = phrases,
        )
    }

    private fun buildPhrases(tokens: List<String>): Set<String> =
        buildSet {
            if (tokens.size < 2) return@buildSet
            tokens.windowed(size = 2, step = 1, partialWindows = false)
                .mapTo(this) { it.joinToString(" ") }
            if (tokens.size >= 3) {
                tokens.windowed(size = 3, step = 1, partialWindows = false)
                    .mapTo(this) { it.joinToString(" ") }
            }
        }

    private fun List<String>.toLinkedSet(): LinkedHashSet<String> = LinkedHashSet(this)

    private const val MIN_TERM_LENGTH = 3
    private val QUESTION_PREFIXES = listOf(
        Regex("^how\\s+do\\s+i\\b") to QuestionIntentType.Procedure,
        Regex("^how\\s+can\\s+i\\b") to QuestionIntentType.Procedure,
        Regex("^how\\s+does\\b") to QuestionIntentType.Procedure,
        Regex("^what\\s+is\\b") to QuestionIntentType.Definition,
        Regex("^what\\s+are\\b") to QuestionIntentType.Definition,
        Regex("^what\\s+does\\b") to QuestionIntentType.Explanation,
        Regex("^can\\s+i\\b") to QuestionIntentType.Permission,
        Regex("^may\\s+i\\b") to QuestionIntentType.Permission,
        Regex("^explain\\b") to QuestionIntentType.Explanation,
        Regex("^tell\\s+me\\s+about\\b") to QuestionIntentType.Explanation,
        Regex("^compare\\b") to QuestionIntentType.Comparison,
    )
    private val GENERIC_ACTION_TERMS = setOf(
        "learn",
        "use",
        "make",
        "take",
        "tell",
        "explain",
        "describe",
        "show",
        "know",
        "give",
        "need",
        "find",
    )
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
        "can",
        "may",
        "you",
        "your",
        "into",
        "them",
        "they",
        "their",
        "would",
        "could",
        "should",
        "while",
        "then",
        "than",
        "just",
        "also",
        "like",
        "about",
        "over",
        "under",
        "once",
        "after",
        "before",
        "because",
        "through",
        "a",
        "an",
        "i",
        "me",
        "my",
        "to",
        "of",
        "in",
        "on",
        "at",
        "by",
        "or",
        "if",
        "it",
        "be",
        "is",
        "do",
        "we",
    )
}

internal fun String.normalizeQuestionText(): String =
    lowercase()
        .replace(Regex("""[^a-z0-9]+"""), " ")
        .trim()

private fun String.normalizeKeywordTerm(): String =
    when {
        length > 4 && endsWith("ies") -> dropLast(3) + "y"
        length > 4 && endsWith("es") -> dropLast(2)
        length > 3 && endsWith("s") -> dropLast(1)
        else -> this
    }

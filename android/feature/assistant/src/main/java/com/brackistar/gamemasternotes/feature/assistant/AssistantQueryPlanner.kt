package com.brackistar.gamemasternotes.feature.assistant

data class QueryPlan(
    val retrievalText: String,
    val isFollowUp: Boolean,
)

object AssistantQueryPlanner {
    fun plan(question: String, previousQuestion: String?): QueryPlan {
        val normalized = question.trim()
        if (previousQuestion.isNullOrBlank() || !isFollowUp(normalized)) {
            return QueryPlan(normalized, isFollowUp = false)
        }
        return QueryPlan("${previousQuestion.trim()} $normalized", isFollowUp = true)
    }

    private fun isFollowUp(question: String): Boolean {
        val lower = question.lowercase()
        return question.length <= 100 && (
            lower.startsWith("what about ") ||
                lower.startsWith("how does that") ||
                lower.startsWith("how do they") ||
                lower.startsWith("why does that") ||
                lower.startsWith("and ") ||
                lower.startsWith("also ") ||
                lower.startsWith("give me an example") ||
                lower.startsWith("tell me more") ||
                lower.contains(" that ") ||
                lower.contains(" it ")
            )
    }
}

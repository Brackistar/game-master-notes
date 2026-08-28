package com.brackistar.gamemasternotes.feature.assistant

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class AssistantQueryPlannerTest {
    @Test
    fun followUpIncludesPreviousQuestionForRetrieval() {
        val plan = AssistantQueryPlanner.plan(
            question = "What about the second type?",
            previousQuestion = "How does paradox affect vulgar magic?",
        )

        assertTrue(plan.isFollowUp)
        assertEquals("How does paradox affect vulgar magic? What about the second type?", plan.retrievalText)
    }

    @Test
    fun standaloneQuestionIsNotPollutedByPreviousTurn() {
        val plan = AssistantQueryPlanner.plan("What is a cabal?", "How does paradox work?")

        assertEquals("What is a cabal?", plan.retrievalText)
        assertTrue(!plan.isFollowUp)
    }
}

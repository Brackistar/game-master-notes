package com.brackistar.gamemasternotes.core.retrieval

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class HybridRetrievalTest {
    @Test fun fusionRewardsCandidatesRecoveredByBothRankings() {
        val fused = ReciprocalRankFusion.fuse(
            lexical = listOf(RankedCandidate("literal", 8.0), RankedCandidate("shared", 7.0)),
            vector = listOf(RankedCandidate("shared", .9), RankedCandidate("semantic", .8)),
            limit = 3,
        )
        assertEquals("shared", fused.first().sourceId)
        assertEquals(3, fused.map { it.sourceId }.distinct().size)
    }

    @Test fun vectorSearchIsBoundedAndDeterministic() {
        val results = BruteForceVectorSearch.search(
            floatArrayOf(1f, 0f),
            listOf("b" to floatArrayOf(1f, 0f), "a" to floatArrayOf(1f, 0f), "c" to floatArrayOf(0f, 1f)),
            limit = 2,
        )
        assertEquals(listOf("a", "b"), results.map { it.sourceId })
    }

    @Test fun relationshipExpansionPrefersAdjacentAndOverlappingChunks() {
        val seed = RelationshipCandidate("seed", "doc", 4, 10, 11, setOf("arden"))
        val results = BoundedRelationshipExpansion.expand(
            listOf(seed),
            listOf(
                RelationshipCandidate("name", "other", 9, 2, 2, setOf("arden")),
                RelationshipCandidate("next", "doc", 5, 11, 12),
                RelationshipCandidate("far", "doc", 20, 50, 51),
            ),
            maxAdded = 2,
        )
        assertEquals(listOf("next", "name"), results.map { it.sourceId })
        assertTrue(results.none { it.sourceId == "far" })
    }

    @Test fun syntheticEvaluationCorpusHasTheRequiredCoverage() {
        assertEquals(50, SyntheticEvaluationCorpus.cases.size)
        assertTrue(SyntheticEvaluationCorpus.cases.count { it.category in setOf("paraphrase", "relationship") } >= 15)
        assertTrue(SyntheticEvaluationCorpus.cases.any { it.category == "unanswerable" })
        assertTrue(SyntheticEvaluationCorpus.cases.any { it.category == "adversarial" })
    }

    @Test fun retrievalMetricsAreDeterministic() {
        assertEquals(.5, RetrievalMetrics.recallAt(setOf("a", "b"), listOf("x", "a", "b"), 2), 0.0)
        assertEquals(.5, RetrievalMetrics.reciprocalRank(setOf("a"), listOf("x", "a")), 0.0)
    }

    @Test fun wordPieceTokenizerMatchesPinnedBertVocabulary() {
        val vocab = java.io.File("src/main/assets/minilm/vocab.txt").readLines()
        val encoded = WordPieceTokenizer(vocab).encode("The Silver Ladder guards the hidden library.", 12)
        assertEquals(
            listOf(101L, 1996L, 3165L, 10535L, 4932L, 1996L, 5023L, 3075L, 1012L, 102L, 0L, 0L),
            encoded.inputIds.toList(),
        )
        assertEquals(10, encoded.attentionMask.count { it == 1L })
    }
}

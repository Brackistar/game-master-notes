from __future__ import annotations

from pack_builder.content_processing.content_classification import (
    ContentLabel,
    LabeledExample,
    classify_paragraph,
    classifier_mode,
    evaluate_examples,
)


RULE_EXAMPLES = [
    "Roll 1d20 and add your Strength modifier. On a success, the target takes 2d6 damage.",
    "Armor Class: 16; Hit Points: 45; Speed: 30 ft.",
    "1. Choose a target.\n2. Roll an attack.\n3. Apply damage.",
    "The target must make a saving throw. On a failure, it takes 3d8 damage.",
]

NARRATIVE_EXAMPLES = [
    "The riders arrived at the ancient village as moonlight touched the ruined temple.",
    "“We should leave before dawn,” Mara whispered.",
    "The ancient kingdom stood beyond the mountain. The river crossed the forgotten forest. A temple watched over the village.",
    "They traveled through the forest and discovered the ruins beside the river.",
]

AMBIGUOUS_EXAMPLES = [
    "A hero looked toward the chamber.",
    "The travelers carried 20 torches into the cavern.",
    "When Mara arrived at the village, roll 1d20 and make a saving throw.",
    "A quiet road crossed the valley toward the distant walls.",
]


def test_classifier_covers_rules_narrative_and_abstention() -> None:
    assert all(classify_paragraph(text).label is ContentLabel.RULESET for text in RULE_EXAMPLES)
    assert all(classify_paragraph(text).label is ContentLabel.NARRATIVE for text in NARRATIVE_EXAMPLES)
    assert all(classify_paragraph(text).label is ContentLabel.AMBIGUOUS for text in AMBIGUOUS_EXAMPLES)


def test_classifier_is_deterministic_and_reports_feature_ids() -> None:
    first = classify_paragraph(RULE_EXAMPLES[0])
    second = classify_paragraph(RULE_EXAMPLES[0])

    assert first == second
    assert first.margin == abs(first.ruleset_score - first.narrative_score)
    assert "dice_notation" in first.matched_feature_ids
    assert "rule_modality" in first.matched_feature_ids


def test_non_english_mode_uses_structural_features_only() -> None:
    result = classify_paragraph("CA: 16; PV: 45; Speed: 30 ft.", language="es")
    dialogue = classify_paragraph("“Debemos partir antes del amanecer”, dijo Mara.", language="es")

    assert classifier_mode("es") == "structural-fallback"
    assert result.label is ContentLabel.RULESET
    assert dialogue.label is ContentLabel.NARRATIVE
    assert "rule_terms" not in result.matched_feature_ids


def test_synthetic_holdout_meets_high_purity_contract() -> None:
    examples = [
        *(LabeledExample(text, ContentLabel.RULESET) for text in RULE_EXAMPLES),
        *(LabeledExample(text, ContentLabel.NARRATIVE) for text in NARRATIVE_EXAMPLES),
        *(LabeledExample(text, ContentLabel.AMBIGUOUS) for text in AMBIGUOUS_EXAMPLES),
    ]

    metrics = evaluate_examples(examples)

    assert metrics["ruleset"]["precision"] >= 0.9
    assert metrics["narrative"]["precision"] >= 0.9
    assert metrics["ambiguous_recall"] >= 0.9
    assert metrics["confident_coverage"] >= 0.6

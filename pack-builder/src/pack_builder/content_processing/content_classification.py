from __future__ import annotations

import re
from collections import Counter
from dataclasses import dataclass
from enum import Enum
from typing import Iterable


CLASSIFIER_VERSION = "rules-narrative-v1"
MIN_WINNING_SCORE = 3
MIN_SCORE_MARGIN = 2
MIN_CONFIDENT_CHARACTERS = 1800
MIN_CONFIDENT_PARAGRAPHS = 2


class ContentLabel(str, Enum):
    RULESET = "ruleset"
    NARRATIVE = "narrative"
    AMBIGUOUS = "ambiguous"


@dataclass(frozen=True)
class ClassificationResult:
    label: ContentLabel
    ruleset_score: int
    narrative_score: int
    margin: int
    matched_feature_ids: tuple[str, ...]


@dataclass(frozen=True)
class LabeledExample:
    text: str
    expected: ContentLabel
    language: str = "en"


_STRUCTURAL_FEATURES: tuple[tuple[str, ContentLabel, int, re.Pattern[str]], ...] = (
    ("dice_notation", ContentLabel.RULESET, 4, re.compile(r"\b\d*d(?:4|6|8|10|12|20|100)(?:\s*[+-]\s*\d+)?\b", re.I)),
    ("stat_value", ContentLabel.RULESET, 3, re.compile(r"(?:^|[;|])\s*[\w -]{2,24}:\s*[+-]?\d+(?:\s|$|[;|])", re.I)),
    ("numeric_unit", ContentLabel.RULESET, 2, re.compile(r"\b\d+\s*(?:ft\.?|feet|meters?|metres?|squares?|rounds?|minutes?|hours?|%)\b", re.I)),
    ("procedural_list", ContentLabel.RULESET, 2, re.compile(r"(?:^|\n)\s*(?:\d+[.)]|[-*])\s+", re.M)),
    ("dialogue", ContentLabel.NARRATIVE, 3, re.compile(r"(?:[\"“][^\"”]{12,}[\"”]|^[—-]\s*[A-ZÁÉÍÓÚÑ])", re.M)),
)

_ENGLISH_FEATURES: tuple[tuple[str, ContentLabel, int, re.Pattern[str]], ...] = (
    ("rule_terms", ContentLabel.RULESET, 3, re.compile(r"\b(?:armor class|difficulty class|saving throw|ability check|attack roll|damage roll|hit points?|initiative|bonus action|spell slot)\b", re.I)),
    ("rule_modality", ContentLabel.RULESET, 2, re.compile(r"\b(?:must|may|cannot|can\s+use|roll|reroll|succeed|fail|target|takes?\s+\d+|gains?\s+\d+)\b", re.I)),
    ("condition_result", ContentLabel.RULESET, 2, re.compile(r"\b(?:if|when|whenever|on a success|on a failure|until the end)\b.{0,100}\b(?:then|must|may|takes?|gains?|becomes?|ends?)\b", re.I | re.S)),
    ("scene_event", ContentLabel.NARRATIVE, 3, re.compile(r"\b(?:arrived|departed|entered|walked|rode|travelled|traveled|whispered|remembered|discovered|watched|waited|stood|slept|awoke)\b", re.I)),
    ("setting_prose", ContentLabel.NARRATIVE, 2, re.compile(r"\b(?:village|kingdom|forest|mountain|river|temple|ruins?|tavern|moonlight|ancient|forgotten|legend|centuries ago)\b", re.I)),
)


def classifier_mode(language: str) -> str:
    return "english-lexical-and-structural" if language.lower().startswith("en") else "structural-fallback"


def classify_paragraph(text: str, language: str = "en") -> ClassificationResult:
    normalized = " ".join(text.split())
    ruleset_score = 0
    narrative_score = 0
    matched: list[str] = []
    features = list(_STRUCTURAL_FEATURES)
    if language.lower().startswith("en"):
        features.extend(_ENGLISH_FEATURES)

    for feature_id, label, weight, pattern in features:
        if pattern.search(text):
            matched.append(feature_id)
            if label is ContentLabel.RULESET:
                ruleset_score += weight
            else:
                narrative_score += weight

    sentence_count = len(re.findall(r"[.!?](?:\s|$)", normalized))
    if sentence_count >= 3 and ruleset_score == 0:
        narrative_score += 2
        matched.append("descriptive_sequence")

    winning_score = max(ruleset_score, narrative_score)
    margin = abs(ruleset_score - narrative_score)
    strongly_mixed = ruleset_score >= MIN_WINNING_SCORE and narrative_score >= MIN_WINNING_SCORE
    if winning_score < MIN_WINNING_SCORE or margin < MIN_SCORE_MARGIN or strongly_mixed:
        label = ContentLabel.AMBIGUOUS
    elif ruleset_score > narrative_score:
        label = ContentLabel.RULESET
    else:
        label = ContentLabel.NARRATIVE
    return ClassificationResult(
        label=label,
        ruleset_score=ruleset_score,
        narrative_score=narrative_score,
        margin=margin,
        matched_feature_ids=tuple(matched),
    )


def summarize_classifications(
    results: Iterable[tuple[str, ClassificationResult]],
    *,
    language: str,
) -> dict[str, object]:
    rows = list(results)
    label_counts = Counter(result.label.value for _, result in rows)
    feature_counts = Counter(
        feature_id
        for _, result in rows
        for feature_id in result.matched_feature_ids
    )
    confident_characters = {
        label.value: sum(len(text) for text, result in rows if result.label is label)
        for label in (ContentLabel.RULESET, ContentLabel.NARRATIVE)
    }
    confident_paragraphs = {
        label.value: label_counts[label.value]
        for label in (ContentLabel.RULESET, ContentLabel.NARRATIVE)
    }
    total = len(rows)
    confident = confident_paragraphs["ruleset"] + confident_paragraphs["narrative"]
    return {
        "classifier_version": CLASSIFIER_VERSION,
        "mode": classifier_mode(language),
        "thresholds": {
            "minimum_winning_score": MIN_WINNING_SCORE,
            "minimum_score_margin": MIN_SCORE_MARGIN,
            "minimum_confident_characters": MIN_CONFIDENT_CHARACTERS,
            "minimum_confident_paragraphs": MIN_CONFIDENT_PARAGRAPHS,
        },
        "label_counts": {
            "ruleset": label_counts["ruleset"],
            "narrative": label_counts["narrative"],
            "ambiguous": label_counts["ambiguous"],
        },
        "confident_characters": confident_characters,
        "confident_paragraphs": confident_paragraphs,
        "ambiguous_duplication_count": label_counts["ambiguous"],
        "confident_coverage": round(confident / total, 4) if total else 0.0,
        "feature_counts": dict(sorted(feature_counts.items())),
    }


def evaluate_examples(examples: Iterable[LabeledExample]) -> dict[str, object]:
    rows = [(example, classify_paragraph(example.text, example.language)) for example in examples]
    metrics: dict[str, object] = {}
    for label in (ContentLabel.RULESET, ContentLabel.NARRATIVE):
        true_positive = sum(1 for example, result in rows if example.expected is label and result.label is label)
        predicted = sum(1 for _, result in rows if result.label is label)
        expected = sum(1 for example, _ in rows if example.expected is label)
        precision = true_positive / predicted if predicted else 0.0
        recall = true_positive / expected if expected else 0.0
        f1 = 2 * precision * recall / (precision + recall) if precision + recall else 0.0
        metrics[label.value] = {
            "precision": round(precision, 4),
            "recall": round(recall, 4),
            "f1": round(f1, 4),
            "support": expected,
        }
    ambiguous_examples = sum(1 for example, _ in rows if example.expected is ContentLabel.AMBIGUOUS)
    ambiguous_correct = sum(
        1
        for example, result in rows
        if example.expected is ContentLabel.AMBIGUOUS and result.label is ContentLabel.AMBIGUOUS
    )
    confident = sum(1 for _, result in rows if result.label is not ContentLabel.AMBIGUOUS)
    metrics["ambiguous_recall"] = round(ambiguous_correct / ambiguous_examples, 4) if ambiguous_examples else 0.0
    metrics["confident_coverage"] = round(confident / len(rows), 4) if rows else 0.0
    labels = tuple(ContentLabel)
    metrics["confusion_counts"] = {
        f"{expected.value}->{predicted.value}": sum(
            1
            for example, result in rows
            if example.expected is expected and result.label is predicted
        )
        for expected in labels
        for predicted in labels
    }
    return metrics

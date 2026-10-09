from __future__ import annotations

import re
from collections import Counter

from pack_builder.core_domain.models import (
    ExtractedDocument,
    ExtractedPage,
    replace_document_pages,
)


def clean_documents(
    documents: list[ExtractedDocument],
) -> tuple[list[ExtractedDocument], dict[str, object]]:
    cleaned_documents: list[ExtractedDocument] = []
    removed_lines: dict[str, list[str]] = {}
    hyphen_repairs = 0
    split_word_repairs = 0
    table_line_repairs = 0

    for document in documents:
        repeated_lines = detect_repeated_lines(document.pages)
        cleaned_pages: list[ExtractedPage] = []
        for page in document.pages:
            cleaned_text, repair_counts = clean_page_text(page.text, repeated_lines)
            hyphen_repairs += repair_counts["hyphen"]
            split_word_repairs += repair_counts["split_word"]
            table_line_repairs += repair_counts["table_line"]
            cleaned_pages.append(
                ExtractedPage(
                    page_number=page.page_number,
                    text=cleaned_text,
                    warnings=page.warnings,
                )
            )
        removed_lines[document.document_id] = sorted(repeated_lines)
        cleaned_documents.append(replace_document_pages(document, cleaned_pages))

    return cleaned_documents, {
        "enabled": True,
        "removed_repeated_lines": removed_lines,
        "hyphen_repairs": hyphen_repairs,
        "split_word_repairs": split_word_repairs,
        "table_line_repairs": table_line_repairs,
    }


def detect_repeated_lines(pages: list[ExtractedPage]) -> set[str]:
    if len(pages) < 3:
        return set()

    counts: Counter[str] = Counter()
    for page in pages:
        counts.update(set(candidate_repeated_lines(page.text)))

    threshold = max(3, int(len(pages) * 0.6))
    return {line for line, count in counts.items() if count >= threshold}


def candidate_repeated_lines(text: str) -> list[str]:
    lines = [normalize_line(line) for line in text.splitlines()]
    return [
        line
        for line in lines
        if line and not line.isdigit() and len(line) <= 120
    ]


def clean_page_text(
    text: str,
    repeated_lines: set[str],
) -> tuple[str, dict[str, int]]:
    lines = [
        preserve_table_spacing(line)
        for line in text.splitlines()
        if normalize_line(line) not in repeated_lines
    ]
    repaired, hyphen_count = repair_hyphenation("\n".join(lines))
    repaired, split_count = repair_known_split_words(repaired)
    table_count = sum(1 for line in lines if " | " in line)
    return repaired, {
        "hyphen": hyphen_count,
        "split_word": split_count,
        "table_line": table_count,
    }


def repair_hyphenation(text: str) -> tuple[str, int]:
    repaired, count = re.subn(r"(\w)-\n(\w)", r"\1\2", text)
    return repaired, count


def repair_known_split_words(text: str) -> tuple[str, int]:
    repairs = {
        r"\bcomfort able\b": "comfortable",
        r"\bvulner able\b": "vulnerable",
        r"\bSuper nal\b": "Supernal",
        r"\bAwak ened\b": "Awakened",
        r"\bMan hattan\b": "Manhattan",
    }
    count = 0
    repaired = text
    for pattern, replacement in repairs.items():
        repaired, repairs_done = re.subn(pattern, replacement, repaired)
        count += repairs_done
    return repaired, count


def preserve_table_spacing(line: str) -> str:
    if len(re.findall(r"\S\s{2,}\S", line)) < 2:
        return line
    cells = [cell.strip() for cell in re.split(r"\s{2,}", line.strip())]
    if len(cells) < 3:
        return line
    return " | ".join(cell for cell in cells if cell)


def normalize_line(line: str) -> str:
    return re.sub(r"\s+", " ", line).strip()




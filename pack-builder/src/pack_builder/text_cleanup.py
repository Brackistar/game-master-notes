from __future__ import annotations

import re
from collections import Counter

from pack_builder.models import ExtractedDocument, ExtractedPage


def clean_documents(
    documents: list[ExtractedDocument],
) -> tuple[list[ExtractedDocument], dict[str, object]]:
    cleaned_documents: list[ExtractedDocument] = []
    removed_lines: dict[str, list[str]] = {}
    hyphen_repairs = 0

    for document in documents:
        repeated_lines = detect_repeated_lines(document.pages)
        cleaned_pages: list[ExtractedPage] = []
        for page in document.pages:
            cleaned_text, repairs = clean_page_text(page.text, repeated_lines)
            hyphen_repairs += repairs
            cleaned_pages.append(
                ExtractedPage(
                    page_number=page.page_number,
                    text=cleaned_text,
                    warnings=page.warnings,
                )
            )
        removed_lines[document.document_id] = sorted(repeated_lines)
        cleaned_documents.append(
            ExtractedDocument(
                document_id=document.document_id,
                source_path=document.source_path,
                source_filename=document.source_filename,
                source_checksum=document.source_checksum,
                page_count=document.page_count,
                pages=cleaned_pages,
            )
        )

    return cleaned_documents, {
        "enabled": True,
        "removed_repeated_lines": removed_lines,
        "hyphen_repairs": hyphen_repairs,
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


def clean_page_text(text: str, repeated_lines: set[str]) -> tuple[str, int]:
    lines = [
        line
        for line in text.splitlines()
        if normalize_line(line) not in repeated_lines
    ]
    return repair_hyphenation("\n".join(lines))


def repair_hyphenation(text: str) -> tuple[str, int]:
    repaired, count = re.subn(r"(\w)-\n(\w)", r"\1\2", text)
    return repaired, count


def normalize_line(line: str) -> str:
    return re.sub(r"\s+", " ", line).strip()

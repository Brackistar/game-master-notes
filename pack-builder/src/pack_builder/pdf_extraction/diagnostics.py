from __future__ import annotations

import re

from pack_builder.core_domain.models import ExtractedDocument, ExtractedPage
from pack_builder.ocr_detection.detector import detect_ocr_pages


def advanced_extraction_diagnostics(
    documents: list[ExtractedDocument],
) -> dict[str, object]:
    ocr_detection = detect_ocr_pages(documents)
    return {
        "ocr_detection": ocr_detection,
        "ocr_candidates": ocr_detection["ocr_candidates"],
        "table_warnings": page_refs(documents, has_table_shape),
        "multicolumn_warnings": page_refs(documents, has_multicolumn_shape),
        "merged_word_warnings": page_refs(documents, has_merged_words),
    }


def page_refs(
    documents: list[ExtractedDocument],
    predicate,
) -> list[dict[str, object]]:
    refs: list[dict[str, object]] = []
    for document in documents:
        for page in document.pages:
            if predicate(page):
                refs.append(
                    {"document_id": document.document_id, "page": page.page_number}
                )
    return refs


def has_table_shape(page: ExtractedPage) -> bool:
    table_like_lines = 0
    for line in page.text.splitlines():
        if len(re.findall(r"\s{2,}", line)) >= 2 or "|" in line:
            table_like_lines += 1
    return table_like_lines >= 3


def has_multicolumn_shape(page: ExtractedPage) -> bool:
    lines = [line for line in page.text.splitlines() if line.strip()]
    if len(lines) < 8:
        return False
    wide_gap_lines = sum(1 for line in lines if re.search(r"\S\s{8,}\S", line))
    return wide_gap_lines >= max(4, len(lines) // 3)


def has_merged_words(page: ExtractedPage) -> bool:
    words = re.findall(r"[a-z]{5,}[A-Z][a-z]{3,}", page.text)
    return len(words) >= 2




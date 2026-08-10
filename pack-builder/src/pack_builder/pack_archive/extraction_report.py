from __future__ import annotations

import hashlib

from pack_builder.core_domain.constants import SUSPICIOUS_PAGE_CHAR_THRESHOLD
from pack_builder.pdf_extraction.diagnostics import advanced_extraction_diagnostics
from pack_builder.core_domain.models import ExtractedDocument, ExtractedPage


def build_extraction_report(
    extractor_name: str,
    documents: list[ExtractedDocument],
    max_chars_per_chunk: int,
    cleanup_report: dict[str, object],
    toc_report: dict[str, object],
    chunk_quality_report: dict[str, object],
) -> dict[str, object]:
    page_report = collect_page_report(documents)
    return {
        "extractor": extractor_name,
        "chunking": {
            "max_chars_per_chunk": max_chars_per_chunk,
            "strategy": "paragraph-aware",
        },
        "cleanup": cleanup_report,
        "toc_cleanup": toc_report,
        "chunk_quality": chunk_quality_report,
        "advanced_extraction": advanced_extraction_diagnostics(documents),
        "page_text_lengths": page_report["page_text_lengths"],
        "empty_pages": page_report["empty_pages"],
        "suspicious_pages": page_report["suspicious_pages"],
        "duplicate_pages": page_report["duplicate_pages"],
        "warnings": page_report["warnings"],
        "errors": [],
    }


def collect_page_report(documents: list[ExtractedDocument]) -> dict[str, object]:
    seen_page_texts: dict[str, dict[str, object]] = {}
    report: dict[str, object] = {
        "page_text_lengths": {},
        "empty_pages": [],
        "suspicious_pages": [],
        "duplicate_pages": [],
        "warnings": [],
    }
    for document in documents:
        lengths = collect_document_pages(document, seen_page_texts, report)
        report["page_text_lengths"][document.document_id] = lengths
    return report


def collect_document_pages(
    document: ExtractedDocument,
    seen_page_texts: dict[str, dict[str, object]],
    report: dict[str, object],
) -> list[dict[str, int]]:
    lengths: list[dict[str, int]] = []
    for page in document.pages:
        text_len = len(page.text.strip())
        lengths.append({"page": page.page_number, "characters": text_len})
        record_page_stats(document, page, text_len, report)
        record_duplicate_page(document, page, seen_page_texts, report)
    return lengths


def record_page_stats(
    document: ExtractedDocument,
    page: ExtractedPage,
    text_len: int,
    report: dict[str, object],
) -> None:
    if text_len == 0:
        report["empty_pages"].append(page_location(document, page.page_number))
    elif text_len < SUSPICIOUS_PAGE_CHAR_THRESHOLD:
        report["suspicious_pages"].append(
            page_location(document, page.page_number) | {"characters": text_len}
        )
    for warning in page.warnings:
        report["warnings"].append(
            page_location(document, page.page_number) | {"warning": warning}
        )


def record_duplicate_page(
    document: ExtractedDocument,
    page: ExtractedPage,
    seen_page_texts: dict[str, dict[str, object]],
    report: dict[str, object],
) -> None:
    normalized_text = " ".join(page.text.split())
    if not normalized_text:
        return
    text_hash = hashlib.sha256(normalized_text.encode("utf-8")).hexdigest()
    first_seen = seen_page_texts.get(text_hash)
    if not first_seen:
        seen_page_texts[text_hash] = page_location(document, page.page_number)
        return
    report["duplicate_pages"].append(
        page_location(document, page.page_number)
        | {
            "matches_document_id": first_seen["document_id"],
            "matches_page": first_seen["page"],
        }
    )


def page_location(document: ExtractedDocument, page_number: int) -> dict[str, object]:
    return {"document_id": document.document_id, "page": page_number}




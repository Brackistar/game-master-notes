from __future__ import annotations

from pack_builder.core_domain.models import ExtractedDocument, ExtractedPage

LOW_TEXT_THRESHOLD = 40


def detect_ocr_pages(documents: list[ExtractedDocument]) -> dict[str, object]:
    findings = {
        "ocr_candidates": [],
        "image_only_pages": [],
        "blank_pages": [],
        "low_text_image_pages": [],
    }
    for document in documents:
        for page in document.pages:
            add_page_finding(findings, document, page)
    return findings


def add_page_finding(
    findings: dict[str, list[dict[str, object]]],
    document: ExtractedDocument,
    page: ExtractedPage,
) -> None:
    text_len = len(page.text.strip())
    visual_count = page.ocr_signals.image_count + page.ocr_signals.drawing_count
    if text_len == 0 and visual_count > 0:
        finding = page_ref(document, page, "image_without_text")
        findings["ocr_candidates"].append(finding)
        findings["image_only_pages"].append(finding)
    elif text_len == 0:
        findings["blank_pages"].append(page_ref(document, page, "no_text_or_visuals"))
    elif text_len < LOW_TEXT_THRESHOLD and visual_count > 0:
        finding = page_ref(document, page, "low_text_with_visuals")
        findings["ocr_candidates"].append(finding)
        findings["low_text_image_pages"].append(finding)


def page_ref(
    document: ExtractedDocument,
    page: ExtractedPage,
    reason: str,
) -> dict[str, object]:
    return {
        "document_id": document.document_id,
        "page": page.page_number,
        "reason": reason,
        "characters": len(page.text.strip()),
        "image_count": page.ocr_signals.image_count,
        "drawing_count": page.ocr_signals.drawing_count,
    }

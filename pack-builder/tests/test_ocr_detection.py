from __future__ import annotations

from pathlib import Path

from pack_builder.core_domain.models import (
    ExtractedDocument,
    ExtractedPage,
    PageOcrSignals,
)
from pack_builder.ocr_detection.detector import detect_ocr_pages


def document(pages: list[ExtractedPage]) -> ExtractedDocument:
    return ExtractedDocument(
        document_id="doc",
        source_path=Path("book.pdf"),
        source_filename="book.pdf",
        source_checksum="abc",
        page_count=len(pages),
        pages=pages,
    )


def test_detect_ocr_pages_classifies_image_only_pages() -> None:
    page = ExtractedPage(1, "", ocr_signals=PageOcrSignals(image_count=1))

    report = detect_ocr_pages([document([page])])

    assert report["ocr_candidates"][0]["reason"] == "image_without_text"
    assert report["image_only_pages"][0]["page"] == 1


def test_detect_ocr_pages_separates_true_blank_pages() -> None:
    page = ExtractedPage(1, "")

    report = detect_ocr_pages([document([page])])

    assert report["ocr_candidates"] == []
    assert report["blank_pages"][0]["reason"] == "no_text_or_visuals"


def test_detect_ocr_pages_flags_low_text_image_pages() -> None:
    page = ExtractedPage(
        1,
        "tiny caption",
        ocr_signals=PageOcrSignals(image_count=1),
    )

    report = detect_ocr_pages([document([page])])

    assert report["ocr_candidates"][0]["reason"] == "low_text_with_visuals"
    assert report["low_text_image_pages"][0]["characters"] == 12

from __future__ import annotations

from pathlib import Path

from pack_builder.pdf_extraction.diagnostics import advanced_extraction_diagnostics
from pack_builder.core_domain.models import ExtractedDocument, ExtractedPage


def document(pages: list[ExtractedPage]) -> ExtractedDocument:
    return ExtractedDocument(
        document_id="doc",
        source_path=Path("book.pdf"),
        source_filename="book.pdf",
        source_checksum="abc",
        page_count=len(pages),
        pages=pages,
    )


def test_advanced_diagnostics_flags_ocr_candidates() -> None:
    report = advanced_extraction_diagnostics([document([ExtractedPage(1, "")])])

    assert report["ocr_candidates"] == [{"document_id": "doc", "page": 1}]


def test_advanced_diagnostics_flags_table_and_multicolumn_shapes() -> None:
    table_text = "A  B  C\n1  2  3\n4  5  6"
    multicolumn_text = "\n".join([f"left {i}          right {i}" for i in range(8)])

    report = advanced_extraction_diagnostics(
        [document([ExtractedPage(1, table_text), ExtractedPage(2, multicolumn_text)])]
    )

    assert report["table_warnings"] == [{"document_id": "doc", "page": 1}]
    assert report["multicolumn_warnings"] == [{"document_id": "doc", "page": 2}]


def test_advanced_diagnostics_flags_merged_words() -> None:
    text = "outsideThemes theirunlocked Theythe anotherMergedWord"

    report = advanced_extraction_diagnostics([document([ExtractedPage(1, text)])])

    assert report["merged_word_warnings"] == [{"document_id": "doc", "page": 1}]




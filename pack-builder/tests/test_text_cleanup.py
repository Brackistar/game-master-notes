from __future__ import annotations

from pathlib import Path

from pack_builder.core_domain.models import ExtractedDocument, ExtractedPage
from pack_builder.content_processing.text_cleanup import clean_documents, repair_hyphenation


def test_repair_hyphenation_joins_split_words() -> None:
    text, repairs = repair_hyphenation("ancient mag-\nic wards")

    assert text == "ancient magic wards"
    assert repairs == 1


def test_clean_documents_removes_repeated_lines() -> None:
    document = ExtractedDocument(
        document_id="doc",
        source_path=Path("book.pdf"),
        source_filename="book.pdf",
        source_checksum="abc",
        page_count=3,
        pages=[
            ExtractedPage(1, "Book Title\nFirst page text"),
            ExtractedPage(2, "Book Title\nSecond page text"),
            ExtractedPage(3, "Book Title\nThird page text"),
        ],
    )

    cleaned, report = clean_documents([document])

    assert "Book Title" not in cleaned[0].pages[0].text
    assert report["removed_repeated_lines"]["doc"] == ["Book Title"]




from __future__ import annotations

from pathlib import Path

from pack_builder.content_processing.front_matter_cleanup import (
    is_front_matter_page,
    remove_front_matter_pages,
)
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


def test_is_front_matter_page_detects_credits_and_legal_text() -> None:
    text = "Credits\nWriters: One Person\nAll rights reserved."

    assert is_front_matter_page(text)


def test_remove_front_matter_pages_respects_max_page() -> None:
    front = ExtractedPage(3, "Credits\nWriters: One Person")
    late_front = ExtractedPage(10, "Credits\nWriters: Appendix Team")
    normal = ExtractedPage(4, "Character creation starts here.")

    cleaned, report = remove_front_matter_pages(
        [document([front, normal, late_front])],
        max_page=3,
    )

    assert [page.page_number for page in cleaned[0].pages] == [4, 10]
    assert report["removed_pages"]["doc"] == [3]

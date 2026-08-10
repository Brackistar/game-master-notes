from __future__ import annotations

from pathlib import Path

from pack_builder.models import ExtractedDocument, ExtractedPage
from pack_builder.toc_cleanup import is_toc_page, remove_toc_pages


def document(pages: list[ExtractedPage]) -> ExtractedDocument:
    return ExtractedDocument(
        document_id="doc",
        source_path=Path("book.pdf"),
        source_filename="book.pdf",
        source_checksum="abc",
        page_count=len(pages),
        pages=pages,
    )


def test_is_toc_page_detects_table_of_contents_label() -> None:
    assert is_toc_page("TABLE OF CONTENTS\nIntroduction 11")


def test_is_toc_page_detects_dense_title_page_number_lines() -> None:
    text = "\n".join(
        [
            "Introduction 11",
            "Themes 12",
            "Chapter One 20",
            "Chapter Two 40",
            "Spellcasting 111",
        ]
    )

    assert is_toc_page(text)


def test_is_toc_page_detects_dense_inline_extracted_toc() -> None:
    text = (
        "Introduction 11 Concepts 42 Chapter Three: Supernal Lore 79 "
        "Stereotypes 42 Themes 11 How to Use This Book 12 Mysteries 45 "
        "Character Creation 79 Step One: Character Concept 79 Skills 80 "
        "Advantages 81 The Awakening 81"
    )

    assert is_toc_page(text)


def test_remove_toc_pages_respects_max_page() -> None:
    toc = ExtractedPage(4, "TABLE OF CONTENTS\nIntroduction 11")
    late_toc = ExtractedPage(30, "TABLE OF CONTENTS\nIndex 300")
    normal = ExtractedPage(5, "A normal page of prose.")

    cleaned, report = remove_toc_pages([document([toc, normal, late_toc])], max_page=20)

    assert [page.page_number for page in cleaned[0].pages] == [5, 30]
    assert report["removed_pages"]["doc"] == [4]

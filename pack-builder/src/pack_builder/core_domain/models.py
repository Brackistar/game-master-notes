from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path


@dataclass(frozen=True)
class ExtractedPage:
    page_number: int
    text: str
    warnings: list[str] = field(default_factory=list)


@dataclass(frozen=True)
class ExtractedDocument:
    document_id: str
    source_path: Path
    source_filename: str
    source_checksum: str
    page_count: int
    pages: list[ExtractedPage]


def replace_document_pages(
    document: ExtractedDocument,
    pages: list[ExtractedPage],
) -> ExtractedDocument:
    return ExtractedDocument(
        document_id=document.document_id,
        source_path=document.source_path,
        source_filename=document.source_filename,
        source_checksum=document.source_checksum,
        page_count=document.page_count,
        pages=pages,
    )


@dataclass(frozen=True)
class SourceChunk:
    chunk_id: str
    document_id: str
    page_start: int
    page_end: int
    citation_label: str
    text: str
    char_count: int
    embedding_row_index: int

    def to_json(self) -> dict[str, object]:
        return {
            "chunk_id": self.chunk_id,
            "document_id": self.document_id,
            "page_start": self.page_start,
            "page_end": self.page_end,
            "citation_label": self.citation_label,
            "text": self.text,
            "char_count": self.char_count,
            "embedding_row_index": self.embedding_row_index,
        }




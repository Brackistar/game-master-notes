from __future__ import annotations

import hashlib
import re
from dataclasses import dataclass

from pack_builder.constants import DEFAULT_MAX_CHARS_PER_CHUNK
from pack_builder.models import ExtractedPage, SourceChunk


@dataclass(frozen=True)
class Paragraph:
    text: str
    page_start: int
    page_end: int


def normalize_text(text: str) -> str:
    text = text.replace("\r\n", "\n").replace("\r", "\n").replace("\f", "\n")
    text = text.replace("\t", " ")
    lines = [re.sub(r"[ ]+", " ", line).strip() for line in text.split("\n")]

    normalized_lines: list[str] = []
    blank_seen = False
    for line in lines:
        if not line:
            if normalized_lines and not blank_seen:
                normalized_lines.append("")
            blank_seen = True
            continue
        normalized_lines.append(line)
        blank_seen = False

    return "\n".join(normalized_lines).strip()


def split_page_paragraphs(page: ExtractedPage) -> list[Paragraph]:
    normalized = normalize_text(page.text)
    if not normalized:
        return []

    raw_paragraphs = re.split(r"\n\s*\n", normalized)
    paragraphs: list[Paragraph] = []
    for raw in raw_paragraphs:
        paragraph = " ".join(line.strip() for line in raw.splitlines() if line.strip())
        paragraph = re.sub(r"[ ]+", " ", paragraph).strip()
        if paragraph:
            paragraphs.append(
                Paragraph(
                    text=paragraph,
                    page_start=page.page_number,
                    page_end=page.page_number,
                )
            )
    return paragraphs


def split_long_paragraph(paragraph: Paragraph, max_chars: int) -> list[Paragraph]:
    if len(paragraph.text) <= max_chars:
        return [paragraph]

    parts: list[Paragraph] = []
    remaining = paragraph.text
    while len(remaining) > max_chars:
        boundary = remaining.rfind(" ", 0, max_chars)
        if boundary < max_chars // 2:
            boundary = max_chars
        parts.append(
            Paragraph(
                text=remaining[:boundary].strip(),
                page_start=paragraph.page_start,
                page_end=paragraph.page_end,
            )
        )
        remaining = remaining[boundary:].strip()
    if remaining:
        parts.append(
            Paragraph(
                text=remaining,
                page_start=paragraph.page_start,
                page_end=paragraph.page_end,
            )
        )
    return parts


def citation_label(document_id: str, page_start: int, page_end: int) -> str:
    if page_start == page_end:
        return f"{document_id} p. {page_start}"
    return f"{document_id} pp. {page_start}-{page_end}"


def deterministic_chunk_id(
    pack_id: str,
    document_id: str,
    page_start: int,
    page_end: int,
    chunk_index: int,
) -> str:
    source = f"{pack_id}|{document_id}|{page_start}|{page_end}|{chunk_index}"
    return f"chunk-{hashlib.sha256(source.encode('utf-8')).hexdigest()[:16]}"


def chunk_pages(
    *,
    pack_id: str,
    document_id: str,
    pages: list[ExtractedPage],
    max_chars: int = DEFAULT_MAX_CHARS_PER_CHUNK,
) -> list[SourceChunk]:
    paragraphs: list[Paragraph] = []
    for page in pages:
        for paragraph in split_page_paragraphs(page):
            paragraphs.extend(split_long_paragraph(paragraph, max_chars))

    chunks: list[SourceChunk] = []
    current_texts: list[str] = []
    current_start = 0
    current_end = 0

    def flush() -> None:
        nonlocal current_texts, current_start, current_end
        if not current_texts:
            return
        text = "\n\n".join(current_texts)
        index = len(chunks)
        chunks.append(
            SourceChunk(
                chunk_id=deterministic_chunk_id(
                    pack_id, document_id, current_start, current_end, index
                ),
                document_id=document_id,
                page_start=current_start,
                page_end=current_end,
                citation_label=citation_label(document_id, current_start, current_end),
                text=text,
                char_count=len(text),
                embedding_row_index=index,
            )
        )
        current_texts = []
        current_start = 0
        current_end = 0

    for paragraph in paragraphs:
        proposed_len = len(paragraph.text)
        if current_texts:
            proposed_len += sum(len(text) for text in current_texts)
            proposed_len += 2 * len(current_texts)

        if current_texts and proposed_len > max_chars:
            flush()

        if not current_texts:
            current_start = paragraph.page_start
        current_end = paragraph.page_end
        current_texts.append(paragraph.text)

    flush()
    return chunks

from __future__ import annotations

import hashlib

from pack_builder.core_domain.models import SourceChunk


def improve_chunks(
    chunks: list[SourceChunk],
    *,
    overlap_chars: int,
    deduplicate: bool,
) -> tuple[list[SourceChunk], dict[str, object]]:
    improved = apply_chunk_overlap(chunks, overlap_chars)
    deduped, duplicate_ids = deduplicate_chunks(improved) if deduplicate else (improved, [])
    return reindex_chunk_rows(deduped), {
        "overlap_chars": overlap_chars,
        "deduplicate": deduplicate,
        "removed_duplicate_chunk_ids": duplicate_ids,
    }


def apply_chunk_overlap(chunks: list[SourceChunk], overlap_chars: int) -> list[SourceChunk]:
    if overlap_chars <= 0:
        return chunks

    overlapped: list[SourceChunk] = []
    previous_text = ""
    for chunk in chunks:
        prefix = tail_text(previous_text, overlap_chars)
        text = f"{prefix}\n\n{chunk.text}" if prefix else chunk.text
        overlapped.append(replace_chunk_text(chunk, text))
        previous_text = chunk.text
    return overlapped


def deduplicate_chunks(chunks: list[SourceChunk]) -> tuple[list[SourceChunk], list[str]]:
    seen: set[str] = set()
    kept: list[SourceChunk] = []
    duplicate_ids: list[str] = []
    for chunk in chunks:
        fingerprint = text_fingerprint(chunk.text)
        if fingerprint in seen:
            duplicate_ids.append(chunk.chunk_id)
            continue
        seen.add(fingerprint)
        kept.append(chunk)
    return kept, duplicate_ids


def reindex_chunk_rows(chunks: list[SourceChunk]) -> list[SourceChunk]:
    return [
        SourceChunk(
            chunk_id=chunk.chunk_id,
            document_id=chunk.document_id,
            page_start=chunk.page_start,
            page_end=chunk.page_end,
            citation_label=chunk.citation_label,
            text=chunk.text,
            char_count=len(chunk.text),
            embedding_row_index=index,
        )
        for index, chunk in enumerate(chunks)
    ]


def replace_chunk_text(chunk: SourceChunk, text: str) -> SourceChunk:
    return SourceChunk(
        chunk_id=chunk.chunk_id,
        document_id=chunk.document_id,
        page_start=chunk.page_start,
        page_end=chunk.page_end,
        citation_label=chunk.citation_label,
        text=text,
        char_count=len(text),
        embedding_row_index=chunk.embedding_row_index,
    )


def tail_text(text: str, max_chars: int) -> str:
    if len(text) <= max_chars:
        return text
    boundary = text.rfind(" ", 0, max_chars)
    if boundary < max_chars // 2:
        return text[-max_chars:]
    return text[-boundary:].strip()


def text_fingerprint(text: str) -> str:
    normalized = " ".join(text.lower().split())
    return hashlib.sha256(normalized.encode("utf-8")).hexdigest()




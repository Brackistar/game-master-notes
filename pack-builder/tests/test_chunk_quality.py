from __future__ import annotations

from pack_builder.chunk_quality import improve_chunks
from pack_builder.models import SourceChunk


def chunk(chunk_id: str, text: str, row: int) -> SourceChunk:
    return SourceChunk(
        chunk_id=chunk_id,
        document_id="doc",
        page_start=1,
        page_end=1,
        citation_label="doc p. 1",
        text=text,
        char_count=len(text),
        embedding_row_index=row,
    )


def test_improve_chunks_adds_overlap_text() -> None:
    chunks, report = improve_chunks(
        [
            chunk("chunk-1", "alpha beta gamma", 0),
            chunk("chunk-2", "delta epsilon", 1),
        ],
        overlap_chars=10,
        deduplicate=True,
    )

    assert chunks[1].text.startswith("gamma\n\ndelta")
    assert report["overlap_chars"] == 10


def test_improve_chunks_removes_duplicate_text_and_reindexes() -> None:
    chunks, report = improve_chunks(
        [
            chunk("chunk-1", "Same Text", 0),
            chunk("chunk-2", "same text", 1),
            chunk("chunk-3", "Different", 2),
        ],
        overlap_chars=0,
        deduplicate=True,
    )

    assert [item.chunk_id for item in chunks] == ["chunk-1", "chunk-3"]
    assert [item.embedding_row_index for item in chunks] == [0, 1]
    assert report["removed_duplicate_chunk_ids"] == ["chunk-2"]

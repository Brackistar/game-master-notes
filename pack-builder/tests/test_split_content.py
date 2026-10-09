from __future__ import annotations

from pathlib import Path

import numpy as np
import pytest

from pack_builder.content_processing.content_classification import ContentLabel
from pack_builder.content_processing.pipeline import prepare_split_pack_content
from pack_builder.core_domain.models import ExtractedDocument, ExtractedPage
from pack_builder.embedding_generation.embeddings import DeterministicEmbeddingProvider
from pack_builder.pack_archive.reader import read_chunks, read_extraction_report
from pack_builder.pack_archive.validate import validate_pack
from pack_builder.pack_archive.writer import build_split_packs, split_output_paths


class SyntheticExtractor:
    name = "synthetic"

    def extract(self, pdf_path: Path):  # pragma: no cover - extraction is patched
        raise AssertionError("unexpected extraction call")


class FailingSecondEmbeddingProvider:
    model_id = "failing-test"
    model_revision = "failing-v1"
    dimensions = 4

    def __init__(self) -> None:
        self.calls = 0

    def encode(self, texts: list[str]) -> np.ndarray:
        self.calls += 1
        if self.calls == 2:
            return np.zeros((0, self.dimensions), dtype=np.float32)
        embeddings = np.zeros((len(texts), self.dimensions), dtype=np.float32)
        embeddings[:, 0] = 1.0
        return embeddings


def long_paragraph(seed: str, target: int = 1000) -> str:
    return (seed + " ") * (target // (len(seed) + 1) + 1)


def synthetic_document(path: Path) -> ExtractedDocument:
    rules_one = long_paragraph("Roll 1d20; the target must make a saving throw and takes 2d6 damage.")
    rules_two = long_paragraph("Armor Class: 16; Hit Points: 45; Speed: 30 ft.")
    narrative_one = long_paragraph("The riders arrived at the ancient village beside the forgotten river.")
    narrative_two = long_paragraph("Mara whispered as they traveled through the forest and discovered the ruins.")
    ambiguous = "A quiet road crossed the valley toward the distant walls."
    return ExtractedDocument(
        document_id="synthetic-book",
        source_path=path,
        source_filename=path.name,
        source_checksum="abc123",
        page_count=2,
        pages=[
            ExtractedPage(1, f"{rules_one}\n\n{narrative_one}\n\n{ambiguous}"),
            ExtractedPage(2, f"{rules_two}\n\n{narrative_two}"),
        ],
    )


def patch_extraction(monkeypatch: pytest.MonkeyPatch, pdf_path: Path) -> None:
    monkeypatch.setattr(
        "pack_builder.content_processing.pipeline.extract_document",
        lambda path, extractor: synthetic_document(pdf_path),
    )


def split_kwargs(pdf_path: Path, out_path: Path) -> dict[str, object]:
    return {
        "pdf_paths": [pdf_path],
        "out_path": out_path,
        "title": "Synthetic Book",
        "system": "Test System",
        "edition": "1e",
        "language": "en",
        "extractor": SyntheticExtractor(),
        "max_chars_per_chunk": 1800,
        "clean_text": False,
        "remove_front_matter": False,
        "front_matter_max_page": 3,
        "remove_toc_pages": False,
        "toc_max_page": 20,
        "deduplicate_chunks": True,
        "chunk_overlap_chars": 0,
    }


def test_split_pipeline_qualifies_both_categories_and_duplicates_ambiguous(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    pdf_path = tmp_path / "book.pdf"
    pdf_path.write_bytes(b"synthetic")
    patch_extraction(monkeypatch, pdf_path)

    kwargs = split_kwargs(pdf_path, tmp_path / "book.gmnpack")
    kwargs.pop("out_path")
    kwargs["deduplicate"] = kwargs.pop("deduplicate_chunks")
    prepared = prepare_split_pack_content(**kwargs)

    assert prepared.category_decisions["ruleset"]["qualifies"] is True
    assert prepared.category_decisions["narrative"]["qualifies"] is True
    assert prepared.classification_report["label_counts"]["ambiguous"] == 1
    assert any("quiet road" in chunk.text for chunk in prepared.categories[ContentLabel.RULESET].chunks)
    assert any("quiet road" in chunk.text for chunk in prepared.categories[ContentLabel.NARRATIVE].chunks)
    assert all(chunk.page_start <= chunk.page_end for content in prepared.categories.values() for chunk in content.chunks)


def test_split_writer_creates_valid_independent_packs(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    pdf_path = tmp_path / "book.pdf"
    pdf_path.write_bytes(b"synthetic")
    patch_extraction(monkeypatch, pdf_path)
    base = tmp_path / "book.gmnpack"

    result = build_split_packs(
        **split_kwargs(pdf_path, base),
        embedding_provider=DeterministicEmbeddingProvider(),
    )

    paths = split_output_paths(base)
    assert set(result.results) == {"ruleset", "narrative"}
    for label, path in paths.items():
        assert path.exists()
        assert validate_pack(path).ok
        assert result.results[label.value].manifest["title"].endswith(label.value.title())
        report = read_extraction_report(path)
        assert report["content_category"] == label.value
        assert report["content_classification"]["classifier_version"] == "rules-narrative-v1"
    rules_text = " ".join(chunk["text"] for chunk in read_chunks(paths[ContentLabel.RULESET]))
    narrative_text = " ".join(chunk["text"] for chunk in read_chunks(paths[ContentLabel.NARRATIVE]))
    assert "saving throw" in rules_text
    assert "ancient village" not in rules_text
    assert "ancient village" in narrative_text
    assert "saving throw" not in narrative_text


def test_split_writer_preflights_all_outputs_before_writing(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    pdf_path = tmp_path / "book.pdf"
    pdf_path.write_bytes(b"synthetic")
    patch_extraction(monkeypatch, pdf_path)
    base = tmp_path / "book.gmnpack"
    paths = split_output_paths(base)
    paths[ContentLabel.NARRATIVE].write_bytes(b"existing")

    with pytest.raises(FileExistsError, match="split output already exists"):
        build_split_packs(
            **split_kwargs(pdf_path, base),
            embedding_provider=DeterministicEmbeddingProvider(),
        )

    assert not paths[ContentLabel.RULESET].exists()
    assert paths[ContentLabel.NARRATIVE].read_bytes() == b"existing"


def test_split_writer_force_replaces_both_outputs(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    pdf_path = tmp_path / "book.pdf"
    pdf_path.write_bytes(b"synthetic")
    patch_extraction(monkeypatch, pdf_path)
    base = tmp_path / "book.gmnpack"
    paths = split_output_paths(base)
    for path in paths.values():
        path.write_bytes(b"existing")

    build_split_packs(
        **split_kwargs(pdf_path, base),
        embedding_provider=DeterministicEmbeddingProvider(),
        force=True,
    )

    assert all(validate_pack(path).ok for path in paths.values())


def test_split_writer_leaves_no_final_output_when_second_category_fails(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    pdf_path = tmp_path / "book.pdf"
    pdf_path.write_bytes(b"synthetic")
    patch_extraction(monkeypatch, pdf_path)
    base = tmp_path / "book.gmnpack"
    paths = split_output_paths(base)

    with pytest.raises(ValueError, match="embedding provider returned shape"):
        build_split_packs(
            **split_kwargs(pdf_path, base),
            embedding_provider=FailingSecondEmbeddingProvider(),
        )

    assert all(not path.exists() for path in paths.values())
    assert not list(tmp_path.glob("*.tmp.gmnpack"))

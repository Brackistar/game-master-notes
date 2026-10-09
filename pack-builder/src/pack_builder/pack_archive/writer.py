from __future__ import annotations

import io
import json
import os
import tempfile
from time import perf_counter
import zipfile
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path

import numpy as np

from pack_builder.core_domain.constants import (
    CHUNKS_FILE,
    DEFAULT_MAX_CHARS_PER_CHUNK,
    DOCUMENTS_FILE,
    EMBEDDINGS_FILE,
    EXTRACTION_REPORT_FILE,
    GENERATOR_VERSION,
    MANIFEST_FILE,
    PACK_SCHEMA_VERSION,
    SUSPICIOUS_PAGE_CHAR_THRESHOLD,
)
from pack_builder.embedding_generation.embeddings import EmbeddingProvider
from pack_builder.pack_archive.extraction_report import build_extraction_report
from pack_builder.core_domain.models import ExtractedDocument, SourceChunk
from pack_builder.content_processing.content_classification import ContentLabel
from pack_builder.content_processing.pipeline import (
    PreparedPackContent,
    prepare_pack_content,
    prepare_split_pack_content,
)
from pack_builder.pdf_extraction.extract import PdfExtractor
from pack_builder.pack_archive.validate import validate_pack


@dataclass(frozen=True)
class BuildResult:
    manifest: dict[str, object]
    documents: list[dict[str, object]]
    chunks: list[SourceChunk]
    extraction_report: dict[str, object]


@dataclass(frozen=True)
class SplitBuildResult:
    results: dict[str, BuildResult]
    decisions: dict[str, dict[str, object]]
    extraction_report: dict[str, object]

    @property
    def success(self) -> bool:
        return bool(self.results) or any(
            bool(decision.get("qualifies")) for decision in self.decisions.values()
        )


def document_metadata(document: ExtractedDocument) -> dict[str, object]:
    text_lengths = [len(page.text.strip()) for page in document.pages]
    return {
        "document_id": document.document_id,
        "source_filename": document.source_filename,
        "source_checksum": document.source_checksum,
        "page_count": document.page_count,
        "extraction_stats": {
            "empty_page_count": sum(1 for length in text_lengths if length == 0),
            "suspicious_page_count": sum(
                1
                for length in text_lengths
                if 0 < length < SUSPICIOUS_PAGE_CHAR_THRESHOLD
            ),
            "total_characters": sum(text_lengths),
        },
    }


def manifest_data(
    *,
    pack_id: str,
    title: str,
    system: str,
    edition: str,
    language: str,
    documents: list[ExtractedDocument],
    extractor_name: str,
    embedding_provider: EmbeddingProvider,
    chunk_count: int,
    build_options: dict[str, object],
) -> dict[str, object]:
    return {
        "schema_version": PACK_SCHEMA_VERSION,
        "pack_id": pack_id,
        "title": title,
        "system": system,
        "edition": edition,
        "language": language,
        "source_pdfs": [
            {
                "filename": document.source_filename,
                "checksum": document.source_checksum,
            }
            for document in documents
        ],
        "generator_version": GENERATOR_VERSION,
        "extractor_name": extractor_name,
        "embedding_model_id": embedding_provider.model_id,
        "embedding_model_revision": embedding_provider.model_revision,
        "embedding_dimensions": embedding_provider.dimensions,
        "chunk_count": chunk_count,
        "build_options": build_options,
        "created_at": datetime.now(timezone.utc).replace(microsecond=0).isoformat(),
    }


def write_pack_archive(
    *,
    out_path: Path,
    manifest: dict[str, object],
    documents: list[dict[str, object]],
    chunks: list[SourceChunk],
    embeddings: np.ndarray,
    extraction_report: dict[str, object],
) -> None:
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(out_path, mode="w", compression=zipfile.ZIP_DEFLATED) as archive:
        archive.writestr(MANIFEST_FILE, json.dumps(manifest, indent=2) + "\n")
        archive.writestr(
            DOCUMENTS_FILE, json.dumps({"documents": documents}, indent=2) + "\n"
        )
        archive.writestr(
            CHUNKS_FILE,
            "".join(json.dumps(chunk.to_json()) + "\n" for chunk in chunks),
        )
        embedding_buffer = io.BytesIO()
        np.save(embedding_buffer, embeddings.astype(np.float32), allow_pickle=False)
        archive.writestr(EMBEDDINGS_FILE, embedding_buffer.getvalue())
        archive.writestr(
            EXTRACTION_REPORT_FILE, json.dumps(extraction_report, indent=2) + "\n"
        )


def build_pack(
    *,
    pdf_paths: list[Path],
    out_path: Path,
    title: str,
    system: str,
    edition: str,
    language: str,
    extractor: PdfExtractor,
    embedding_provider: EmbeddingProvider,
    max_chars_per_chunk: int = DEFAULT_MAX_CHARS_PER_CHUNK,
    clean_text: bool = True,
    remove_front_matter: bool = True,
    front_matter_max_page: int = 3,
    remove_toc_pages: bool = True,
    toc_max_page: int = 20,
    deduplicate_chunks: bool = True,
    chunk_overlap_chars: int = 0,
) -> BuildResult:
    started_at = perf_counter()
    pack_content = prepare_pack_content(
        pdf_paths=pdf_paths,
        title=title,
        system=system,
        edition=edition,
        extractor=extractor,
        max_chars_per_chunk=max_chars_per_chunk,
        clean_text=clean_text,
        remove_front_matter=remove_front_matter,
        front_matter_max_page=front_matter_max_page,
        remove_toc_pages=remove_toc_pages,
        toc_max_page=toc_max_page,
        deduplicate=deduplicate_chunks,
        chunk_overlap_chars=chunk_overlap_chars,
    )
    if not pack_content.chunks:
        raise ValueError("no text chunks were created from the provided PDFs")
    prepared_at = perf_counter()

    embeddings = embedding_provider.encode([chunk.text for chunk in pack_content.chunks])
    embedded_at = perf_counter()
    if embeddings.shape != (len(pack_content.chunks), embedding_provider.dimensions):
        raise ValueError(
            "embedding provider returned shape "
            f"{embeddings.shape}, expected "
            f"{(len(pack_content.chunks), embedding_provider.dimensions)}"
        )

    manifest = manifest_data(
        pack_id=pack_content.pack_id,
        title=title,
        system=system,
        edition=edition,
        language=language,
        documents=pack_content.documents,
        extractor_name=extractor.name,
        embedding_provider=embedding_provider,
        chunk_count=len(pack_content.chunks),
        build_options=build_options_data(
            extractor.name,
            max_chars_per_chunk,
            chunk_overlap_chars,
            clean_text,
            remove_front_matter,
            front_matter_max_page,
            remove_toc_pages,
            toc_max_page,
            deduplicate_chunks,
        ),
    )
    document_rows = [document_metadata(document) for document in pack_content.documents]
    report = build_extraction_report(
        extractor.name,
        pack_content.documents,
        max_chars_per_chunk,
        pack_content.cleanup_report,
        pack_content.front_matter_report,
        pack_content.toc_report,
        pack_content.chunk_quality_report,
        build_timing_report(started_at, prepared_at, embedded_at),
    )

    write_pack_archive(
        out_path=out_path,
        manifest=manifest,
        documents=document_rows,
        chunks=pack_content.chunks,
        embeddings=embeddings,
        extraction_report=report,
    )
    return BuildResult(
        manifest=manifest,
        documents=document_rows,
        chunks=pack_content.chunks,
        extraction_report=report,
    )


def preview_pack(
    *,
    pdf_paths: list[Path],
    title: str,
    system: str,
    edition: str,
    language: str,
    extractor: PdfExtractor,
    max_chars_per_chunk: int = DEFAULT_MAX_CHARS_PER_CHUNK,
    clean_text: bool = True,
    remove_front_matter: bool = True,
    front_matter_max_page: int = 3,
    remove_toc_pages: bool = True,
    toc_max_page: int = 20,
    deduplicate_chunks: bool = True,
    chunk_overlap_chars: int = 0,
) -> BuildResult:
    started_at = perf_counter()
    pack_content = prepare_pack_content(
        pdf_paths=pdf_paths,
        title=title,
        system=system,
        edition=edition,
        extractor=extractor,
        max_chars_per_chunk=max_chars_per_chunk,
        clean_text=clean_text,
        remove_front_matter=remove_front_matter,
        front_matter_max_page=front_matter_max_page,
        remove_toc_pages=remove_toc_pages,
        toc_max_page=toc_max_page,
        deduplicate=deduplicate_chunks,
        chunk_overlap_chars=chunk_overlap_chars,
    )
    report = build_extraction_report(
        extractor.name,
        pack_content.documents,
        max_chars_per_chunk,
        pack_content.cleanup_report,
        pack_content.front_matter_report,
        pack_content.toc_report,
        pack_content.chunk_quality_report,
        build_timing_report(started_at, perf_counter(), None),
    )
    manifest = {
        "schema_version": PACK_SCHEMA_VERSION,
        "pack_id": pack_content.pack_id,
        "title": title,
        "system": system,
        "edition": edition,
        "language": language,
        "source_pdfs": [
            {"filename": document.source_filename, "checksum": document.source_checksum}
            for document in pack_content.documents
        ],
        "generator_version": GENERATOR_VERSION,
        "extractor_name": extractor.name,
        "embedding_model_id": None,
        "embedding_model_revision": None,
        "embedding_dimensions": None,
        "chunk_count": len(pack_content.chunks),
        "build_options": build_options_data(
            extractor.name,
            max_chars_per_chunk,
            chunk_overlap_chars,
            clean_text,
            remove_front_matter,
            front_matter_max_page,
            remove_toc_pages,
            toc_max_page,
            deduplicate_chunks,
        ),
        "created_at": None,
        "dry_run": True,
    }
    return BuildResult(
        manifest=manifest,
        documents=[document_metadata(document) for document in pack_content.documents],
        chunks=pack_content.chunks,
        extraction_report=report,
    )


def build_options_data(
    extractor_name: str,
    max_chars_per_chunk: int,
    chunk_overlap_chars: int,
    clean_text: bool,
    remove_front_matter: bool,
    front_matter_max_page: int,
    remove_toc_pages: bool,
    toc_max_page: int,
    deduplicate_chunks: bool,
) -> dict[str, object]:
    return {
        "extractor": extractor_name,
        "max_chars_per_chunk": max_chars_per_chunk,
        "chunk_overlap_chars": chunk_overlap_chars,
        "clean_text": clean_text,
        "remove_front_matter": remove_front_matter,
        "front_matter_max_page": front_matter_max_page,
        "remove_toc_pages": remove_toc_pages,
        "toc_max_page": toc_max_page,
        "deduplicate_chunks": deduplicate_chunks,
    }


def build_timing_report(
    started_at: float,
    prepared_at: float,
    embedded_at: float | None,
) -> dict[str, object]:
    measured_at = perf_counter()
    timing = {
        "prepare_seconds": round(prepared_at - started_at, 3),
        "measured_before_archive_seconds": round(measured_at - started_at, 3),
    }
    if embedded_at is not None:
        timing["embedding_seconds"] = round(embedded_at - prepared_at, 3)
        timing["post_embedding_seconds"] = round(measured_at - embedded_at, 3)
    return timing


def split_output_paths(out_path: Path) -> dict[ContentLabel, Path]:
    return {
        ContentLabel.RULESET: out_path.with_name(f"{out_path.stem}-ruleset.gmnpack"),
        ContentLabel.NARRATIVE: out_path.with_name(f"{out_path.stem}-narrative.gmnpack"),
    }


def build_split_packs(
    *,
    pdf_paths: list[Path],
    out_path: Path,
    title: str,
    system: str,
    edition: str,
    language: str,
    extractor: PdfExtractor,
    embedding_provider: EmbeddingProvider,
    max_chars_per_chunk: int = DEFAULT_MAX_CHARS_PER_CHUNK,
    clean_text: bool = True,
    remove_front_matter: bool = True,
    front_matter_max_page: int = 3,
    remove_toc_pages: bool = True,
    toc_max_page: int = 20,
    deduplicate_chunks: bool = True,
    chunk_overlap_chars: int = 0,
    force: bool = False,
) -> SplitBuildResult:
    started_at = perf_counter()
    split_content = prepare_split_pack_content(
        pdf_paths=pdf_paths,
        title=title,
        system=system,
        edition=edition,
        language=language,
        extractor=extractor,
        max_chars_per_chunk=max_chars_per_chunk,
        clean_text=clean_text,
        remove_front_matter=remove_front_matter,
        front_matter_max_page=front_matter_max_page,
        remove_toc_pages=remove_toc_pages,
        toc_max_page=toc_max_page,
        deduplicate=deduplicate_chunks,
        chunk_overlap_chars=chunk_overlap_chars,
    )
    prepared_at = perf_counter()
    output_paths = split_output_paths(out_path)
    decisions = attach_output_paths(split_content.category_decisions, output_paths)
    qualifying = [label for label in output_paths if decisions[label.value]["qualifies"]]
    preflight_split_outputs([output_paths[label] for label in qualifying], force=force)
    if not qualifying:
        return SplitBuildResult({}, decisions, aggregate_split_report(split_content.classification_report, decisions))

    temporary_paths: dict[ContentLabel, Path] = {}
    results: dict[str, BuildResult] = {}
    try:
        for label in qualifying:
            final_path = output_paths[label]
            final_path.parent.mkdir(parents=True, exist_ok=True)
            file_descriptor, temp_name = tempfile.mkstemp(
                prefix=f".{final_path.stem}-",
                suffix=".tmp.gmnpack",
                dir=final_path.parent,
            )
            os.close(file_descriptor)
            temp_path = Path(temp_name)
            temporary_paths[label] = temp_path
            suffix = "Ruleset" if label is ContentLabel.RULESET else "Narrative"
            result = build_prepared_pack(
                pack_content=split_content.categories[label],
                out_path=temp_path,
                title=f"{title} - {suffix}",
                system=system,
                edition=edition,
                language=language,
                extractor=extractor,
                embedding_provider=embedding_provider,
                max_chars_per_chunk=max_chars_per_chunk,
                chunk_overlap_chars=chunk_overlap_chars,
                clean_text=clean_text,
                remove_front_matter=remove_front_matter,
                front_matter_max_page=front_matter_max_page,
                remove_toc_pages=remove_toc_pages,
                toc_max_page=toc_max_page,
                deduplicate_chunks=deduplicate_chunks,
                started_at=started_at,
                prepared_at=prepared_at,
                classification_report=split_content.classification_report,
                category=label.value,
            )
            validation = validate_pack(temp_path)
            if not validation.ok:
                raise ValueError(
                    f"generated {label.value} pack failed validation: "
                    + "; ".join(validation.errors)
                )
            results[label.value] = result
        commit_split_archives(temporary_paths, output_paths, force=force)
    finally:
        for temp_path in temporary_paths.values():
            temp_path.unlink(missing_ok=True)

    return SplitBuildResult(
        results,
        decisions,
        aggregate_split_report(split_content.classification_report, decisions, results),
    )


def preview_split_packs(
    *,
    pdf_paths: list[Path],
    out_path: Path,
    title: str,
    system: str,
    edition: str,
    language: str,
    extractor: PdfExtractor,
    max_chars_per_chunk: int = DEFAULT_MAX_CHARS_PER_CHUNK,
    clean_text: bool = True,
    remove_front_matter: bool = True,
    front_matter_max_page: int = 3,
    remove_toc_pages: bool = True,
    toc_max_page: int = 20,
    deduplicate_chunks: bool = True,
    chunk_overlap_chars: int = 0,
) -> SplitBuildResult:
    split_content = prepare_split_pack_content(
        pdf_paths=pdf_paths,
        title=title,
        system=system,
        edition=edition,
        language=language,
        extractor=extractor,
        max_chars_per_chunk=max_chars_per_chunk,
        clean_text=clean_text,
        remove_front_matter=remove_front_matter,
        front_matter_max_page=front_matter_max_page,
        remove_toc_pages=remove_toc_pages,
        toc_max_page=toc_max_page,
        deduplicate=deduplicate_chunks,
        chunk_overlap_chars=chunk_overlap_chars,
    )
    decisions = attach_output_paths(split_content.category_decisions, split_output_paths(out_path))
    return SplitBuildResult({}, decisions, aggregate_split_report(split_content.classification_report, decisions))


def build_prepared_pack(
    *,
    pack_content: PreparedPackContent,
    out_path: Path,
    title: str,
    system: str,
    edition: str,
    language: str,
    extractor: PdfExtractor,
    embedding_provider: EmbeddingProvider,
    max_chars_per_chunk: int,
    chunk_overlap_chars: int,
    clean_text: bool,
    remove_front_matter: bool,
    front_matter_max_page: int,
    remove_toc_pages: bool,
    toc_max_page: int,
    deduplicate_chunks: bool,
    started_at: float,
    prepared_at: float,
    classification_report: dict[str, object],
    category: str,
) -> BuildResult:
    if not pack_content.chunks:
        raise ValueError(f"no text chunks were created for {category}")
    embeddings = embedding_provider.encode([chunk.text for chunk in pack_content.chunks])
    embedded_at = perf_counter()
    expected_shape = (len(pack_content.chunks), embedding_provider.dimensions)
    if embeddings.shape != expected_shape:
        raise ValueError(f"embedding provider returned shape {embeddings.shape}, expected {expected_shape}")
    options = build_options_data(
        extractor.name,
        max_chars_per_chunk,
        chunk_overlap_chars,
        clean_text,
        remove_front_matter,
        front_matter_max_page,
        remove_toc_pages,
        toc_max_page,
        deduplicate_chunks,
    )
    options["split_content_types"] = True
    options["content_category"] = category
    manifest = manifest_data(
        pack_id=pack_content.pack_id,
        title=title,
        system=system,
        edition=edition,
        language=language,
        documents=pack_content.documents,
        extractor_name=extractor.name,
        embedding_provider=embedding_provider,
        chunk_count=len(pack_content.chunks),
        build_options=options,
    )
    document_rows = [document_metadata(document) for document in pack_content.documents]
    report = build_extraction_report(
        extractor.name,
        pack_content.documents,
        max_chars_per_chunk,
        pack_content.cleanup_report,
        pack_content.front_matter_report,
        pack_content.toc_report,
        pack_content.chunk_quality_report,
        build_timing_report(started_at, prepared_at, embedded_at),
    )
    report["content_classification"] = classification_report
    report["content_category"] = category
    write_pack_archive(
        out_path=out_path,
        manifest=manifest,
        documents=document_rows,
        chunks=pack_content.chunks,
        embeddings=embeddings,
        extraction_report=report,
    )
    return BuildResult(manifest, document_rows, pack_content.chunks, report)


def attach_output_paths(
    decisions: dict[str, dict[str, object]],
    output_paths: dict[ContentLabel, Path],
) -> dict[str, dict[str, object]]:
    return {
        label.value: decisions[label.value] | {"output_path": str(path)}
        for label, path in output_paths.items()
    }


def preflight_split_outputs(paths: list[Path], *, force: bool) -> None:
    conflicts = [path for path in paths if path.exists()]
    if conflicts and not force:
        raise FileExistsError("split output already exists: " + ", ".join(str(path) for path in conflicts))


def commit_split_archives(
    temporary_paths: dict[ContentLabel, Path],
    output_paths: dict[ContentLabel, Path],
    *,
    force: bool,
) -> None:
    backups: dict[Path, Path] = {}
    committed: list[Path] = []
    try:
        if force:
            for label in temporary_paths:
                destination = output_paths[label]
                if destination.exists():
                    file_descriptor, backup_name = tempfile.mkstemp(
                        prefix=f".{destination.stem}-backup-",
                        suffix=".gmnpack",
                        dir=destination.parent,
                    )
                    os.close(file_descriptor)
                    backup = Path(backup_name)
                    backup.unlink()
                    os.replace(destination, backup)
                    backups[destination] = backup
        for label, temporary_path in temporary_paths.items():
            destination = output_paths[label]
            os.replace(temporary_path, destination)
            committed.append(destination)
    except Exception:
        for destination in committed:
            destination.unlink(missing_ok=True)
        for destination, backup in backups.items():
            if backup.exists():
                os.replace(backup, destination)
        raise
    else:
        for backup in backups.values():
            backup.unlink(missing_ok=True)


def aggregate_split_report(
    classification_report: dict[str, object],
    decisions: dict[str, dict[str, object]],
    results: dict[str, BuildResult] | None = None,
) -> dict[str, object]:
    return {
        "content_classification": classification_report,
        "categories": decisions,
        "written_categories": sorted((results or {}).keys()),
    }




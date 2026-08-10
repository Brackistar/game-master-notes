from __future__ import annotations

import io
import json
import zipfile
from dataclasses import dataclass, field
from pathlib import Path

import numpy as np

from pack_builder.constants import (
    CHUNKS_FILE,
    DOCUMENTS_FILE,
    EMBEDDINGS_FILE,
    EXTRACTION_REPORT_FILE,
    MANIFEST_FILE,
    PACK_SCHEMA_VERSION,
    REQUIRED_PACK_FILES,
)


@dataclass(frozen=True)
class ValidationResult:
    ok: bool
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)
    manifest: dict[str, object] | None = None


def _read_json(archive: zipfile.ZipFile, filename: str) -> dict[str, object]:
    with archive.open(filename) as handle:
        return json.loads(handle.read().decode("utf-8"))


def _read_chunks(archive: zipfile.ZipFile) -> list[dict[str, object]]:
    with archive.open(CHUNKS_FILE) as handle:
        lines = handle.read().decode("utf-8").splitlines()
    return [json.loads(line) for line in lines if line.strip()]


def _read_embeddings(archive: zipfile.ZipFile) -> np.ndarray:
    with archive.open(EMBEDDINGS_FILE) as handle:
        return np.load(io.BytesIO(handle.read()), allow_pickle=False)


def validate_pack(pack_path: Path) -> ValidationResult:
    errors: list[str] = []
    warnings: list[str] = []
    manifest: dict[str, object] | None = None

    if not pack_path.exists():
        return ValidationResult(ok=False, errors=[f"pack does not exist: {pack_path}"])

    try:
        with zipfile.ZipFile(pack_path, mode="r") as archive:
            names = set(archive.namelist())
            missing = sorted(REQUIRED_PACK_FILES - names)
            if missing:
                return ValidationResult(
                    ok=False,
                    errors=[f"missing required file: {name}" for name in missing],
                )

            try:
                manifest = _read_json(archive, MANIFEST_FILE)
                documents_payload = _read_json(archive, DOCUMENTS_FILE)
                _read_json(archive, EXTRACTION_REPORT_FILE)
                chunks = _read_chunks(archive)
                embeddings = _read_embeddings(archive)
            except Exception as exc:
                return ValidationResult(ok=False, errors=[f"malformed pack: {exc}"])
    except zipfile.BadZipFile:
        return ValidationResult(ok=False, errors=["pack is not a valid ZIP archive"])

    if manifest.get("schema_version") != PACK_SCHEMA_VERSION:
        errors.append("manifest schema_version is unsupported or missing")

    required_manifest_fields = {
        "pack_id",
        "title",
        "system",
        "edition",
        "language",
        "source_pdfs",
        "generator_version",
        "extractor_name",
        "embedding_model_id",
        "embedding_dimensions",
        "chunk_count",
        "created_at",
    }
    for field_name in sorted(required_manifest_fields):
        if field_name not in manifest:
            errors.append(f"manifest missing field: {field_name}")

    documents = documents_payload.get("documents")
    if not isinstance(documents, list) or not documents:
        errors.append("documents.json must contain a non-empty documents list")

    chunk_count = manifest.get("chunk_count")
    if chunk_count != len(chunks):
        errors.append(
            f"manifest chunk_count {chunk_count!r} does not match chunks {len(chunks)}"
        )

    dimensions = manifest.get("embedding_dimensions")
    if len(embeddings.shape) != 2:
        errors.append("embeddings.npy must be a 2D array")
    else:
        if embeddings.shape[0] != len(chunks):
            errors.append(
                f"embedding rows {embeddings.shape[0]} do not match chunks {len(chunks)}"
            )
        if dimensions != embeddings.shape[1]:
            errors.append(
                f"manifest embedding_dimensions {dimensions!r} "
                f"does not match embeddings width {embeddings.shape[1]}"
            )

    required_chunk_fields = {
        "chunk_id",
        "document_id",
        "page_start",
        "page_end",
        "citation_label",
        "text",
        "char_count",
        "embedding_row_index",
    }
    seen_chunk_ids: set[str] = set()
    for index, chunk in enumerate(chunks):
        if not isinstance(chunk, dict):
            errors.append(f"chunk row {index} is not an object")
            continue
        for field_name in sorted(required_chunk_fields):
            if field_name not in chunk:
                errors.append(f"chunk row {index} missing field: {field_name}")
        chunk_id = chunk.get("chunk_id")
        if isinstance(chunk_id, str):
            if chunk_id in seen_chunk_ids:
                errors.append(f"duplicate chunk_id: {chunk_id}")
            seen_chunk_ids.add(chunk_id)
        if chunk.get("embedding_row_index") != index:
            errors.append(f"chunk row {index} has mismatched embedding_row_index")
        text = chunk.get("text")
        if not isinstance(text, str) or not text.strip():
            errors.append(f"chunk row {index} has empty text")
        elif chunk.get("char_count") != len(text):
            errors.append(f"chunk row {index} char_count does not match text length")
        page_start = chunk.get("page_start")
        page_end = chunk.get("page_end")
        if isinstance(page_start, int) and isinstance(page_end, int):
            if page_start <= 0 or page_end < page_start:
                errors.append(f"chunk row {index} has invalid page range")
        else:
            errors.append(f"chunk row {index} page_start/page_end must be integers")

    if embeddings.dtype != np.float32:
        warnings.append(f"embeddings dtype is {embeddings.dtype}, expected float32")

    return ValidationResult(
        ok=not errors,
        errors=errors,
        warnings=warnings,
        manifest=manifest,
    )

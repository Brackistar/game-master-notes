from __future__ import annotations

import io
import json
import zipfile
from dataclasses import dataclass, field
from pathlib import Path

import numpy as np

from pack_builder.core_domain.constants import (
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


@dataclass(frozen=True)
class PackPayload:
    manifest: dict[str, object]
    documents_payload: dict[str, object]
    chunks: list[dict[str, object]]
    embeddings: np.ndarray


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
    if not pack_path.exists():
        return ValidationResult(ok=False, errors=[f"pack does not exist: {pack_path}"])

    payload_result = read_pack_payload(pack_path)
    if isinstance(payload_result, ValidationResult):
        return payload_result

    errors: list[str] = []
    warnings: list[str] = []
    validate_manifest(payload_result.manifest, errors)
    validate_documents(payload_result.documents_payload, errors)
    validate_embedding_shape(
        payload_result.manifest,
        payload_result.chunks,
        payload_result.embeddings,
        errors,
    )
    validate_chunks(payload_result.chunks, errors)

    if payload_result.embeddings.dtype != np.float32:
        warnings.append(
            f"embeddings dtype is {payload_result.embeddings.dtype}, expected float32"
        )

    return ValidationResult(
        ok=not errors,
        errors=errors,
        warnings=warnings,
        manifest=payload_result.manifest,
    )


def read_pack_payload(pack_path: Path) -> PackPayload | ValidationResult:
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
                payload = PackPayload(
                    manifest=_read_json(archive, MANIFEST_FILE),
                    documents_payload=_read_json(archive, DOCUMENTS_FILE),
                    chunks=_read_chunks(archive),
                    embeddings=_read_embeddings(archive),
                )
                _read_json(archive, EXTRACTION_REPORT_FILE)
                return payload
            except Exception as exc:
                return ValidationResult(ok=False, errors=[f"malformed pack: {exc}"])
    except zipfile.BadZipFile:
        return ValidationResult(ok=False, errors=["pack is not a valid ZIP archive"])


def validate_manifest(manifest: dict[str, object], errors: list[str]) -> None:
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


def validate_documents(
    documents_payload: dict[str, object],
    errors: list[str],
) -> None:
    documents = documents_payload.get("documents")
    if not isinstance(documents, list) or not documents:
        errors.append("documents.json must contain a non-empty documents list")


def validate_embedding_shape(
    manifest: dict[str, object],
    chunks: list[dict[str, object]],
    embeddings: np.ndarray,
    errors: list[str],
) -> None:
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


def validate_chunks(chunks: list[dict[str, object]], errors: list[str]) -> None:
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
        validate_chunk_object(
            chunk,
            index,
            required_chunk_fields,
            seen_chunk_ids,
            errors,
        )


def validate_chunk_object(
    chunk: dict[str, object],
    index: int,
    required_chunk_fields: set[str],
    seen_chunk_ids: set[str],
    errors: list[str],
) -> None:
    if not isinstance(chunk, dict):
        errors.append(f"chunk row {index} is not an object")
        return
    validate_required_chunk_fields(chunk, index, required_chunk_fields, errors)
    validate_chunk_id(chunk, seen_chunk_ids, errors)
    validate_chunk_row_index(chunk, index, errors)
    validate_chunk_text(chunk, index, errors)
    validate_chunk_page_range(chunk, index, errors)


def validate_required_chunk_fields(
    chunk: dict[str, object],
    index: int,
    required_chunk_fields: set[str],
    errors: list[str],
) -> None:
    for field_name in sorted(required_chunk_fields):
        if field_name not in chunk:
            errors.append(f"chunk row {index} missing field: {field_name}")


def validate_chunk_id(
    chunk: dict[str, object],
    seen_chunk_ids: set[str],
    errors: list[str],
) -> None:
    chunk_id = chunk.get("chunk_id")
    if not isinstance(chunk_id, str):
        return
    if chunk_id in seen_chunk_ids:
        errors.append(f"duplicate chunk_id: {chunk_id}")
    seen_chunk_ids.add(chunk_id)


def validate_chunk_row_index(
    chunk: dict[str, object],
    index: int,
    errors: list[str],
) -> None:
    if chunk.get("embedding_row_index") != index:
        errors.append(f"chunk row {index} has mismatched embedding_row_index")


def validate_chunk_text(
    chunk: dict[str, object],
    index: int,
    errors: list[str],
) -> None:
    text = chunk.get("text")
    if not isinstance(text, str) or not text.strip():
        errors.append(f"chunk row {index} has empty text")
        return
    if chunk.get("char_count") != len(text):
        errors.append(f"chunk row {index} char_count does not match text length")


def validate_chunk_page_range(
    chunk: dict[str, object],
    index: int,
    errors: list[str],
) -> None:
    page_start = chunk.get("page_start")
    page_end = chunk.get("page_end")
    if not isinstance(page_start, int) or not isinstance(page_end, int):
        errors.append(f"chunk row {index} page_start/page_end must be integers")
        return
    if page_start <= 0 or page_end < page_start:
        errors.append(f"chunk row {index} has invalid page range")




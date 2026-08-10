from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path

from pack_builder.constants import (
    DEFAULT_EMBEDDING_MODEL_ID,
    DEFAULT_EXTRACTOR,
    DEFAULT_MAX_CHARS_PER_CHUNK,
)


@dataclass(frozen=True)
class BuildOptions:
    pdf_paths: list[Path]
    system: str
    edition: str
    title: str
    out_path: Path
    extractor: str = DEFAULT_EXTRACTOR
    language: str = "en"
    embedding_provider: str = "sentence-transformers"
    embedding_model: str = DEFAULT_EMBEDDING_MODEL_ID
    max_chars_per_chunk: int = DEFAULT_MAX_CHARS_PER_CHUNK
    chunk_overlap_chars: int = 0
    clean_text: bool = True
    deduplicate_chunks: bool = True
    force: bool = False
    dry_run: bool = False
    report_out: Path | None = None


@dataclass(frozen=True)
class BuildOverrides:
    pdfs: list[Path] | None = None
    system: str | None = None
    edition: str | None = None
    title: str | None = None
    out: Path | None = None
    extractor: str | None = None
    language: str | None = None
    embedding_provider: str | None = None
    embedding_model: str | None = None
    max_chars_per_chunk: int | None = None
    chunk_overlap_chars: int | None = None
    clean_text: bool | None = None
    deduplicate_chunks: bool | None = None
    force: bool = False
    dry_run: bool = False
    report_out: Path | None = None


def load_build_config(config_path: Path | None) -> dict[str, object]:
    if not config_path:
        return {}
    try:
        payload = json.loads(config_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        raise ValueError(f"build config is not valid JSON: {exc}") from exc
    if not isinstance(payload, dict):
        raise ValueError("build config must be a JSON object")
    return payload


def resolve_build_options(
    config_path: Path | None,
    overrides: BuildOverrides,
) -> BuildOptions:
    config = load_build_config(config_path)
    options = BuildOptions(
        pdf_paths=_resolve_pdfs(config, overrides.pdfs),
        system=_required_str(config, "system", overrides.system),
        edition=_required_str(config, "edition", overrides.edition),
        title=_required_str(config, "title", overrides.title),
        out_path=_required_path(config, "out", overrides.out),
        extractor=_optional_str(config, "extractor", overrides.extractor, DEFAULT_EXTRACTOR),
        language=_optional_str(config, "language", overrides.language, "en"),
        embedding_provider=_optional_str(
            config,
            "embedding_provider",
            overrides.embedding_provider,
            "sentence-transformers",
        ),
        embedding_model=_optional_str(
            config,
            "embedding_model",
            overrides.embedding_model,
            DEFAULT_EMBEDDING_MODEL_ID,
        ),
        max_chars_per_chunk=_optional_int(
            config,
            "max_chars_per_chunk",
            overrides.max_chars_per_chunk,
            DEFAULT_MAX_CHARS_PER_CHUNK,
        ),
        chunk_overlap_chars=_optional_int(
            config,
            "chunk_overlap_chars",
            overrides.chunk_overlap_chars,
            0,
        ),
        clean_text=_optional_bool(config, "clean_text", overrides.clean_text, True),
        deduplicate_chunks=_optional_bool(
            config,
            "deduplicate_chunks",
            overrides.deduplicate_chunks,
            True,
        ),
        force=_optional_bool(config, "force", overrides.force),
        dry_run=_optional_bool(config, "dry_run", overrides.dry_run),
        report_out=_optional_path(config, "report_out", overrides.report_out),
    )
    validate_build_options(options)
    return options


def validate_build_options(options: BuildOptions) -> None:
    if not options.pdf_paths:
        raise ValueError("missing required values: pdfs")
    for pdf_path in options.pdf_paths:
        if not pdf_path.exists() or not pdf_path.is_file():
            raise ValueError(f"PDF does not exist: {pdf_path}")
    if options.max_chars_per_chunk < 200:
        raise ValueError("--max-chars-per-chunk must be at least 200")
    if options.chunk_overlap_chars < 0:
        raise ValueError("--chunk-overlap-chars must be zero or greater")
    if options.chunk_overlap_chars >= options.max_chars_per_chunk:
        raise ValueError("--chunk-overlap-chars must be smaller than chunk size")
    if options.out_path.exists() and not options.force and not options.dry_run:
        raise FileExistsError(f"output already exists: {options.out_path}")


def _config_value(
    config: dict[str, object],
    key: str,
    explicit_value: object,
    default_value: object = None,
) -> object:
    if explicit_value is not None:
        return explicit_value
    return config.get(key, default_value)


def _resolve_pdfs(config: dict[str, object], explicit_pdfs: list[Path] | None) -> list[Path]:
    if explicit_pdfs:
        return explicit_pdfs
    raw_pdfs = config.get("pdfs", [])
    if not isinstance(raw_pdfs, list):
        raise ValueError("build config field 'pdfs' must be a list")
    return [Path(str(path)) for path in raw_pdfs]


def _required_str(
    config: dict[str, object],
    key: str,
    explicit_value: str | None,
) -> str:
    value = _config_value(config, key, explicit_value)
    if not value:
        raise ValueError(f"missing required values: {key}")
    return str(value)


def _optional_str(
    config: dict[str, object],
    key: str,
    explicit_value: str | None,
    default_value: str,
) -> str:
    return str(_config_value(config, key, explicit_value, default_value))


def _required_path(
    config: dict[str, object],
    key: str,
    explicit_value: Path | None,
) -> Path:
    value = _config_value(config, key, explicit_value)
    if not value:
        raise ValueError(f"missing required values: {key}")
    return Path(str(value))


def _optional_path(
    config: dict[str, object],
    key: str,
    explicit_value: Path | None,
) -> Path | None:
    value = _config_value(config, key, explicit_value)
    return Path(str(value)) if value else None


def _optional_int(
    config: dict[str, object],
    key: str,
    explicit_value: int | None,
    default_value: int,
) -> int:
    return int(_config_value(config, key, explicit_value, default_value))


def _optional_bool(
    config: dict[str, object],
    key: str,
    explicit_value: bool | None,
    default_value: bool = False,
) -> bool:
    if explicit_value is not None:
        return explicit_value
    return bool(config.get(key, default_value))

from __future__ import annotations

import json
from enum import Enum
from pathlib import Path
from typing import Annotated

import typer
from rich.console import Console
from rich.table import Table

from pack_builder.constants import (
    DEFAULT_EMBEDDING_MODEL_ID,
    DEFAULT_EXTRACTOR,
    DEFAULT_MAX_CHARS_PER_CHUNK,
)
from pack_builder.embeddings import get_embedding_provider
from pack_builder.pack_writer import build_pack, preview_pack
from pack_builder.pdf_extract import get_extractor
from pack_builder.validate import validate_pack


class ExtractorName(str, Enum):
    pymupdf = "pymupdf"
    pdfplumber = "pdfplumber"


class EmbeddingProviderName(str, Enum):
    sentence_transformers = "sentence-transformers"
    deterministic = "deterministic"


app = typer.Typer(no_args_is_help=True)
console = Console()


def write_json_file(path: Path, payload: dict[str, object]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")


@app.command()
def build(
    pdfs: Annotated[
        list[Path],
        typer.Argument(
            exists=True,
            file_okay=True,
            dir_okay=False,
            readable=True,
            help="One or more user-owned PDFs to package.",
        ),
    ],
    system: Annotated[str, typer.Option("--system", help="Game system name.")],
    edition: Annotated[str, typer.Option("--edition", help="Game system edition.")],
    title: Annotated[str, typer.Option("--title", help="Pack title.")],
    out: Annotated[Path, typer.Option("--out", help="Output .gmnpack path.")],
    extractor: Annotated[
        ExtractorName,
        typer.Option("--extractor", help="PDF extraction adapter."),
    ] = ExtractorName(DEFAULT_EXTRACTOR),
    language: Annotated[str, typer.Option("--language", help="Pack language code.")] = "en",
    embedding_provider: Annotated[
        EmbeddingProviderName,
        typer.Option(
            "--embedding-provider",
            help="Embedding provider. Deterministic is for tests and local fixtures.",
        ),
    ] = EmbeddingProviderName.sentence_transformers,
    embedding_model: Annotated[
        str,
        typer.Option("--embedding-model", help="Sentence Transformers model id."),
    ] = DEFAULT_EMBEDDING_MODEL_ID,
    max_chars_per_chunk: Annotated[
        int,
        typer.Option(
            "--max-chars-per-chunk",
            min=200,
            help="Maximum normalized characters per retrieval chunk.",
        ),
    ] = DEFAULT_MAX_CHARS_PER_CHUNK,
    force: Annotated[
        bool,
        typer.Option("--force", help="Overwrite an existing output pack."),
    ] = False,
    dry_run: Annotated[
        bool,
        typer.Option("--dry-run", help="Extract and chunk without writing a pack."),
    ] = False,
    report_out: Annotated[
        Path | None,
        typer.Option("--report-out", help="Write extraction report JSON to this path."),
    ] = None,
    verbose: Annotated[
        bool,
        typer.Option("--verbose", help="Print extraction and quality details."),
    ] = False,
) -> None:
    """Build a .gmnpack archive from one or more PDFs."""
    if out.exists() and not force and not dry_run:
        console.print(f"[red]Build failed:[/red] output already exists: {out}")
        console.print("Use --force to overwrite it.")
        raise typer.Exit(code=1)

    try:
        pdf_extractor = get_extractor(extractor.value)
        if dry_run:
            build_result = preview_pack(
                pdf_paths=pdfs,
                title=title,
                system=system,
                edition=edition,
                language=language,
                extractor=pdf_extractor,
                max_chars_per_chunk=max_chars_per_chunk,
            )
        else:
            embeddings = get_embedding_provider(embedding_provider.value, embedding_model)
            build_result = build_pack(
                pdf_paths=pdfs,
                out_path=out,
                title=title,
                system=system,
                edition=edition,
                language=language,
                extractor=pdf_extractor,
                embedding_provider=embeddings,
                max_chars_per_chunk=max_chars_per_chunk,
            )
    except Exception as exc:
        console.print(f"[red]Build failed:[/red] {exc}")
        raise typer.Exit(code=1) from exc

    if report_out:
        write_json_file(report_out, build_result.extraction_report)
        console.print(f"[green]Wrote report[/green] {report_out}")

    if dry_run:
        console.print("[green]Dry run complete.[/green]")
        console.print(
            f"Pack {build_result.manifest['pack_id']} would contain "
            f"{build_result.manifest['chunk_count']} chunks."
        )
        if verbose:
            print_quality_summary(build_result.extraction_report)
        return

    validation = validate_pack(out)
    if not validation.ok:
        console.print("[red]Pack was written but failed validation:[/red]")
        for error in validation.errors:
            console.print(f"- {error}")
        raise typer.Exit(code=1)

    console.print(f"[green]Wrote[/green] {out}")
    console.print(
        f"Pack {build_result.manifest['pack_id']} with "
        f"{build_result.manifest['chunk_count']} chunks and "
        f"{build_result.manifest['embedding_dimensions']}d embeddings."
    )
    if verbose:
        print_quality_summary(build_result.extraction_report)


def print_quality_summary(extraction_report: dict[str, object]) -> None:
    empty_pages = extraction_report.get("empty_pages", [])
    suspicious_pages = extraction_report.get("suspicious_pages", [])
    duplicate_pages = extraction_report.get("duplicate_pages", [])
    warnings = extraction_report.get("warnings", [])
    console.print(
        "Quality: "
        f"{len(empty_pages)} empty pages, "
        f"{len(suspicious_pages)} suspicious pages, "
        f"{len(duplicate_pages)} duplicate pages, "
        f"{len(warnings)} warnings."
    )


@app.command()
def inspect(
    pack: Annotated[
        Path,
        typer.Argument(exists=True, file_okay=True, dir_okay=False, readable=True),
    ],
    as_json: Annotated[
        bool,
        typer.Option("--json", help="Print machine-readable JSON."),
    ] = False,
) -> None:
    """Print pack metadata and validation status."""
    result = validate_pack(pack)
    if not result.manifest:
        console.print("[red]Could not read manifest.[/red]")
        for error in result.errors:
            console.print(f"- {error}")
        raise typer.Exit(code=1)

    manifest = result.manifest
    if as_json:
        console.print_json(
            data={
                "ok": result.ok,
                "manifest": manifest,
                "errors": result.errors,
                "warnings": result.warnings,
            }
        )
        if not result.ok:
            raise typer.Exit(code=1)
        return

    table = Table(title=str(pack))
    table.add_column("Field")
    table.add_column("Value")
    for field_name in [
        "pack_id",
        "title",
        "system",
        "edition",
        "language",
        "schema_version",
        "extractor_name",
        "embedding_model_id",
        "embedding_dimensions",
        "chunk_count",
        "created_at",
    ]:
        table.add_row(field_name, str(manifest.get(field_name, "")))
    console.print(table)

    if result.ok:
        console.print("[green]Validation: ok[/green]")
    else:
        console.print("[red]Validation: failed[/red]")
        for error in result.errors:
            console.print(f"- {error}")
        raise typer.Exit(code=1)


@app.command()
def validate(
    pack: Annotated[
        Path,
        typer.Argument(exists=True, file_okay=True, dir_okay=False, readable=True),
    ],
    as_json: Annotated[
        bool,
        typer.Option("--json", help="Print machine-readable JSON."),
    ] = False,
) -> None:
    """Validate a .gmnpack archive."""
    result = validate_pack(pack)
    if as_json:
        console.print_json(
            data={
                "ok": result.ok,
                "errors": result.errors,
                "warnings": result.warnings,
                "manifest": result.manifest,
            }
        )
        if not result.ok:
            raise typer.Exit(code=1)
        return

    if result.ok:
        console.print("[green]Pack is valid.[/green]")
        for warning in result.warnings:
            console.print(f"[yellow]Warning:[/yellow] {warning}")
        return

    console.print("[red]Pack is invalid.[/red]")
    for error in result.errors:
        console.print(f"- {error}")
    raise typer.Exit(code=1)


if __name__ == "__main__":
    app()

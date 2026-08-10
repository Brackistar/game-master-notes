from __future__ import annotations

from enum import Enum
from pathlib import Path
from typing import Annotated

import typer
from rich.console import Console
from rich.table import Table

from pack_builder.build_config import BuildOptions, BuildOverrides, resolve_build_options
from pack_builder.embeddings import get_embedding_provider
from pack_builder.extractor_compare import compare_extractors
from pack_builder.pack_reader import read_chunks, read_extraction_report
from pack_builder.pack_writer import build_pack, preview_pack
from pack_builder.pdf_extract import get_extractor
from pack_builder.schema_contract import pack_schema_contract
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
    import json

    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")


@app.command()
def build(
    pdfs: Annotated[
        list[Path] | None,
        typer.Argument(
            file_okay=True,
            dir_okay=False,
            readable=True,
            help="One or more user-owned PDFs to package.",
        ),
    ] = None,
    config: Annotated[
        Path | None,
        typer.Option("--config", exists=True, help="JSON build config file."),
    ] = None,
    system: Annotated[str | None, typer.Option("--system", help="Game system name.")] = None,
    edition: Annotated[
        str | None, typer.Option("--edition", help="Game system edition.")
    ] = None,
    title: Annotated[str | None, typer.Option("--title", help="Pack title.")] = None,
    out: Annotated[
        Path | None, typer.Option("--out", help="Output .gmnpack path.")
    ] = None,
    extractor: Annotated[
        ExtractorName | None,
        typer.Option("--extractor", help="PDF extraction adapter."),
    ] = None,
    language: Annotated[
        str | None, typer.Option("--language", help="Pack language code.")
    ] = None,
    embedding_provider: Annotated[
        EmbeddingProviderName | None,
        typer.Option(
            "--embedding-provider",
            help="Embedding provider. Deterministic is for tests and local fixtures.",
        ),
    ] = None,
    embedding_model: Annotated[
        str | None,
        typer.Option("--embedding-model", help="Sentence Transformers model id."),
    ] = None,
    max_chars_per_chunk: Annotated[
        int | None,
        typer.Option(
            "--max-chars-per-chunk",
            min=200,
            help="Maximum normalized characters per retrieval chunk.",
        ),
    ] = None,
    chunk_overlap_chars: Annotated[
        int | None,
        typer.Option(
            "--chunk-overlap-chars",
            min=0,
            help="Characters of previous chunk context to prepend to each chunk.",
        ),
    ] = None,
    clean_text: Annotated[
        bool | None,
        typer.Option(
            "--clean-text/--no-clean-text",
            help="Remove repeated lines and repair hyphenation before chunking.",
        ),
    ] = None,
    deduplicate_chunks: Annotated[
        bool | None,
        typer.Option(
            "--deduplicate-chunks/--no-deduplicate-chunks",
            help="Remove duplicate normalized chunk text.",
        ),
    ] = None,
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
    try:
        options = resolve_build_options(
            config,
            BuildOverrides(
                pdfs=pdfs,
                system=system,
                edition=edition,
                title=title,
                out=out,
                extractor=extractor.value if extractor else None,
                language=language,
                embedding_provider=embedding_provider.value
                if embedding_provider
                else None,
                embedding_model=embedding_model,
                max_chars_per_chunk=max_chars_per_chunk,
                chunk_overlap_chars=chunk_overlap_chars,
                clean_text=clean_text,
                deduplicate_chunks=deduplicate_chunks,
                force=force,
                dry_run=dry_run,
                report_out=report_out,
            ),
        )
        build_result = run_build(options)
    except Exception as exc:
        console.print(f"[red]Build failed:[/red] {exc}")
        if isinstance(exc, FileExistsError):
            console.print("Use --force to overwrite it.")
        raise typer.Exit(code=1) from exc

    if options.report_out:
        write_json_file(options.report_out, build_result.extraction_report)
        console.print(f"[green]Wrote report[/green] {options.report_out}")

    if options.dry_run:
        console.print("[green]Dry run complete.[/green]")
        console.print(
            f"Pack {build_result.manifest['pack_id']} would contain "
            f"{build_result.manifest['chunk_count']} chunks."
        )
        if verbose:
            print_quality_summary(build_result.extraction_report)
        return

    validation = validate_pack(options.out_path)
    if not validation.ok:
        console.print("[red]Pack was written but failed validation:[/red]")
        for error in validation.errors:
            console.print(f"- {error}")
        raise typer.Exit(code=1)

    console.print(f"[green]Wrote[/green] {options.out_path}")
    console.print(
        f"Pack {build_result.manifest['pack_id']} with "
        f"{build_result.manifest['chunk_count']} chunks and "
        f"{build_result.manifest['embedding_dimensions']}d embeddings."
    )
    if verbose:
        print_quality_summary(build_result.extraction_report)


def run_build(options: BuildOptions):
    pdf_extractor = get_extractor(options.extractor)
    if options.dry_run:
        with console.status("Extracting and chunking PDFs..."):
            return preview_pack(
                pdf_paths=options.pdf_paths,
                title=options.title,
                system=options.system,
                edition=options.edition,
                language=options.language,
                extractor=pdf_extractor,
                max_chars_per_chunk=options.max_chars_per_chunk,
                clean_text=options.clean_text,
                deduplicate_chunks=options.deduplicate_chunks,
                chunk_overlap_chars=options.chunk_overlap_chars,
            )

    embeddings = get_embedding_provider(
        options.embedding_provider,
        options.embedding_model,
    )
    with console.status("Extracting, chunking, embedding, and writing pack..."):
        return build_pack(
            pdf_paths=options.pdf_paths,
            out_path=options.out_path,
            title=options.title,
            system=options.system,
            edition=options.edition,
            language=options.language,
            extractor=pdf_extractor,
            embedding_provider=embeddings,
            max_chars_per_chunk=options.max_chars_per_chunk,
            clean_text=options.clean_text,
            deduplicate_chunks=options.deduplicate_chunks,
            chunk_overlap_chars=options.chunk_overlap_chars,
        )


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


@app.command("report")
def report(
    pack: Annotated[
        Path,
        typer.Argument(exists=True, file_okay=True, dir_okay=False, readable=True),
    ],
    as_json: Annotated[
        bool,
        typer.Option("--json", help="Print full extraction report JSON."),
    ] = False,
) -> None:
    """Print extraction quality report summary."""
    extraction_report = read_extraction_report(pack)
    if as_json:
        console.print_json(data=extraction_report)
        return

    table = Table(title=f"Extraction report: {pack}")
    table.add_column("Metric")
    table.add_column("Value")
    table.add_row("extractor", str(extraction_report.get("extractor", "")))
    chunking = extraction_report.get("chunking", {})
    if isinstance(chunking, dict):
        table.add_row(
            "max_chars_per_chunk",
            str(chunking.get("max_chars_per_chunk", "")),
        )
    for field_name in ["empty_pages", "suspicious_pages", "duplicate_pages", "warnings", "errors"]:
        value = extraction_report.get(field_name, [])
        table.add_row(field_name, str(len(value) if isinstance(value, list) else value))
    console.print(table)


@app.command("compare-extractors")
def compare_extractors_command(
    pdf: Annotated[
        Path,
        typer.Argument(exists=True, file_okay=True, dir_okay=False, readable=True),
    ],
    as_json: Annotated[
        bool,
        typer.Option("--json", help="Print comparison report as JSON."),
    ] = False,
) -> None:
    """Compare PyMuPDF and pdfplumber extraction output."""
    comparison = compare_extractors(pdf)
    if as_json:
        console.print_json(data=comparison)
        return

    table = Table(title=f"Extractor comparison: {pdf}")
    table.add_column("Metric")
    table.add_column("Value")
    summary = comparison["summary"]
    if isinstance(summary, dict):
        for key, value in summary.items():
            table.add_row(key, str(value))
    console.print(table)


@app.command("schema")
def schema(
    as_json: Annotated[
        bool,
        typer.Option("--json", help="Print schema contract as JSON."),
    ] = False,
) -> None:
    """Print the .gmnpack schema contract."""
    contract = pack_schema_contract()
    if as_json:
        console.print_json(data=contract)
        return

    table = Table(title=".gmnpack schema")
    table.add_column("Field")
    table.add_column("Value")
    table.add_row("schema_version", str(contract["schema_version"]))
    table.add_row("embedding_format", str(contract["embedding_format"]))
    table.add_row("archive_files", ", ".join(contract["archive_files"]))
    console.print(table)


@app.command("sample-chunks")
def sample_chunks(
    pack: Annotated[
        Path,
        typer.Argument(exists=True, file_okay=True, dir_okay=False, readable=True),
    ],
    limit: Annotated[
        int,
        typer.Option("--limit", min=1, help="Maximum chunks to print."),
    ] = 3,
    contains: Annotated[
        str | None,
        typer.Option("--contains", help="Only sample chunks containing this text."),
    ] = None,
    as_json: Annotated[
        bool,
        typer.Option("--json", help="Print sampled chunks as JSON."),
    ] = False,
) -> None:
    """Print sample chunks from a pack for manual quality inspection."""
    chunks = read_chunks(pack)
    if contains:
        needle = contains.lower()
        chunks = [chunk for chunk in chunks if needle in str(chunk.get("text", "")).lower()]
    sampled = chunks[:limit]
    if as_json:
        console.print_json(data={"chunks": sampled, "count": len(sampled)})
        return

    for chunk in sampled:
        console.rule(str(chunk.get("citation_label", chunk.get("chunk_id", "chunk"))))
        text = str(chunk.get("text", ""))
        console.print(text[:700] + ("..." if len(text) > 700 else ""))


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

# Data Layer Plan

## Metadata

- File: `docs/data-layer-plan.md`
- Created: 2026-08-10
- Last updated: 2026-08-10
- User: brackistar

Related diagram: [data-layer-plan.mmd](data-layer-plan.mmd)

## Purpose

The data layer is the app's memory. It stores user-created campaign material, imported sourcebook chunks, citations, indexes, model metadata, and benchmark results. The local AI should use this data through retrieval instead of acting as the source of truth.

## Product Goals

- Store multiple game systems, campaigns, and sessions.
- Store flexible lore entities without overfitting one RPG system.
- Store sourcebook pack metadata and citeable chunks.
- Support keyword search with SQLite FTS.
- Support semantic search with local vectors.
- Preserve citations from imported sourcebooks.
- Support migrations as the app evolves.

## Storage Stack

- SQLite as the local database.
- Room for typed Android persistence.
- SQLite FTS5 for keyword search.
- Local vector storage for embeddings, starting simple and replaceable.
- App-private file storage for imported pack archives and GGUF model files.

## Core Entities

Start with these entities:

- `System`: game system or ruleset.
- `SourcebookPack`: imported pack metadata.
- `SourceDocument`: document inside a pack.
- `SourceChunk`: citeable sourcebook text chunk.
- `Campaign`: campaign owned by the user.
- `Session`: planned or completed game session.
- `Note`: freeform user note.
- `LoreEntity`: NPC, location, faction, item, event, concept, or custom entity.
- `Tag`: user-defined label.
- `Link`: relationship between notes, entities, sessions, campaigns, and source chunks.
- `Embedding`: vector metadata and pointer/storage for note or chunk embeddings.
- `ModelProfile`: installed local model metadata.
- `BenchmarkRun`: model benchmark result on device.

## Sourcebook Pack Tables

Pack import should preserve:

- Pack id and schema version.
- Source PDF filename and checksum.
- Game system, edition, title, and language.
- Document sections and page ranges where known.
- Chunk text, chunk id, page start, page end, and citation label.
- Embedding model id and dimensions.

Imported chunk ids should remain stable so citations, search results, and saved assistant context can refer back to the same source.

## Note and Lore Model

Avoid over-modeling RPG-specific details early. Use a flexible shared model:

- A lore entity has a type, name, description, campaign scope, tags, and metadata.
- Notes can link to any entity, session, campaign, or source chunk.
- Links should have a relationship type such as `mentions`, `involves`, `supports`, `contradicts`, or `inspired_by`.

This supports many systems without needing a separate schema for each game.

## Search Indexes

FTS should cover:

- Note title and body.
- Lore entity name and description.
- Source chunk text.
- Tags where useful.

Vector search should cover:

- Source chunks.
- User notes.
- Lore entity summaries.

For v1, vector search can be implemented with a simple local index if the corpus is modest. If performance becomes poor, replace the index behind a repository interface.

## Import Behavior

Pack import should:

- Validate manifest and schema version.
- Validate chunk count and embedding count.
- Reject duplicate pack ids unless the user chooses replace.
- Import in a transaction where practical.
- Record import status and errors.
- Avoid loading entire large packs into memory.

## Migration Strategy

- Start schema versioning immediately.
- Add Room migrations for every database change after initial release.
- Keep imported pack schema version separate from Android database schema version.
- Preserve source chunk ids across migrations.

## Testing

- Unit test entity mapping and repositories.
- Migration tests from every released schema.
- Import validation tests with tiny synthetic packs.
- FTS search tests for notes and source chunks.
- Vector search tests with deterministic toy embeddings.

## Risks

- Vector storage choices may need revision after real pack sizes are known.
- Too many entity types can make the UI and schema brittle.
- Large packs can make migrations slow.
- Reimporting packs must not break saved citations.

## Suggestions

- Keep the first data model boring and durable.
- Use flexible metadata only where it prevents schema churn.
- Build explicit citation objects early.
- Treat import and search as first-class features, not utility code hidden behind the UI.

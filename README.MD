# game-master-notes

An AI-powered note-taking and world-management app for game masters running multiple games and campaigns.

## Overview

`game-master-notes` is an early-stage project focused on a simple idea: a game master should be able to capture ideas quickly, organize worlds cleanly, and get useful AI help without turning prep into a second job.

The product is being designed around a live transcription-first workflow. Instead of relying only on typed notes, the goal is to let a GM speak ideas as they come, turn those into structured notes, connect them to worlds and campaigns, and retrieve them later through keyword and semantic search. On top of that foundation, the project will include an optional council-style AI assistant for lore checks, plot brainstorming, NPC design, and continuity support.

## Why This Project

Running more than one game usually means scattered notes, forgotten ideas, duplicate prep, and too much context living only in your head. This project exists to make campaign knowledge easier to capture, easier to revisit, and easier to grow over time.

The approach is intentionally practical: prioritize usability first, keep AI costs under control, and use OpenAI as the initial provider so the first versions can focus on product quality instead of local model infrastructure. The goal is not to add AI everywhere, but to make it genuinely helpful where it saves time.

This repository is being built in public by [`brackistar`](https://github.com/brackistar), starting from architecture and scope decisions before expanding into implementation.

## Planned Features

- Multi-world and multi-campaign organization
- Session notes with rich text editing
- Live voice transcription for fast idea capture
- AI-powered summaries and entity extraction
- Semantic search across notes, sessions, and world data
- Manual council AI for lore, NPCs, plot hooks, and continuity checks
- Shared web and desktop experience

## Architecture

The current v1 direction is:

- Frontend: React + TypeScript + Vite
- Desktop wrapper: Tauri
- Core backend: Go
- AI orchestration service: Python + Strands Agents
- Data layer: PostgreSQL + pgvector
- AI provider: OpenAI
- Communication: HTTP/JSON + WebSocket for client traffic, gRPC for internal services

This split keeps the core application logic in Go while using Python where it is most useful for AI orchestration and agent workflows.

## AI and Cost Strategy

The bread-and-butter workflow is live transcription, so the system is being designed with that as the primary AI cost driver. More advanced council-style assistance is manual by design, which helps keep usage predictable and prevents unnecessary background model calls.

The current expectation is that balanced solo usage should remain relatively inexpensive, with transcription as the main recurring cost and deeper AI reasoning used selectively. More detailed cost breakdowns and model routing notes will live in future documentation as the implementation becomes concrete.

## Current Status

This project is currently in planning and early implementation setup.

The scope, architecture, and product direction are defined, and the next step is building the core foundations: data model, service boundaries, note flows, and live transcription pipeline. The repository is intentionally light at the moment and should be read as a project brief and build roadmap rather than a finished application.

## Roadmap

### Phase 1: Foundation and Data Model

- Define core entities for worlds, campaigns, sessions, notes, and tags
- Set up the shared architecture and service boundaries
- Establish the database and API foundations

### Phase 2: Notes and Search

- Build the main note-taking flows
- Add structured organization and linking
- Introduce keyword and semantic search

### Phase 3: Live Transcription

- Implement realtime voice capture and transcript streaming
- Save finalized transcript content into session workflows
- Turn spoken ideas into usable notes

### Phase 4: Council AI

- Add manual AI workflows for lore, plot, NPCs, and continuity
- Keep retrieval grounded in campaign and world data
- Make deeper reasoning opt-in

### Phase 5: Polish, Observability, and Deployment

- Improve reliability and UX
- Add usage tracking and cost visibility
- Prepare production deployment and project hardening

## Repository Structure

The repository is still being bootstrapped, but the intended top-level structure is straightforward:

- `frontend/` for the shared React application
- `backend/` for the Go core service and related backend components

This will evolve as the implementation begins, but the project is intentionally keeping a clean separation between UI, core application logic, and AI orchestration.

## Contributing

Issues, feedback, and ideas are welcome.

If you are interested in game-master tooling, AI-assisted knowledge systems, or helping shape the project direction early, feel free to open an issue or start a discussion. A dedicated contributing guide can be added once the initial implementation structure is in place.

## License

No license has been defined yet for this repository.

Until a license file is added, please treat the project as source-available for review only rather than assuming open-source reuse permissions.

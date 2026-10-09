# Frontend Architecture Roadmap (Web)

This is the source-of-truth roadmap for the frontend.

## 1) Scope

- Web UI architecture and implementation rules
- Frontend domain/workflow composition
- HTTP integration contracts with Go backend
- Testing strategy and quality gates
- Frontend handoff checklist for future sessions

## 2) Locked Decisions

- Stack: React + TypeScript + Vite
- Package manager: npm
- Styling: CSS Modules + shared design tokens
- Drag and drop: dnd-kit
- Initial workflow focus: campaign notes
- Selection model (Phase 1): campaign-first
- API style: REST over HTTP/JSON (`/api/v1` when exposed by backend routing)
- Error handling: user-safe messages only, no tech-stack leakage
- Frontend DTOs are frontend-owned (do not reuse backend structs directly)

## 3) Frontend Architecture Model

### Layering

1. App shell and route composition
2. Feature modules (campaigns, notes, worlds, etc.)
3. UI components (panel/layout/widgets)
4. Application services (API clients + mapping)
5. Shared utilities (validation, formatting, helpers)

### Rules

- Keep feature logic in feature modules, not in generic shared helpers.
- Keep transport mapping inside application services.
- Components consume frontend DTO/view models only.
- Avoid coupling UI components directly to fetch/HTTP calls.
- Prefer explicit interfaces for service dependencies to support unit testing and mocking.

## 4) Initial UX Frame (Phase 1)

- Four-region layout:
  - Left panel: campaign list/actions/search
  - Center panel: main workflow panel (placeholder, switches to create-campaign view)
  - Right panel: contextual actions/details (placeholder for now)
  - Bottom bar: global tools (global search + quick add note entry)
- Left panel requirements:
  - Scrollable vertical list
  - Long names truncated, left aligned
  - Add, delete, search controls on top
  - Drag-drop manual ordering
  - Case-insensitive contains search
  - Delete requires confirmation
- Ordering persistence for now: local UI state only (no backend/localStorage yet)

## 5) Data and Integration Contracts

- Define frontend models:
  - `CampaignViewModel`
  - `NoteViewModel` (when notes panel is implemented)
- Define per-feature data source interfaces:
  - Example: `CampaignDataSource` (`list`, `create`, `delete`, `reorder`)
- Implementations:
  - Phase 1: mock/in-memory adapter
  - Phase 2: HTTP adapter to backend API
- Keep request/response mapping isolated in adapter layer.

## 6) Validation and Error Strategy

- Client validation for campaign naming:
  - trim + required + min 3 + max 50
- API errors mapped to user-facing, domain-level messages.
- Do not expose raw backend/internal error payloads directly in UI.

## 7) Testing Strategy

- Unit/component tests: Vitest + React Testing Library
- Prioritize behavior tests over implementation details:
  - selection state
  - search filtering
  - drag-drop reorder state update
  - add/delete flows and confirmation behavior
- Keep tests independent of backend by mocking data source interfaces.
- Add integration/E2E later after main workflows stabilize.

## 8) Performance and Accessibility Baselines

- Avoid unnecessary rerenders in list-heavy panels.
- Use stable keys and memoization where meaningful.
- Keyboard-accessible controls for panel actions.
- Ensure focus handling for modal/create flows.
- Maintain readable contrast and responsive behavior for laptop screens.

## 9) Delivery Phases (Frontend)

1. App shell + campaign panel interactions (mock data)
2. Campaign + notes workflow in center panel
3. HTTP wiring to backend services
4. Cross-domain navigation (worlds, sessions, tags, players)
5. Frontend hardening (a11y, perf, E2E, observability)

## 10) Frontend Handoff Checklist (Per Session)

1. Confirm frontend project boots (`npm install`, `npm run dev`)
2. Confirm tests baseline (`npm run test`)
3. Read this file before implementing new UI/domain work
4. Validate new features preserve layering and DTO boundaries
5. Update this document when changing a locked decision

## 11) Decision Change Log

```md
## Decision Update - YYYY-MM-DD
- Changed:
- Reason:
- Impacted areas:
```

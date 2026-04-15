# Frontend (Web)

This frontend follows `docs/frontend-architecture.md` as source of truth.

## Start

```bash
npm install
npm run dev
```

`npm run dev` runs API-integrated mode by default.

Use mock mode only when explicitly needed:

```bash
npm run dev:mock
```

## Test

```bash
npm run test
```

## Current Baseline

- Vite + React + TypeScript
- CSS Modules + shared design tokens
- dnd-kit enabled for campaign ordering
- Initial campaign-notes shell:
  - Left campaign panel (add/delete/search/reorder)
  - Center panel with create-campaign mode, world-note list, selected note markdown view, and campaign-note edit mode
  - Right panel campaign-note list (select note to open in center)
  - Bottom global tools bar

## Data Source Injection

Campaign data origin is isolated behind classes:

- `MockCampaignDataSource` for development/local workflow
- `ApiCampaignDataSource` for backend integration

Environment switches:

- `VITE_CAMPAIGN_DATA_SOURCE=mock|api` (default behavior is API unless explicitly set to `mock`)
- `VITE_API_BASE_URL=/api/v1` (default)
- `VITE_DEFAULT_WORLD_ID=<world-ulid>` for campaign create payloads in API mode

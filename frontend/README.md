# Frontend (Web)

This frontend follows `docs/frontend-architecture.md` as source of truth.

## Start

```bash
npm install
npm run dev
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
  - Center main panel placeholder with create-campaign mode
  - Right context placeholder panel
  - Bottom global tools bar

## Data Source Injection

Campaign data origin is isolated behind classes:

- `MockCampaignDataSource` for development/local workflow
- `ApiCampaignDataSource` for backend integration

Environment switches:

- `VITE_CAMPAIGN_DATA_SOURCE=mock|api`
- `VITE_API_BASE_URL=/api/v1` (default)
- `VITE_DEFAULT_WORLD_ID=<world-ulid>` for campaign create payloads in API mode

# GMN Error Token Contract

This document defines the stable `GMN_*` error-token contract emitted by database `fn_*` functions (`002_service_functions`) and consumed by repository/service error mapping.

## Contract Rules

- Tokens are emitted through PostgreSQL exceptions with `SQLSTATE P0001`.
- Token names are stable. Renames are breaking changes.
- New tokens must be added to:
  - `backend/go/src/repository/error/gmn_contract.go`
  - this document
  - service/repository mapping tests

## Tokens

### Campaign player

- `GMN_CAMPAIGN_NOT_FOUND`
- `GMN_CAMPAIGN_DELETED`
- `GMN_PLAYER_NOT_FOUND`
- `GMN_PLAYER_DELETED`
- `GMN_CAMPAIGN_PLAYER_ALREADY_ACTIVE`
- `GMN_CAMPAIGN_PLAYER_NOT_ACTIVE`

### Note owner

- `GMN_NOTE_NOT_FOUND`
- `GMN_NOTE_DELETED`
- `GMN_OWNER_NOT_FOUND_WORLD`
- `GMN_OWNER_DELETED_WORLD`
- `GMN_OWNER_NOT_FOUND_PLANE`
- `GMN_OWNER_DELETED_PLANE`
- `GMN_OWNER_NOT_FOUND_CAMPAIGN`
- `GMN_OWNER_DELETED_CAMPAIGN`
- `GMN_OWNER_NOT_FOUND_SESSION`
- `GMN_OWNER_DELETED_SESSION`
- `GMN_OWNER_NOT_FOUND_PLAYER`
- `GMN_OWNER_DELETED_PLAYER`
- `GMN_NOTE_OWNER_ALREADY_ACTIVE`
- `GMN_NOTE_OWNER_NOT_ACTIVE`

### Note tag

- `GMN_TAG_NOT_FOUND`
- `GMN_TAG_DELETED`
- `GMN_NOTE_TAG_ALREADY_ACTIVE`
- `GMN_NOTE_TAG_NOT_ACTIVE`

### Note link

- `GMN_NOTE_LINK_SELF_REFERENCE`
- `GMN_SOURCE_NOTE_NOT_FOUND`
- `GMN_SOURCE_NOTE_DELETED`
- `GMN_TARGET_NOTE_NOT_FOUND`
- `GMN_TARGET_NOTE_DELETED`
- `GMN_NOTE_LINK_ALREADY_ACTIVE`
- `GMN_NOTE_LINK_NOT_ACTIVE`

### Map note placement

- `GMN_MAP_NOTE_NOT_FOUND`
- `GMN_MAP_NOTE_DELETED`
- `GMN_MAP_NOTE_PLACEMENT_X_OUT_OF_RANGE`
- `GMN_MAP_NOTE_PLACEMENT_Y_OUT_OF_RANGE`
- `GMN_MAP_NOTE_PLACEMENT_NOT_ACTIVE`


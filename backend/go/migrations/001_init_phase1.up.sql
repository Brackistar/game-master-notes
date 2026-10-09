BEGIN;

-- =========================
-- Shared domain types
-- =========================

CREATE DOMAIN ulid AS VARCHAR(26)
    CHECK (char_length(VALUE) = 26);

-- =========================
-- Enums
-- =========================

CREATE TYPE world_status AS ENUM ('draft', 'active', 'archived');
CREATE TYPE note_type AS ENUM ('general', 'summary_note', 'map', 'character', 'location');
CREATE TYPE owner_type AS ENUM ('world', 'plane', 'campaign', 'session', 'player');
CREATE TYPE asset_type AS ENUM ('image');
CREATE TYPE note_link_type AS ENUM ('related', 'contains', 'mentions', 'depends_on', 'located_in');

-- =========================
-- Base tables
-- =========================

CREATE TABLE worlds (
    id          ulid PRIMARY KEY,
    plane_id    ulid NOT NULL,
    name        TEXT NOT NULL CHECK (char_length(trim(name)) > 0),
    description TEXT NOT NULL DEFAULT '',
    status      world_status NOT NULL DEFAULT 'draft',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    version     INTEGER NOT NULL DEFAULT 1 CHECK (version > 0)
);

CREATE TABLE planes (
    id          ulid PRIMARY KEY,
    name        TEXT NOT NULL CHECK (char_length(trim(name)) > 0),
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    version     INTEGER NOT NULL DEFAULT 1 CHECK (version > 0)
);

ALTER TABLE worlds
    ADD CONSTRAINT worlds_plane_id_fkey
    FOREIGN KEY (plane_id) REFERENCES planes(id) ON DELETE RESTRICT;

CREATE TABLE campaigns (
    id          ulid PRIMARY KEY,
    world_id    ulid NOT NULL REFERENCES worlds(id) ON DELETE RESTRICT,
    name        TEXT NOT NULL CHECK (char_length(trim(name)) > 0),
    start_date  DATE,
    end_date    DATE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    version     INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    CONSTRAINT campaigns_date_range_chk
        CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)
);

CREATE TABLE players (
    id          ulid PRIMARY KEY,
    name        TEXT NOT NULL CHECK (char_length(trim(name)) > 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    version     INTEGER NOT NULL DEFAULT 1 CHECK (version > 0)
);

CREATE TABLE sessions (
    id          ulid PRIMARY KEY,
    campaign_id ulid NOT NULL REFERENCES campaigns(id) ON DELETE RESTRICT,
    played_on   DATE,
    summary_md  TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    version     INTEGER NOT NULL DEFAULT 1 CHECK (version > 0)
);

CREATE TABLE notes (
    id            ulid PRIMARY KEY,
    title         TEXT NOT NULL CHECK (char_length(trim(title)) > 0),
    content_md    TEXT NOT NULL DEFAULT '',
    note_type     note_type NOT NULL DEFAULT 'general',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    version       INTEGER NOT NULL DEFAULT 1 CHECK (version > 0)
);

CREATE TABLE tags (
    id          ulid PRIMARY KEY,
    name        TEXT NOT NULL CHECK (char_length(trim(name)) > 0),
    campaign_id ulid REFERENCES campaigns(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    version     INTEGER NOT NULL DEFAULT 1 CHECK (version > 0)
);

CREATE TABLE note_assets (
    id           ulid PRIMARY KEY,
    note_id      ulid NOT NULL REFERENCES notes(id) ON DELETE RESTRICT,
    asset_type   asset_type NOT NULL DEFAULT 'image',
    storage_path TEXT NOT NULL CHECK (char_length(trim(storage_path)) > 0),
    mime_type    TEXT NOT NULL CHECK (char_length(trim(mime_type)) > 0),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    version      INTEGER NOT NULL DEFAULT 1 CHECK (version > 0)
);

-- =========================
-- Relationship tables
-- =========================

CREATE TABLE campaign_players (
    campaign_id ulid NOT NULL REFERENCES campaigns(id) ON DELETE RESTRICT,
    player_id   ulid NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE note_owners (
    note_id     ulid NOT NULL REFERENCES notes(id) ON DELETE RESTRICT,
    owner_type  owner_type NOT NULL,
    owner_id    ulid NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE note_tags (
    note_id     ulid NOT NULL REFERENCES notes(id) ON DELETE RESTRICT,
    tag_id      ulid NOT NULL REFERENCES tags(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE map_note_placements (
    id             ulid PRIMARY KEY,
    map_note_id    ulid NOT NULL REFERENCES notes(id) ON DELETE RESTRICT,
    target_note_id ulid NOT NULL REFERENCES notes(id) ON DELETE RESTRICT,
    x              SMALLINT NOT NULL,
    y              SMALLINT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    version        INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    CONSTRAINT map_note_placements_x_range_chk CHECK (x >= 0 AND x <= 100),
    CONSTRAINT map_note_placements_y_range_chk CHECK (y >= 0 AND y <= 100)
);

CREATE TABLE note_links (
    id             ulid PRIMARY KEY,
    source_note_id ulid NOT NULL REFERENCES notes(id) ON DELETE RESTRICT,
    target_note_id ulid NOT NULL REFERENCES notes(id) ON DELETE RESTRICT,
    link_type      note_link_type NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    version        INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    CONSTRAINT note_links_not_self_chk CHECK (source_note_id <> target_note_id)
);

-- =========================
-- Indexes
-- =========================

CREATE INDEX idx_worlds_deleted_at ON worlds (deleted_at);

CREATE INDEX idx_planes_deleted_at ON planes (deleted_at);

CREATE INDEX idx_worlds_plane_id ON worlds (plane_id);
CREATE INDEX idx_campaigns_world_id ON campaigns (world_id);
CREATE INDEX idx_campaigns_deleted_at ON campaigns (deleted_at);

CREATE INDEX idx_players_deleted_at ON players (deleted_at);

CREATE INDEX idx_sessions_campaign_id ON sessions (campaign_id);
CREATE INDEX idx_sessions_ordering ON sessions (played_on DESC, created_at DESC, id DESC);
CREATE INDEX idx_sessions_deleted_at ON sessions (deleted_at);

CREATE INDEX idx_notes_note_type ON notes (note_type);
CREATE INDEX idx_notes_deleted_at ON notes (deleted_at);

CREATE INDEX idx_tags_campaign_id ON tags (campaign_id);
CREATE INDEX idx_tags_deleted_at ON tags (deleted_at);

CREATE INDEX idx_note_assets_note_id ON note_assets (note_id);
CREATE INDEX idx_note_assets_deleted_at ON note_assets (deleted_at);

CREATE INDEX idx_campaign_players_campaign_id ON campaign_players (campaign_id);
CREATE INDEX idx_campaign_players_player_id ON campaign_players (player_id);
CREATE UNIQUE INDEX ux_campaign_players_active
    ON campaign_players (campaign_id, player_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_note_owners_note_id ON note_owners (note_id);
CREATE INDEX idx_note_owners_owner_lookup ON note_owners (owner_type, owner_id);
CREATE UNIQUE INDEX ux_note_owners_active
    ON note_owners (note_id, owner_type, owner_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_note_tags_note_id ON note_tags (note_id);
CREATE INDEX idx_note_tags_tag_id ON note_tags (tag_id);
CREATE UNIQUE INDEX ux_note_tags_active
    ON note_tags (note_id, tag_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX ux_tags_global_name_active
    ON tags (lower(name))
    WHERE campaign_id IS NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX ux_tags_campaign_name_active
    ON tags (campaign_id, lower(name))
    WHERE campaign_id IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX idx_map_note_placements_map_note_id ON map_note_placements (map_note_id);
CREATE INDEX idx_map_note_placements_target_note_id ON map_note_placements (target_note_id);
CREATE UNIQUE INDEX ux_map_note_placements_active
    ON map_note_placements (map_note_id, target_note_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_note_links_source_note_id ON note_links (source_note_id);
CREATE INDEX idx_note_links_target_note_id ON note_links (target_note_id);
CREATE UNIQUE INDEX ux_note_links_active
    ON note_links (source_note_id, target_note_id, link_type)
    WHERE deleted_at IS NULL;

-- =========================
-- Triggers / functions
-- =========================

CREATE OR REPLACE FUNCTION soft_delete_note_links_on_note_delete()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
        UPDATE note_links
        SET
            deleted_at = NEW.deleted_at,
            updated_at = now(),
            version = version + 1
        WHERE deleted_at IS NULL
          AND (source_note_id = NEW.id OR target_note_id = NEW.id);
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_soft_delete_note_links_on_note_delete
AFTER UPDATE OF deleted_at ON notes
FOR EACH ROW
EXECUTE FUNCTION soft_delete_note_links_on_note_delete();

COMMIT;

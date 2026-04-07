BEGIN;

DROP TRIGGER IF EXISTS trg_soft_delete_note_links_on_note_delete ON notes;
DROP FUNCTION IF EXISTS soft_delete_note_links_on_note_delete();

DROP TABLE IF EXISTS note_links;
DROP TABLE IF EXISTS map_note_placements;
DROP TABLE IF EXISTS note_tags;
DROP TABLE IF EXISTS note_owners;
DROP TABLE IF EXISTS campaign_players;
DROP TABLE IF EXISTS note_assets;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS notes;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS planes;
DROP TABLE IF EXISTS worlds;

DROP TYPE IF EXISTS note_link_type;
DROP TYPE IF EXISTS asset_type;
DROP TYPE IF EXISTS owner_type;
DROP TYPE IF EXISTS note_type;
DROP TYPE IF EXISTS world_status;

DROP DOMAIN IF EXISTS ulid;

COMMIT;

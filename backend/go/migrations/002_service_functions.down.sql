BEGIN;

DROP FUNCTION IF EXISTS fn_remove_map_note_placement(ulid, ulid);
DROP FUNCTION IF EXISTS fn_upsert_map_note_placement(ulid, ulid, ulid, SMALLINT, SMALLINT);
DROP FUNCTION IF EXISTS fn_remove_note_link(ulid, ulid, note_link_type);
DROP FUNCTION IF EXISTS fn_add_note_link(ulid, ulid, ulid, note_link_type);
DROP FUNCTION IF EXISTS fn_remove_note_tag(ulid, ulid);
DROP FUNCTION IF EXISTS fn_add_note_tag(ulid, ulid);
DROP FUNCTION IF EXISTS fn_remove_note_owner(ulid, owner_type, ulid);
DROP FUNCTION IF EXISTS fn_add_note_owner(ulid, owner_type, ulid);
DROP FUNCTION IF EXISTS fn_remove_player_from_campaign(ulid, ulid);
DROP FUNCTION IF EXISTS fn_add_player_to_campaign(ulid, ulid);

COMMIT;

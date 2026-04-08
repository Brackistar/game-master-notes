BEGIN;

-- ==========================================
-- Campaign players
-- ==========================================

CREATE OR REPLACE FUNCTION fn_add_player_to_campaign(
    p_campaign_id ulid,
    p_player_id ulid
)
RETURNS campaign_players
LANGUAGE plpgsql
AS $$
DECLARE
    v_row campaign_players%ROWTYPE;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM campaigns WHERE id = p_campaign_id) THEN
        RAISE EXCEPTION 'GMN_CAMPAIGN_NOT_FOUND' USING ERRCODE = 'P0001';
    END IF;
    IF EXISTS (SELECT 1 FROM campaigns WHERE id = p_campaign_id AND deleted_at IS NOT NULL) THEN
        RAISE EXCEPTION 'GMN_CAMPAIGN_DELETED' USING ERRCODE = 'P0001';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM players WHERE id = p_player_id) THEN
        RAISE EXCEPTION 'GMN_PLAYER_NOT_FOUND' USING ERRCODE = 'P0001';
    END IF;
    IF EXISTS (SELECT 1 FROM players WHERE id = p_player_id AND deleted_at IS NOT NULL) THEN
        RAISE EXCEPTION 'GMN_PLAYER_DELETED' USING ERRCODE = 'P0001';
    END IF;

    SELECT *
    INTO v_row
    FROM campaign_players
    WHERE campaign_id = p_campaign_id
      AND player_id = p_player_id
      AND deleted_at IS NULL
    LIMIT 1;

    IF FOUND THEN
        RAISE EXCEPTION 'GMN_CAMPAIGN_PLAYER_ALREADY_ACTIVE' USING ERRCODE = 'P0001';
    END IF;

    UPDATE campaign_players
    SET
        deleted_at = NULL,
        updated_at = now()
    WHERE ctid = (
        SELECT ctid
        FROM campaign_players
        WHERE campaign_id = p_campaign_id
          AND player_id = p_player_id
          AND deleted_at IS NOT NULL
        ORDER BY updated_at DESC
        LIMIT 1
    )
    RETURNING *
    INTO v_row;

    IF FOUND THEN
        RETURN v_row;
    END IF;

    INSERT INTO campaign_players (
        campaign_id,
        player_id,
        created_at,
        updated_at,
        deleted_at
    ) VALUES (
        p_campaign_id,
        p_player_id,
        now(),
        now(),
        NULL
    )
    RETURNING *
    INTO v_row;

    RETURN v_row;
END;
$$;

CREATE OR REPLACE FUNCTION fn_remove_player_from_campaign(
    p_campaign_id ulid,
    p_player_id ulid
)
RETURNS campaign_players
LANGUAGE plpgsql
AS $$
DECLARE
    v_row campaign_players%ROWTYPE;
BEGIN
    UPDATE campaign_players
    SET
        deleted_at = now(),
        updated_at = now()
    WHERE campaign_id = p_campaign_id
      AND player_id = p_player_id
      AND deleted_at IS NULL
    RETURNING *
    INTO v_row;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'GMN_CAMPAIGN_PLAYER_NOT_ACTIVE' USING ERRCODE = 'P0001';
    END IF;

    RETURN v_row;
END;
$$;

-- ==========================================
-- Note owners
-- ==========================================

CREATE OR REPLACE FUNCTION fn_add_note_owner(
    p_note_id ulid,
    p_owner_type owner_type,
    p_owner_id ulid
)
RETURNS note_owners
LANGUAGE plpgsql
AS $$
DECLARE
    v_row note_owners%ROWTYPE;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM notes WHERE id = p_note_id) THEN
        RAISE EXCEPTION 'GMN_NOTE_NOT_FOUND' USING ERRCODE = 'P0001';
    END IF;
    IF EXISTS (SELECT 1 FROM notes WHERE id = p_note_id AND deleted_at IS NOT NULL) THEN
        RAISE EXCEPTION 'GMN_NOTE_DELETED' USING ERRCODE = 'P0001';
    END IF;

    IF p_owner_type = 'world' THEN
        IF NOT EXISTS (SELECT 1 FROM worlds WHERE id = p_owner_id) THEN
            RAISE EXCEPTION 'GMN_OWNER_NOT_FOUND_WORLD' USING ERRCODE = 'P0001';
        END IF;
        IF EXISTS (SELECT 1 FROM worlds WHERE id = p_owner_id AND deleted_at IS NOT NULL) THEN
            RAISE EXCEPTION 'GMN_OWNER_DELETED_WORLD' USING ERRCODE = 'P0001';
        END IF;
    ELSIF p_owner_type = 'plane' THEN
        IF NOT EXISTS (SELECT 1 FROM planes WHERE id = p_owner_id) THEN
            RAISE EXCEPTION 'GMN_OWNER_NOT_FOUND_PLANE' USING ERRCODE = 'P0001';
        END IF;
        IF EXISTS (SELECT 1 FROM planes WHERE id = p_owner_id AND deleted_at IS NOT NULL) THEN
            RAISE EXCEPTION 'GMN_OWNER_DELETED_PLANE' USING ERRCODE = 'P0001';
        END IF;
    ELSIF p_owner_type = 'campaign' THEN
        IF NOT EXISTS (SELECT 1 FROM campaigns WHERE id = p_owner_id) THEN
            RAISE EXCEPTION 'GMN_OWNER_NOT_FOUND_CAMPAIGN' USING ERRCODE = 'P0001';
        END IF;
        IF EXISTS (SELECT 1 FROM campaigns WHERE id = p_owner_id AND deleted_at IS NOT NULL) THEN
            RAISE EXCEPTION 'GMN_OWNER_DELETED_CAMPAIGN' USING ERRCODE = 'P0001';
        END IF;
    ELSIF p_owner_type = 'session' THEN
        IF NOT EXISTS (SELECT 1 FROM sessions WHERE id = p_owner_id) THEN
            RAISE EXCEPTION 'GMN_OWNER_NOT_FOUND_SESSION' USING ERRCODE = 'P0001';
        END IF;
        IF EXISTS (SELECT 1 FROM sessions WHERE id = p_owner_id AND deleted_at IS NOT NULL) THEN
            RAISE EXCEPTION 'GMN_OWNER_DELETED_SESSION' USING ERRCODE = 'P0001';
        END IF;
    ELSIF p_owner_type = 'player' THEN
        IF NOT EXISTS (SELECT 1 FROM players WHERE id = p_owner_id) THEN
            RAISE EXCEPTION 'GMN_OWNER_NOT_FOUND_PLAYER' USING ERRCODE = 'P0001';
        END IF;
        IF EXISTS (SELECT 1 FROM players WHERE id = p_owner_id AND deleted_at IS NOT NULL) THEN
            RAISE EXCEPTION 'GMN_OWNER_DELETED_PLAYER' USING ERRCODE = 'P0001';
        END IF;
    END IF;

    SELECT *
    INTO v_row
    FROM note_owners
    WHERE note_id = p_note_id
      AND owner_type = p_owner_type
      AND owner_id = p_owner_id
      AND deleted_at IS NULL
    LIMIT 1;

    IF FOUND THEN
        RAISE EXCEPTION 'GMN_NOTE_OWNER_ALREADY_ACTIVE' USING ERRCODE = 'P0001';
    END IF;

    UPDATE note_owners
    SET
        deleted_at = NULL,
        updated_at = now()
    WHERE ctid = (
        SELECT ctid
        FROM note_owners
        WHERE note_id = p_note_id
          AND owner_type = p_owner_type
          AND owner_id = p_owner_id
          AND deleted_at IS NOT NULL
        ORDER BY updated_at DESC
        LIMIT 1
    )
    RETURNING *
    INTO v_row;

    IF FOUND THEN
        RETURN v_row;
    END IF;

    INSERT INTO note_owners (
        note_id,
        owner_type,
        owner_id,
        created_at,
        updated_at,
        deleted_at
    ) VALUES (
        p_note_id,
        p_owner_type,
        p_owner_id,
        now(),
        now(),
        NULL
    )
    RETURNING *
    INTO v_row;

    RETURN v_row;
END;
$$;

CREATE OR REPLACE FUNCTION fn_remove_note_owner(
    p_note_id ulid,
    p_owner_type owner_type,
    p_owner_id ulid
)
RETURNS note_owners
LANGUAGE plpgsql
AS $$
DECLARE
    v_row note_owners%ROWTYPE;
BEGIN
    UPDATE note_owners
    SET
        deleted_at = now(),
        updated_at = now()
    WHERE note_id = p_note_id
      AND owner_type = p_owner_type
      AND owner_id = p_owner_id
      AND deleted_at IS NULL
    RETURNING *
    INTO v_row;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'GMN_NOTE_OWNER_NOT_ACTIVE' USING ERRCODE = 'P0001';
    END IF;

    RETURN v_row;
END;
$$;

-- ==========================================
-- Note tags
-- ==========================================

CREATE OR REPLACE FUNCTION fn_add_note_tag(
    p_note_id ulid,
    p_tag_id ulid
)
RETURNS note_tags
LANGUAGE plpgsql
AS $$
DECLARE
    v_row note_tags%ROWTYPE;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM notes WHERE id = p_note_id) THEN
        RAISE EXCEPTION 'GMN_NOTE_NOT_FOUND' USING ERRCODE = 'P0001';
    END IF;
    IF EXISTS (SELECT 1 FROM notes WHERE id = p_note_id AND deleted_at IS NOT NULL) THEN
        RAISE EXCEPTION 'GMN_NOTE_DELETED' USING ERRCODE = 'P0001';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM tags WHERE id = p_tag_id) THEN
        RAISE EXCEPTION 'GMN_TAG_NOT_FOUND' USING ERRCODE = 'P0001';
    END IF;
    IF EXISTS (SELECT 1 FROM tags WHERE id = p_tag_id AND deleted_at IS NOT NULL) THEN
        RAISE EXCEPTION 'GMN_TAG_DELETED' USING ERRCODE = 'P0001';
    END IF;

    SELECT *
    INTO v_row
    FROM note_tags
    WHERE note_id = p_note_id
      AND tag_id = p_tag_id
      AND deleted_at IS NULL
    LIMIT 1;

    IF FOUND THEN
        RAISE EXCEPTION 'GMN_NOTE_TAG_ALREADY_ACTIVE' USING ERRCODE = 'P0001';
    END IF;

    UPDATE note_tags
    SET
        deleted_at = NULL,
        updated_at = now()
    WHERE ctid = (
        SELECT ctid
        FROM note_tags
        WHERE note_id = p_note_id
          AND tag_id = p_tag_id
          AND deleted_at IS NOT NULL
        ORDER BY updated_at DESC
        LIMIT 1
    )
    RETURNING *
    INTO v_row;

    IF FOUND THEN
        RETURN v_row;
    END IF;

    INSERT INTO note_tags (
        note_id,
        tag_id,
        created_at,
        updated_at,
        deleted_at
    ) VALUES (
        p_note_id,
        p_tag_id,
        now(),
        now(),
        NULL
    )
    RETURNING *
    INTO v_row;

    RETURN v_row;
END;
$$;

CREATE OR REPLACE FUNCTION fn_remove_note_tag(
    p_note_id ulid,
    p_tag_id ulid
)
RETURNS note_tags
LANGUAGE plpgsql
AS $$
DECLARE
    v_row note_tags%ROWTYPE;
BEGIN
    UPDATE note_tags
    SET
        deleted_at = now(),
        updated_at = now()
    WHERE note_id = p_note_id
      AND tag_id = p_tag_id
      AND deleted_at IS NULL
    RETURNING *
    INTO v_row;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'GMN_NOTE_TAG_NOT_ACTIVE' USING ERRCODE = 'P0001';
    END IF;

    RETURN v_row;
END;
$$;

-- ==========================================
-- Note links
-- ==========================================

CREATE OR REPLACE FUNCTION fn_add_note_link(
    p_id ulid,
    p_source_note_id ulid,
    p_target_note_id ulid,
    p_link_type note_link_type
)
RETURNS note_links
LANGUAGE plpgsql
AS $$
DECLARE
    v_row note_links%ROWTYPE;
BEGIN
    IF p_source_note_id = p_target_note_id THEN
        RAISE EXCEPTION 'GMN_NOTE_LINK_SELF_REFERENCE' USING ERRCODE = 'P0001';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM notes WHERE id = p_source_note_id) THEN
        RAISE EXCEPTION 'GMN_SOURCE_NOTE_NOT_FOUND' USING ERRCODE = 'P0001';
    END IF;
    IF EXISTS (SELECT 1 FROM notes WHERE id = p_source_note_id AND deleted_at IS NOT NULL) THEN
        RAISE EXCEPTION 'GMN_SOURCE_NOTE_DELETED' USING ERRCODE = 'P0001';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM notes WHERE id = p_target_note_id) THEN
        RAISE EXCEPTION 'GMN_TARGET_NOTE_NOT_FOUND' USING ERRCODE = 'P0001';
    END IF;
    IF EXISTS (SELECT 1 FROM notes WHERE id = p_target_note_id AND deleted_at IS NOT NULL) THEN
        RAISE EXCEPTION 'GMN_TARGET_NOTE_DELETED' USING ERRCODE = 'P0001';
    END IF;

    SELECT *
    INTO v_row
    FROM note_links
    WHERE source_note_id = p_source_note_id
      AND target_note_id = p_target_note_id
      AND link_type = p_link_type
      AND deleted_at IS NULL
    LIMIT 1;

    IF FOUND THEN
        RAISE EXCEPTION 'GMN_NOTE_LINK_ALREADY_ACTIVE' USING ERRCODE = 'P0001';
    END IF;

    SELECT *
    INTO v_row
    FROM note_links
    WHERE source_note_id = p_source_note_id
      AND target_note_id = p_target_note_id
      AND link_type = p_link_type
      AND deleted_at IS NOT NULL
    ORDER BY updated_at DESC, id DESC
    LIMIT 1;

    IF FOUND THEN
        UPDATE note_links
        SET
            deleted_at = NULL,
            updated_at = now(),
            version = version + 1
        WHERE id = v_row.id
        RETURNING *
        INTO v_row;

        RETURN v_row;
    END IF;

    INSERT INTO note_links (
        id,
        source_note_id,
        target_note_id,
        link_type,
        created_at,
        updated_at,
        deleted_at,
        version
    ) VALUES (
        p_id,
        p_source_note_id,
        p_target_note_id,
        p_link_type,
        now(),
        now(),
        NULL,
        1
    )
    RETURNING *
    INTO v_row;

    RETURN v_row;
END;
$$;

CREATE OR REPLACE FUNCTION fn_remove_note_link(
    p_source_note_id ulid,
    p_target_note_id ulid,
    p_link_type note_link_type
)
RETURNS note_links
LANGUAGE plpgsql
AS $$
DECLARE
    v_row note_links%ROWTYPE;
BEGIN
    UPDATE note_links
    SET
        deleted_at = now(),
        updated_at = now(),
        version = version + 1
    WHERE source_note_id = p_source_note_id
      AND target_note_id = p_target_note_id
      AND link_type = p_link_type
      AND deleted_at IS NULL
    RETURNING *
    INTO v_row;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'GMN_NOTE_LINK_NOT_ACTIVE' USING ERRCODE = 'P0001';
    END IF;

    RETURN v_row;
END;
$$;

-- ==========================================
-- Map note placements
-- ==========================================

CREATE OR REPLACE FUNCTION fn_upsert_map_note_placement(
    p_id ulid,
    p_map_note_id ulid,
    p_target_note_id ulid,
    p_x SMALLINT,
    p_y SMALLINT
)
RETURNS map_note_placements
LANGUAGE plpgsql
AS $$
DECLARE
    v_row map_note_placements%ROWTYPE;
BEGIN
    IF p_x < 0 OR p_x > 100 THEN
        RAISE EXCEPTION 'GMN_MAP_NOTE_PLACEMENT_X_OUT_OF_RANGE' USING ERRCODE = 'P0001';
    END IF;
    IF p_y < 0 OR p_y > 100 THEN
        RAISE EXCEPTION 'GMN_MAP_NOTE_PLACEMENT_Y_OUT_OF_RANGE' USING ERRCODE = 'P0001';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM notes WHERE id = p_map_note_id) THEN
        RAISE EXCEPTION 'GMN_MAP_NOTE_NOT_FOUND' USING ERRCODE = 'P0001';
    END IF;
    IF EXISTS (SELECT 1 FROM notes WHERE id = p_map_note_id AND deleted_at IS NOT NULL) THEN
        RAISE EXCEPTION 'GMN_MAP_NOTE_DELETED' USING ERRCODE = 'P0001';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM notes WHERE id = p_target_note_id) THEN
        RAISE EXCEPTION 'GMN_TARGET_NOTE_NOT_FOUND' USING ERRCODE = 'P0001';
    END IF;
    IF EXISTS (SELECT 1 FROM notes WHERE id = p_target_note_id AND deleted_at IS NOT NULL) THEN
        RAISE EXCEPTION 'GMN_TARGET_NOTE_DELETED' USING ERRCODE = 'P0001';
    END IF;

    SELECT *
    INTO v_row
    FROM map_note_placements
    WHERE map_note_id = p_map_note_id
      AND target_note_id = p_target_note_id
      AND deleted_at IS NULL
    LIMIT 1;

    IF FOUND THEN
        UPDATE map_note_placements
        SET
            x = p_x,
            y = p_y,
            updated_at = now(),
            version = version + 1
        WHERE id = v_row.id
        RETURNING *
        INTO v_row;

        RETURN v_row;
    END IF;

    SELECT *
    INTO v_row
    FROM map_note_placements
    WHERE map_note_id = p_map_note_id
      AND target_note_id = p_target_note_id
      AND deleted_at IS NOT NULL
    ORDER BY updated_at DESC, id DESC
    LIMIT 1;

    IF FOUND THEN
        UPDATE map_note_placements
        SET
            x = p_x,
            y = p_y,
            deleted_at = NULL,
            updated_at = now(),
            version = version + 1
        WHERE id = v_row.id
        RETURNING *
        INTO v_row;

        RETURN v_row;
    END IF;

    INSERT INTO map_note_placements (
        id,
        map_note_id,
        target_note_id,
        x,
        y,
        created_at,
        updated_at,
        deleted_at,
        version
    ) VALUES (
        p_id,
        p_map_note_id,
        p_target_note_id,
        p_x,
        p_y,
        now(),
        now(),
        NULL,
        1
    )
    RETURNING *
    INTO v_row;

    RETURN v_row;
END;
$$;

CREATE OR REPLACE FUNCTION fn_remove_map_note_placement(
    p_map_note_id ulid,
    p_target_note_id ulid
)
RETURNS map_note_placements
LANGUAGE plpgsql
AS $$
DECLARE
    v_row map_note_placements%ROWTYPE;
BEGIN
    UPDATE map_note_placements
    SET
        deleted_at = now(),
        updated_at = now(),
        version = version + 1
    WHERE map_note_id = p_map_note_id
      AND target_note_id = p_target_note_id
      AND deleted_at IS NULL
    RETURNING *
    INTO v_row;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'GMN_MAP_NOTE_PLACEMENT_NOT_ACTIVE' USING ERRCODE = 'P0001';
    END IF;

    RETURN v_row;
END;
$$;

COMMIT;

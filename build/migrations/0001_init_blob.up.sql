CREATE TABLE blobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    metadata JSONB,
    value BYTEA COMPRESSION LZ4
);

CREATE UNLOGGED TABLE blobs_cache (
    id UUID PRIMARY KEY REFERENCES blobs(id) ON DELETE CASCADE,
    metadata JSONB,
    value BYTEA,
    expires_at TIMESTAMPTZ NOT NULL
);

-------------
-- Indexes --
-------------

CREATE INDEX i_blobs_cache_expiry ON blobs_cache (expires_at);

--------------
-- Policies --
--------------

-- nobody should write to the cache directly, only postgres itself.
REVOKE INSERT, UPDATE, DELETE ON blobs_cache FROM PUBLIC;

-- Trigger function on blobs table that clears matching cache entries
-- when a blob is updated or deleted
CREATE OR REPLACE FUNCTION clear_blobs_cache() RETURNS trigger AS $$
BEGIN
    DELETE FROM blobs_cache WHERE id = NEW.id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach the trigger function to blobs for UPDATE events
CREATE TRIGGER blobs_update_cache_trigger
AFTER UPDATE ON blobs
FOR EACH ROW EXECUTE FUNCTION clear_blobs_cache ();

-- Function to perform cache lookup
-- This function takes:
--      p_id        : blob's UUID,
-- It returns the BYTEA value from the cache or main table
CREATE OR REPLACE FUNCTION get_blob(
    p_id UUID
) RETURNS TABLE (
    id UUID,
    metadata JSONB,
    value BYTEA
) AS $$
DECLARE
    v_record RECORD;
    v_new_size BIGINT;
    v_current_size BIGINT;
    v_deleted_size BIGINT;
    v_max_cache BIGINT;
    v_cache_ttl INTERVAL;
BEGIN
    -- load necessary values from admin table
    SELECT (a.value->>'maxSize')::BIGINT, (a.value->>'ttl')::INTERVAL
    INTO v_max_cache, v_cache_ttl
    FROM admin a
    WHERE key = 'blobs_cache_config';

    IF NOT FOUND THEN 
        RAISE WARNING 'blobs_cache_config data not found, using defaults (TTL: 1h, size: 1G)';
        v_max_cache := 1<<30;
        v_cache_ttl := '1 hour';
        INSERT INTO admin (key, value)
        VALUES ('blobs_cache_config', json_build_object(
            'maxSize', v_max_cache,
            'ttl', v_cache_ttl
        ));
    END IF;

    -- try to find the item in the cache
    SELECT bc.id, bc.metadata, bc.value INTO v_record FROM blobs_cache bc WHERE bc.id = p_id;
    IF FOUND THEN
        UPDATE blobs_cache bc
        SET expires_at = NOW() + v_cache_ttl
        WHERE bc.id = p_id;
        RETURN QUERY SELECT v_record.id, v_record.metadata, v_record.value;
    ELSE
        -- If not in cache, fetch from main table
        SELECT b.id, b.metadata, b.value INTO v_record FROM blobs b WHERE b.id = p_id;
        IF NOT FOUND THEN
            RETURN QUERY SELECT NULL, NULL, NULL; -- Doesn't exist
        END IF;

        -- Now we have to determine how we fit things into the cache
        v_new_size := octet_length(v_record.value);
        -- If the new object is too big (>60% of cache) then we don't 
        -- try to store it, instead just return it and be done
        if v_new_size > v_max_cache * 6 / 10 THEN
            RETURN QUERY SELECT v_record.id, v_record.metadata, v_record.value;
        END IF;

        SELECT COALESCE(SUM(octet_length(v_record.value)), 0) INTO v_current_size FROM blobs_cache;

        -- While adding the new object would exceed the maximum cache
        -- size, we delete entries based on LRU.
        WHILE (v_current_size + v_new_size) > v_max_cache LOOP
            -- LRU here is synonymous with "soonest to expire"
            DELETE FROM blobs_cache
            WHERE id = (
                SELECT id FROM blobs_cache
                ORDER BY expires_at ASC
                LIMIT 1
            )
            RETURNING octet_length(value) INTO v_deleted_size;
            v_current_size := v_current_size - v_deleted_size;
        END LOOP;

        -- Insert the new cache entry with fresh expiration timestamp
        INSERT INTO blobs_cache (id, metadata, value, expires_at)
        VALUES (v_record.id, v_record.metadata, v_record.value, NOW() + v_cache_ttl);
        
        RETURN QUERY SELECT v_record.id, v_record.metadata, v_record.value;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create periodic cleanup function to remove expired entries
CREATE OR REPLACE PROCEDURE blob_cache_clean_expired() AS $$
DECLARE
    v_deleted_count NUMERIC;
BEGIN
    BEGIN
        DELETE FROM blobs_cache WHERE expires_at < now ()
        RETURNING 1 INTO v_deleted_count;
    EXCEPTION
        WHEN OTHERS THEN
            ROLLBACK;
            RAISE NOTICE 'Cleanup failed: %', SQLERRM;
            RETURN;
    END;
    RAISE INFO 'Cleaned: % items', COALESCE(v_deleted_count, 0);
END;
$$ LANGUAGE plpgsql;
CREATE TABLE blobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    val BYTEA
);

CREATE UNLOGGED TABLE blobs_cache (
    id UUID PRIMARY KEY REFERENCES blobs(id) ON DELETE CASCADE,
    val BYTEA,
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
) RETURNS BYTEA AS $$
DECLARE
    v_blob BYTEA;
    v_cached RECORD;
    v_new_size BIGINT;
    v_current_size BIGINT;
    v_deleted_size BIGINT;
    v_max_cache BIGINT;
    v_cache_ttl INTERVAL;
BEGIN
    -- load necessary values from admin table
    SELECT (value->>'maxSize')::BIGINT, (value->>'ttl')::INTERVAL
    INTO v_max_cache, v_cache_ttl
    FROM admin
    WHERE key = 'blobs_cache_config';

    IF NOT FOUND THEN 
        RAISE WARNING 'blobs_cache_config data not found, using defaults (TTL: 1h, size: 1G)';
        v_max_cache := 1073741824;
        v_cache_ttl := '1 hour';
        INSERT INTO admin (key, value)
        VALUES ('blobs_cache_config', json_build_object(
            'maxSize', v_max_cache,
            'ttl', v_cache_ttl
        ));
    END IF;

    -- try to find the item in the cache
    SELECT * INTO v_cached FROM blobs_cache WHERE id = p_id;
    IF FOUND THEN
        UPDATE blobs_cache
        SET expires_at = now () + v_cache_ttl
        WHERE id = p_id;
        RETURN v_cached.val;
    ELSE
        -- If not in cache, fetch from main table
        SELECT val INTO v_blob FROM blobs WHERE id = p_id;
        IF NOT FOUND THEN
            RETURN NULL; -- Doesn't exist
        END IF;

        -- Now we have to determine how we fit things into the cache
        v_new_size := octet_length(v_blob);
        -- If the new object is too big (>60% of cache) then we don't 
        -- try to store it, instead just return it and be done
        if v_new_size > v_max_cache * 6 / 10 THEN
            RETURN v_blob;
        END IF;

        SELECT COALESCE(SUM(octet_length(val)), 0) INTO v_current_size FROM blobs_cache;

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
            RETURNING octet_length(val) INTO v_deleted_size;
            v_current_size := v_current_size - v_deleted_size;
        END LOOP;

        -- Insert the new cache entry with fresh expiration timestamp
        INSERT INTO blobs_cache (id, val, expires_at)
        VALUES (p_id, v_blob, now() + v_cache_ttl);
        
        RETURN v_blob;
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

-- NOTE: this requires the pg_cron extension
-- https://github.com/citusdata/pg_cron
SELECT cron.schedule_in_database(
    'blob-cache-clean-expired', 
    '*/15 * * * *', 
    'CALL blob_cache_clean_expired()', 
    current_database()
);
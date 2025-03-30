CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    -- use regex as github ids are numeric. for now.
    github_id TEXT UNIQUE CHECK (github_id ~ '^\d+$'),
    display_name VARCHAR(32),
    user_handle TEXT UNIQUE CHECK (user_handle ~ ^(.{2,32})#\d{4}$),
    pronouns VARCHAR(16),
    email TEXT,
    avatar UUID REFERENCES blobs(id),
    superuser BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-------------
-- Indexes --
-------------

CREATE UNIQUE INDEX i_users_ghid ON users (github_id);
CREATE UNIQUE INDEX i_users_userhandle ON users (user_handle);
CREATE INDEX i_users_name ON users USING GIN (to_tsvector('english', display_name));
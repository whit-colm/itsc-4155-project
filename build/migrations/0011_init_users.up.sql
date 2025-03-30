CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    -- use regex as github ids are numeric. for now.
    github_id TEXT UNIQUE CHECK (github_id ~ '^\d+$'),
    display_name VARCHAR(32),
    handle VARCHAR(32) NOT NULL CHECK (length(handle) BETWEEN 2 AND 32),
    discriminator SMALLINT NOT NULL CHECK (discriminator BETWEEN 0 AND 9999),
    pronouns VARCHAR(16),
    email TEXT,
    avatar UUID REFERENCES blobs(id),
    superuser BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_username UNIQUE (handle, discriminator),

    CONSTRAINT valid_handle_chars CHECK (
        handle ~ '^[^@#\n\s][^@#\n]*[^@#\n\s]$'
    )
);

-------------
-- Indexes --
-------------

CREATE UNIQUE INDEX i_users_ghid ON users (github_id);
CREATE UNIQUE INDEX i_users_full_username ON users (
    (handle || '#' || lpad(discriminator::TEXT, 4, '0'))
);
CREATE INDEX i_users_name ON users USING GIN (to_tsvector('english', display_name));
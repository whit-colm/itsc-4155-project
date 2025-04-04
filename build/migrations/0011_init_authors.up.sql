-- Create authors table
CREATE TABLE authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    family_name TEXT NOT NULL,
    given_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-------------
-- Indexes --
-------------

CREATE INDEX i_author_full_name ON authors (given_name, family_name);
CREATE INDEX i_author_family_name ON authors (family_name);
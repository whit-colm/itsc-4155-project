-- Create authors table
CREATE TABLE authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    family_name TEXT NOT NULL,
    given_name TEXT,
    bio TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE author_identifiers (
    author_id UUID NOT NULL REFERENCES authors (id) ON DELETE CASCADE,
    type CHAR(6) NOT NULL CHECK (type IN ('orcid', 'viaf', 'opnlib')),
    identifier TEXT NOT NULL,
    PRIMARY KEY (type, identifier)
);

-------------
-- Indexes --
-------------

CREATE INDEX i_authors_full_name ON authors (TRIM((given_name || ' ' || family_name)));
CREATE INDEX i_authors_family_name ON authors (family_name);

CREATE INDEX i_authors_search ON authors
USING bm25 (id, family_name, given_name, bio)
WITH (key_field='id');

--------------
-- Triggers --
--------------

CREATE TRIGGER t_authors_set_updated_at
BEFORE UPDATE ON authors
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();
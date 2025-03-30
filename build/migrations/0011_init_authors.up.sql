-- Create authors table
CREATE TABLE authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    familyname TEXT NOT NULL,
    givenname TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-------------
-- Indexes --
-------------

CREATE INDEX i_author_full_name ON authors (givenname, familyname);
CREATE INDEX i_author_family_name ON authors (familyname);
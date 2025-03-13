-- author indexes
CREATE INDEX i_author_full_name ON authors (givenname, familyname);
CREATE INDEX i_author_family_name ON authors (familyname);
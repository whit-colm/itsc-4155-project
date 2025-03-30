-- Create admin table. this is single things that need to be preserved
-- but are not necessarily structured
CREATE TABLE admin (
    id SERIAL PRIMARY KEY,
    key TEXT UNIQUE NOT NULL,
    value JSONB
);

-- General scratchpad
CREATE UNLOGGED TABLE scratchpad (
    id SERIAL PRIMARY KEY,
    key TEXT UNIQUE NOT NULL,
    value JSONB
);

-------------
-- Indexes --
-------------

CREATE UNIQUE INDEX i_admin_key ON admin (key);
CREATE UNIQUE INDEX i_scratchpad_key ON scratchpad (key);

----------------
-- Protection --
----------------

-- Empty, should maybe consider doing that though because leaving
-- random *this* out for the taking seems sketch?
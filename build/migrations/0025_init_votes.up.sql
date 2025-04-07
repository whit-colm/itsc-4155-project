CREATE TABLE votes (
    comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    vote SMALLINT CHECK (vote IN (-1, 1)),
    PRIMARY KEY (comment_id, user_id)
);

-------------
-- Indexes --
-------------

CREATE INDEX i_votes_user ON votes(user_id);
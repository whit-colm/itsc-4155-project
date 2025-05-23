CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    book_id UUID REFERENCES books(id) ON DELETE CASCADE,
    poster_id UUID REFERENCES users(id) ON DELETE SET NULL,
    body TEXT,
    rating REAL,
    parent_comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    vote_total INTEGER NOT NULL DEFAULT 0,
    deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_rating CHECK (
        (rating IS NULL) OR
        (rating >= 0.0 AND rating <= 1.0)
    ),
    CONSTRAINT review_xor_reply_xor_deleted CHECK (
        (rating IS NOT NULL AND parent_comment_id IS NULL) OR
        (rating IS NULL AND parent_comment_id IS NOT NULL) OR
        (rating IS NULL AND parent_comment_id IS NULL AND deleted IS true)
    )
);

-------------
-- Indexes --
-------------

CREATE UNIQUE INDEX i_one_user_one_review ON comments (poster_id, book_id)
    WHERE parent_comment_id IS NULL;

CREATE INDEX i_comments_books ON comments (book_id);
CREATE INDEX i_comments_parent ON comments (parent_comment_id);
CREATE INDEX i_comments_user ON comments (poster_id);

CREATE INDEX i_comments_search ON comments 
USING bm25 (
    id,
    book_id,
    poster_id,
    body,
    rating,
    parent_comment_id,
    vote_total,
    deleted,
    created_at,
    updated_at
) WITH (key_field='id');

--------------
-- Triggers --
--------------

CREATE TRIGGER t_comments_set_updated_at
BEFORE UPDATE ON comments
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();
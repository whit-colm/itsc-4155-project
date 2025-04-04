CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    book UUID REFERENCES books(id) ON DELETE CASCADE,
    poster UUID REFERENCES users(id) ON DELETE SET NULL,
    rating REAL,
    parent_comment UUID REFERENCES comments(id) ON DELETE CASCADE,
    votes INTEGER default 0,
    deleted BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_rating CHECK (
        (rating IS NULL) OR
        (rating >= 0.0 AND rating < 1.0)
    ),
    CONSTRAINT one_user_one_review UNIQUE (user, book)
        WHERE parent_comment IS NULL,
    CONSTRAINT review_xor_reply CHECK (
        (rating IS NOT NULL AND parent_comment IS NULL) OR
        (rating IS NULL AND parent_comment IS NOT NULL)
    )
);

-------------
-- Indexes --
-------------

CREATE INDEX i_comments_books ON comments(book);
CREATE INDEX i_comments_parent ON comments(parent_comment);
CREATE INDEX i_comments_user ON comments(author);
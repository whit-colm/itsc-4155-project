-- Create books table
CREATE TABLE books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    title TEXT NOT NULL,
    published DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ISBNs table with full text search
CREATE TABLE isbns (
    isbn VARCHAR(13) PRIMARY KEY CHECK (LENGTH(isbn) IN (10, 13)),
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    isbn_type CHAR(6) NOT NULL CHECK (isbn_type IN ('isbn10', 'isbn13', 'google', 'opnlib')),
    search_vector TSVECTOR GENERATED ALWAYS AS (
        to_tsvector('english', isbn)
    ) STORED
);

-------------
-- Indexes --
-------------

CREATE INDEX i_books_title ON books USING GIN (to_tsvector('english', title));
CREATE UNIQUE INDEX i_isbns_unique_book ON isbns(book_id, isbn_type);

CREATE INDEX i_books_search ON books
USING bm25 (id, title, published)
WITH (key_field='id');

--------------
-- Triggers --
--------------

CREATE TRIGGER t_books_set_updated_at
BEFORE UPDATE ON books
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();
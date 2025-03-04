-- Create books table
CREATE TABLE books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    title TEXT NOT NULL,
    author_id UUID NOT NULL REFERENCES authors(id) ON DELETE RESTRICT,
    published DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ISBNs table with full text search
CREATE TABLE isbns (
    isbn VARCHAR(13) PRIMARY KEY CHECK (LENGTH(isbn) IN (10, 13)),
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    isbn_type VARCHAR(6) NOT NULL CHECK (isbn_type IN ('isbn10', 'isbn13')),
    search_vector TSVECTOR GENERATED ALWAYS AS (
        to_tsvector('english', isbn)
    ) STORED
);
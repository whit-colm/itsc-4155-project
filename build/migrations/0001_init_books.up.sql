-- Create books table
CREATE TABLE books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    published DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ISBNs table with full text search
CREATE TABLE isbns (
    isbn VARCHAR(17) PRIMARY KEY CHECK (LENGTH(isbn) IN (10, 13)),
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    isbn_type VARCHAR(5) NOT NULL CHECK (isbn_type IN ('10', '13')),
    search_vector TSVECTOR GENERATED ALWAYS AS (
        to_tsvector('english', isbn)
    ) STORED
);

-- Add indexes
CREATE INDEX idx_books_title ON books USING gin(to_tsvector('english', title));
CREATE INDEX idx_books_author ON books (author);
CREATE UNIQUE INDEX idx_isbns_unique_book ON isbns(book_id, isbn_type);
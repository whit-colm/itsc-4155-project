-- Add book indexes
CREATE INDEX i_books_title ON books USING GIN (to_tsvector('english', title));
CREATE INDEX i_books_author_id ON books (author_id);
CREATE UNIQUE INDEX i_isbns_unique_book ON isbns(book_id, isbn_type);
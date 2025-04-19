CREATE TABLE books_authors (
    book_id UUID REFERENCES books(id),
    author_id UUID REFERENCES authors(id),
    PRIMARY KEY (book_id, author_id)
);
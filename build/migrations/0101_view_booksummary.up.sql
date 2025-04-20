-- booksummary view gets most important details of book, minimize logic
-- in db driver.
CREATE VIEW v_books_summary AS
    SELECT 
        b.id, 
        b.title,
        b.subtitle,
        b.description,
        b.published,
        b.thumbnail_image,
        COALESCE(
            jsonb_agg(DISTINCT jsonb_build_object(
                'id', a.id,
                'family_name', a.family_name,
                'given_name', a.given_name
            )) FILTER (WHERE a.id IS NOT NULL),
            '[]'::jsonb
        ) AS authors,
        COALESCE(
            jsonb_agg(jsonb_build_object(
                'value', i.isbn,
                'type', i.isbn_type
            )) FILTER (WHERE i.isbn IS NOT NULL),
            '[]'::jsonb
        ) AS isbns
    FROM 
        books b
        LEFT JOIN books_authors ba ON b.id = ba.book_id
        LEFT JOIN authors a ON ba.author_id = a.id
        LEFT JOIN isbns i ON b.id = i.book_id
    GROUP BY
        b.id;
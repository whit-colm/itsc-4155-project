-- booksummary view gets most important details of book, minimize logic
-- in db driver.
CREATE VIEW v_booksummaries AS
    SELECT 
        b.id, 
        b.title,
        b.published,
        a.id AS author_id, 
        a.familyname AS author_lname, 
        a.givenname AS author_fname,
        COALESCE(
            json_agg(json_build_object(
                'value', i.isbn,
                'type', i.isbn_type
            )) FILTER (WHERE i.isbn IS NOT NULL),
            '[]'::json
        ) AS isbns
    FROM 
        books b
        LEFT JOIN authors AS a ON b.author_id = a.id
        LEFT JOIN isbns AS i ON b.id = i.book_id
    GROUP BY
        b.id,
        a.id;
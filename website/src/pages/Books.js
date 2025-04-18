import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import '../styles/Books.css';

function Books({ jwt }) {
  const [books, setBooks] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  // Fetch all books using search endpoint with empty query
  const fetchBooks = async (retries = 3) => {
    setLoading(true);
    setError(null);
    const resultsPerPage = 100;
    const offset = 0;

    try {
      const response = await fetch(
        `/api/search?d=booktitle&q=&r=${resultsPerPage}&o=${offset}`, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${jwt}`,
          },
      });

      if (!response.ok) {
        const errorData = await response.json();
        // If empty‐query error, treat as zero books rather than retry
        if (response.status === 400 && errorData.summary?.includes('must not be empty')) {
          setBooks([]);
          return; // exit early, no retry
        }
        throw new Error(errorData.summary || response.statusText);
      }

      const data = await response.json();
      const bookResults = Array.isArray(data)
        ? data
        : Array.isArray(data.results)
          ? data.results
          : [];
      setBooks(bookResults);

    } catch (err) {
      // Retry only for non-400 errors
      if (retries > 0 && !err.message.includes('must not be empty')) {
        console.warn(`Retrying fetchBooks... (${3 - retries + 1})`);
        setTimeout(() => fetchBooks(retries - 1), 2000);
        return;
      }
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBooks();
  }, [jwt]);

  if (loading) {
    return <div className="loading-message">Loading books...</div>;
  }

  if (error) {
    return <div className="error-message">Error: {error}</div>;
  }

  if (!loading && books.length === 0) {
    return <div className="books-container"><h1>Books</h1><p className="loading-message">No books found.</p></div>;
  }

  return (
    <div className="books-container">
      <h1>Books</h1>
      <ul className="books-list">
        {books.map((book) => (
          <li key={book.uuid} className="book-item">
            <Link to={`/books/${book.uuid}`}>
              <div>
                <h2>{book.title}</h2>
                <p><strong>Author:</strong> {book.author || 'N/A'}</p>
                <p><strong>Published:</strong> {book.published ? new Date(book.published).toLocaleDateString() : 'N/A'}</p>
              </div>
              <div>
                {book.isbn && <p><strong>ISBN:</strong> {book.isbn}</p>}
                {book.genre && <p><strong>Genre:</strong> {book.genre}</p>}
              </div>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default Books;
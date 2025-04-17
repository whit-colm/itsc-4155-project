import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import '../styles/Books.css';

function Books({ jwt }) { // Accept jwt as a prop
  const [books, setBooks] = useState([]);
  const [error, setError] = useState(null); // Add error state
  const [loading, setLoading] = useState(false); // Add loading state

  const fetchBooks = async (retries = 3) => {
    setLoading(true); // Set loading to true
    try {
      const response = await fetch('/api/books', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${jwt}`, // Include Authorization header
        },
      });
      if (!response.ok) {
        throw new Error(`Failed to fetch books: ${response.statusText}`);
      }
      const data = await response.json();
      setBooks(data);
    } catch (err) {
      if (retries > 0) {
        console.warn(`Retrying fetchBooks... (${3 - retries + 1})`);
        fetchBooks(retries - 1);
      } else {
        console.error('Error fetching books:', err);
        setError('Failed to load books. Please try again later.');
      }
    } finally {
      setLoading(false); // Set loading to false
    }
  };

  useEffect(() => {
    fetchBooks();
  }, [jwt]);

  const fetchBookByISBN = async (isbn) => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(`/api/books/isbn/${isbn}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${jwt}`,
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch book by ISBN: ${response.statusText}`);
      }

      const book = await response.json();
      setBooks([book]); // Display only the fetched book
    } catch (err) {
      console.error('Error fetching book by ISBN:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div>Loading books...</div>; // Display loading message
  }

  if (error) {
    return <div className="error-message">Error: {error}</div>; // Display error message
  }

  return (
    <div className="books-container">
      <h1>Books</h1>
      <ul className="books-list">
        {books.map((book) => (
          <li key={book.uuid} className="book-item">
            <Link to={`/books/${book.uuid}`}>
              <h2>{book.title}</h2>
              <p><strong>Author:</strong> {book.author}</p>
              <p><strong>Genre:</strong> {book.genre}</p> {/* Display genre */}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default Books;

import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import '../styles/Books.css';

function Books({ jwt }) { // Accept jwt as a prop
  const [books, setBooks] = useState([]);
  const [error, setError] = useState(null); // Add error state

  useEffect(() => {
    const fetchBooks = async () => {
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
        console.error('Error fetching books:', err);
        setError(err.message); // Set error message
      }
    };

    fetchBooks();
  }, [jwt]);

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
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default Books;

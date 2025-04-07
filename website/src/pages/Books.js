import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import '../styles/Books.css';

function Books() {
  const [books, setBooks] = useState([]);

  useEffect(() => {
    const fetchBooks = async () => {
      try {
        const response = await fetch('/api/books', {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        });
        if (!response.ok) {
          throw new Error('Failed to fetch books');
        }
        const data = await response.json();
        setBooks(data);
      } catch (error) {
        console.error('Error fetching books:', error);
      }
    };

    fetchBooks();
  }, []);

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

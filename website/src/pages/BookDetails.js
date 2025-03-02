import React, { useState, useEffect } from 'react';
import '../styles/BookDetails.css';

function BookDetails({ uuid }) {
  const [book, setBook] = useState(null);

  useEffect(() => {
    const fetchBook = async () => {
      const response = await fetch(`/api/books/${uuid}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        },
      });
      const data = await response.json();
      setBook(data);
    };

    fetchBook();
  }, [uuid]);

  if (!book) {
    return <div>Loading...</div>;
  }

  return (
    <div className="book-details-container">
      <h1>{book.title}</h1>
      <p><strong>Author:</strong> {book.author}</p>
      <p><strong>Published:</strong> {book.published}</p>
      <p><strong>ISBN-10:</strong> {book.isbns.find(isbn => isbn.type === 'isbn10').value}</p>
      <p><strong>ISBN-13:</strong> {book.isbns.find(isbn => isbn.type === 'isbn13').value}</p>
    </div>
  );
}

export default BookDetails;

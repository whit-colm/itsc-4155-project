import React, { useState, useEffect } from 'react';
import '../styles/BookDetails.css';
import Comments from './Comments';

function BookDetails({ uuid, jwt }) {
  const [book, setBook] = useState(null);
  const [coverUrl, setCoverUrl] = useState(null);
  const [error, setError] = useState(null);

  const fetchBook = async (retries = 3) => {
    try {
      const response = await fetch(`/api/books/${uuid}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${jwt}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.summary || `Failed to fetch book details: ${response.statusText}`);
      }

      const data = await response.json();
      setBook(data);

      if (data.cover_blob) { // Changed from bref_cover to cover_blob
        const coverResponse = await fetch(`/api/blob/${data.cover_blob}`, {
          headers: {
            Authorization: `Bearer ${jwt}`, // Blob endpoint requires auth
          },
        });
        
        if (coverResponse.ok) {
          const coverBlob = await coverResponse.blob();
          setCoverUrl(URL.createObjectURL(coverBlob));
        }
      }
    } catch (err) {
      if (retries > 0) {
        console.warn(`Retrying fetchBook... (${3 - retries + 1})`);
        setTimeout(() => fetchBook(retries - 1), 2000); // Add retry delay
      } else {
        setError(err.message);
      }
    }
  };

  useEffect(() => {
    fetchBook();
  }, [uuid, jwt]);

  if (error) {
    return <div className="error-message">Error: {error}</div>;
  }

  if (!book) {
    return <div className="loading-message">Loading...</div>;
  }

  return (
    <div className="book-details-container">
      <h1>{book.title}</h1>
      {coverUrl && <img src={coverUrl} alt={`${book.title} cover`} className="book-cover" />}
      {/* Wrap details in a div */}
      <div className="book-info">
        {book.author && <p><strong>Author:</strong> {book.author}</p>} {/* Added check for author */}
        <p><strong>Published:</strong> {new Date(book.published).toLocaleDateString()}</p>
        <p><strong>ISBN:</strong> {book.isbn}</p> {/* Simplified ISBN handling */}
      </div>
      <Comments bookId={uuid} jwt={jwt} />
    </div>
  );
}

export default BookDetails;
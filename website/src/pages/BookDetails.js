import React, { useState, useEffect } from 'react';
import '../styles/BookDetails.css';
import Comments from './Comments';
import { useParams, Link } from 'react-router-dom';

function BookDetails({ jwt }) {
  const { bookId } = useParams();
  const [book, setBook] = useState(null);
  const [authors, setAuthors] = useState([]);
  const [coverUrl, setCoverUrl] = useState(null);
  const [error, setError] = useState(null);
  const [loadingAuthors, setLoadingAuthors] = useState(false);
  const [authorFetchError, setAuthorFetchError] = useState(false);

  const fetchAuthorDetails = async (authorId) => {
    setAuthorFetchError(false);
    try {
      const response = await fetch(`/api/authors/${authorId}`, {
        headers: {
          ...(jwt && { Authorization: `Bearer ${jwt}` }),
        },
      });
      if (!response.ok) {
        console.error(`Failed to fetch author ${authorId}: ${response.statusText}`);
        setAuthorFetchError(true);
        return null;
      }
      const authorData = await response.json();
      return {
        id: authorData.id,
        name: `${authorData.family_name || ''}`.trim() || 'Unknown Author',
      };
    } catch (err) {
      console.error(`Error fetching author ${authorId}:`, err);
      setAuthorFetchError(true);
      return null;
    }
  };

  const fetchBook = async (retries = 3) => {
    setError(null);
    setLoadingAuthors(false);
    setAuthorFetchError(false);
    try {
      const response = await fetch(`/api/books/${bookId}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          ...(jwt && { Authorization: `Bearer ${jwt}` }),
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.summary || `Failed to fetch book details: ${response.statusText}`);
      }

      const data = await response.json();
      setBook(data);

      if (data.bref_cover_image) {
        try {
          const coverResponse = await fetch(`/api/blob/${data.bref_cover_image}`, {
            headers: {
              ...(jwt && { Authorization: `Bearer ${jwt}` }),
            },
          });
          if (coverResponse.ok) {
            const coverBlob = await coverResponse.blob();
            setCoverUrl(URL.createObjectURL(coverBlob));
          } else {
            console.warn(`Failed to fetch cover image blob: ${coverResponse.statusText}`);
          }
        } catch (imgErr) {
          console.error("Error fetching cover image:", imgErr);
        }
      }

      if (data.authors && data.authors.length > 0) {
        setLoadingAuthors(true);
        const authorPromises = data.authors.map((authorId) => fetchAuthorDetails(authorId));
        const resolvedAuthors = (await Promise.all(authorPromises)).filter((author) => author !== null);
        setAuthors(resolvedAuthors);
        setLoadingAuthors(false);
      } else {
        setAuthors([]);
      }
    } catch (err) {
      if (retries > 0) {
        console.warn(`Retrying fetchBook... (${3 - retries + 1})`);
        setTimeout(() => fetchBook(retries - 1), 2000);
      } else {
        setError(err.message);
        setBook(null);
        setAuthors([]);
        setCoverUrl(null);
      }
    }
  };

  useEffect(() => {
    if (bookId) {
      fetchBook();
    }
    return () => {
      if (coverUrl) {
        URL.revokeObjectURL(coverUrl);
      }
    };
  }, [bookId, jwt]);

  if (error) {
    return <div className="error-message">Error: {error}</div>;
  }

  if (!book) {
    return <div className="loading-message">Loading book details...</div>;
  }

  return (
    <div className="book-details-container">
      <h1>{book.title}</h1>
      {book.subtitle && <h2>{book.subtitle}</h2>}
      {coverUrl && <img src={coverUrl} alt={`${book.title} cover`} className="book-cover" />}
      <div className="book-info">
        <p>
          <strong>Author(s):</strong>{' '}
          {loadingAuthors ? (
            'Loading authors...'
          ) : authorFetchError ? (
            <span style={{ color: 'red' }}>(Could not load author details)</span>
          ) : authors.length > 0 ? (
            authors.map((author, index) => (
              <React.Fragment key={author.id}>
                <Link to={`/authors/${author.id}`}>{author.name}</Link>
                {index < authors.length - 1 ? ', ' : ''}
              </React.Fragment>
            ))
          ) : (
            'Unknown'
          )}
        </p>
        <p>
          <strong>Published:</strong>{' '}
          {book.published && typeof book.published === 'string' ?
            new Date(book.published + 'T00:00:00Z').toLocaleDateString(undefined, { timeZone: 'UTC' })
            : 'Unknown'
          }
        </p>
        {book.isbns && book.isbns.length > 0 && (
          <p>
            <strong>ISBN(s):</strong>{' '}
            {book.isbns.map((isbn, index) => (
              <span key={index}>
                {isbn.value} ({isbn.type.toUpperCase()})
                {index < book.isbns.length - 1 ? ', ' : ''}
              </span>
            ))}
          </p>
        )}
        {book.description && (
          <div className="book-description">
            <strong>Description:</strong>
            <p>{book.description}</p>
          </div>
        )}
      </div>
      <Comments bookId={bookId} jwt={jwt} />
    </div>
  );
}

export default BookDetails;
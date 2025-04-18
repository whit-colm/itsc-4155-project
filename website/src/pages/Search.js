import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import '../styles/Search.css';

function Search() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    const fetchResults = async () => {
      if (!query.trim()) {
        setResults([]);
        setTotalPages(1);
        setPage(1);
        return;
      }
      setLoading(true);
      setError(null);
      try {
        const limit = 10;
        const offset = (page - 1) * limit;
        const response = await fetch(`/api/search?q=${encodeURIComponent(query)}&limit=${limit}&offset=${offset}`);
        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.summary || `Error: ${response.statusText}`);
        }
        const data = await response.json();

        const searchResults = data.results || data || [];
        const totalItems = data.total || searchResults.length;

        setResults(Array.isArray(searchResults) ? searchResults : []);
        setTotalPages(Math.ceil(totalItems / limit));
      } catch (err) {
        setError(err.message);
        setResults([]);
        setTotalPages(1);
        setPage(1);
      } finally {
        setLoading(false);
      }
    };

    const debounceTimeout = setTimeout(() => {
      fetchResults();
    }, 300);

    return () => clearTimeout(debounceTimeout);
  }, [query, page]);

  const handleInputChange = (e) => {
    setQuery(e.target.value);
    setPage(1);
  };

  const handlePageChange = (newPage) => {
    setPage(newPage);
  };

  return (
    <div className="search-container">
      <h1>Search Books, Authors, Comments</h1>
      <input
        type="text"
        value={query}
        onChange={handleInputChange}
        placeholder="Enter your search query..."
        className="search-input"
      />
      {loading && <p className="loading-message">Loading...</p>}
      {error && <p className="error-message">{error}</p>}
      <ul className="search-results-list">
        {results.map((result) => {
          let linkTo = '/';
          let title = 'Unknown Result';
          let details = '';

          if (result.Title && result.AuthorName) {
            linkTo = `/books/${result.ID}`;
            title = result.Title;
            details = `by ${result.AuthorName}`;
          }

          return (
            <li key={result.ID || result.id || JSON.stringify(result)} className="search-result-item">
              <Link to={linkTo}>
                <h2>{title}</h2>
                <p>{details}</p>
              </Link>
            </li>
          );
        })}
      </ul>
      <div className="pagination-controls">
        <button onClick={() => handlePageChange(page - 1)} disabled={page === 1}>
          Previous
        </button>
        <button onClick={() => handlePageChange(page + 1)} disabled={page === totalPages}>
          Next
        </button>
      </div>
    </div>
  );
}

export default Search;
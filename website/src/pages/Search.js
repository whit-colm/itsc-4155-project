import React, { useState } from 'react';
import { Link } from 'react-router-dom'; // Import Link
import ReactMarkdown from 'react-markdown'; // Import ReactMarkdown
import '../styles/Search.css';

function Search({ jwt }) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const [indices, setIndices] = useState(['booktitle']);
  const [resultsPerPage, setResultsPerPage] = useState(25);
  const [offset, setOffset] = useState(0); // State for pagination offset
  const [totalResults, setTotalResults] = useState(0); // State for total results count

  const handleSearch = async (newOffset = 0, retries = 3) => {
    setLoading(true);
    setError(null);
    setOffset(newOffset); // Update offset state

    // Basic validation: ensure at least one index is selected
    if (indices.length === 0) {
        setError("Please select at least one category to search in.");
        setLoading(false);
        setResults([]);
        setTotalResults(0);
        return;
    }

    // Get JWT token
    const token = getJwt() || jwt; // Use helper function or prop

    try {
      const response = await fetch(
        // Use indices, query, resultsPerPage, and newOffset
        `/api/search?d=${indices.join(',')}&q=${encodeURIComponent(query)}&r=${resultsPerPage}&o=${newOffset}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            // Conditionally add Authorization header if token exists
            ...(token && { Authorization: `Bearer ${token}` }),
          },
        }
      );
      if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.summary || 'Search failed');
      }
      const data = await response.json();
      setResults(data.results || []);
      setTotalResults(data.total || 0); // Assuming backend returns total count
    } catch (error) {
      if (retries > 0) {
        console.warn(`Retrying search... (${3 - retries + 1})`);
        // Retry with the same offset
        setTimeout(() => handleSearch(newOffset, retries - 1), 2000);
      } else {
        setError(error.message || 'Failed to perform search. Please try again later.');
        setResults([]);
        setTotalResults(0);
      }
    } finally {
      setLoading(false);
    }
  };

  // Trigger search with offset 0 when initiating a new search
  const initiateSearch = () => {
      handleSearch(0);
  };

  const handleNextPage = () => {
      if (offset + resultsPerPage < totalResults) {
          handleSearch(offset + resultsPerPage);
      }
  };

  const handlePrevPage = () => {
      if (offset > 0) {
          handleSearch(Math.max(0, offset - resultsPerPage));
      }
  };

  // Helper function to get JWT (can be defined outside or imported)
  const getJwt = () => document.cookie
    .split('; ')
    .find((row) => row.startsWith('jwt='))
    ?.split('=')[1];

  const renderResult = (result) => {
    const type = result.apiVersion.split('.')[0];

    // Determine link target based on type
    let linkTarget = '#';
    if (type === 'book') {
        linkTarget = `/books/${result.uuid}`;
    } else if (type === 'author') {
        linkTarget = `/authors/${result.uuid}`;
    } else if (type === 'comment') {
        linkTarget = result.book_uuid ? `/books/${result.book_uuid}#comment-${result.uuid}` : '#';
    }

    return (
      <li key={result.uuid} className="search-result-item">
        <span className={`result-type ${type}`}>
          {type.charAt(0).toUpperCase() + type.slice(1)}
        </span>
        <Link to={linkTarget} className="result-link">
            <div className="result-content">
            {type === 'book' && (
                <>
                <h2>{result.title}</h2>
                <p><strong>Author:</strong> {result.author_name || result.author || 'N/A'}</p>
                {result.published && (
                    <p><strong>Published:</strong> {new Date(result.published).toLocaleDateString()}</p>
                )}
                </>
            )}
            {type === 'author' && (
                <h2>{result.given_name} {result.family_name}</h2>
            )}
            {type === 'comment' && (
                <>
                <div className="comment-body-preview"><ReactMarkdown>{result.body}</ReactMarkdown></div>
                <div className="comment-metrics">
                    {typeof result.rating === 'number' && <span>Rating: {Math.round(result.rating * 100)}%</span>}
                    <span>Votes: {result.votes || 0}</span>
                    {result.book_title && <span className="comment-book-title">On: {result.book_title}</span>}
                </div>
                </>
            )}
            </div>
        </Link>
      </li>
    );
  };

  return (
    <div className="search-container">
      <h1>Search</h1>
      <div className="search-controls">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search books, authors, comments..."
          className="search-input"
          onKeyDown={(e) => e.key === 'Enter' && initiateSearch()}
        />
        <button onClick={initiateSearch} disabled={loading || indices.length === 0}>
          {loading ? 'Searching...' : 'Search'}
        </button>
      </div>

      {error && <p className="error-message">{error}</p>}

      <div className="search-config">
        <div className="config-group index-group">
          <label>Search in:</label>
          <div className="checkbox-group">
             {['booktitle', 'authorname', 'comments', 'isbn'].map(indexValue => (
                 <label key={indexValue}>
                     <input
                         type="checkbox"
                         value={indexValue}
                         checked={indices.includes(indexValue)}
                         onChange={(e) => {
                             const { value, checked } = e.target;
                             setIndices(prev =>
                                 checked ? [...prev, value] : prev.filter(i => i !== value)
                             );
                         }}
                     />
                     {indexValue.charAt(0).toUpperCase() + indexValue.slice(1).replace('name', ' Name').replace('title', ' Title')}
                 </label>
             ))}
          </div>
        </div>

        <div className="config-group results-config">
          <label htmlFor="results-per-page">Results per page:</label>
          <input
            id="results-per-page"
            type="number"
            value={resultsPerPage}
            onChange={(e) => setResultsPerPage(Math.min(100, Math.max(1, parseInt(e.target.value, 10) || 1)))}
            min="1"
            max="100"
          />
        </div>
      </div>

      {!loading && totalResults > 0 && (
          <p className="results-info">
              Showing results {offset + 1} - {Math.min(offset + resultsPerPage, totalResults)} of {totalResults}
          </p>
      )}

      <ul className="search-results">
        {!loading && results.length === 0 && query && <p className="no-results">No results found for "{query}".</p>}
        {results.map(renderResult)}
      </ul>

      {totalResults > resultsPerPage && (
          <div className="pagination-controls">
              <button onClick={handlePrevPage} disabled={loading || offset === 0}>
                  Previous
              </button>
              <button onClick={handleNextPage} disabled={loading || offset + resultsPerPage >= totalResults}>
                  Next
              </button>
          </div>
      )}
    </div>
  );
}

export default Search;
import React, { useState } from 'react';
import '../styles/Search.css';

function Search({ jwt }) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const [indices, setIndices] = useState(['booktitle']);
  const [resultsPerPage, setResultsPerPage] = useState(25);

  const handleSearch = async (retries = 3) => {
    const offset = 0;
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(
        `/api/search?d=${indices.join(',')}&q=${encodeURIComponent(query)}&r=${resultsPerPage}&o=${offset}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${jwt}`,
          },
        }
      );
      if (!response.ok) throw new Error('Search failed');
      const data = await response.json();
      setResults(data.results || []);
    } catch (error) {
      if (retries > 0) {
        console.warn(`Retrying search... (${3 - retries + 1})`);
        handleSearch(retries - 1);
      } else {
        setError('Failed to perform search. Please try again later.');
      }
    } finally {
      setLoading(false);
    }
  };

  const renderResult = (result) => {
    if (!result.apiVersion) {
      console.error('Result missing apiVersion', result);
      return null;
    }
    
    const type = result.apiVersion.split('.')[0];
    return (
      <li key={result.uuid} className="search-result-item">
        <span className={`result-type ${type}`}>
          {type.charAt(0).toUpperCase() + type.slice(1)}
        </span>
        <div className="result-content">
          {type === 'book' && (
            <>
              <h2>{result.title}</h2>
              <p><strong>Author:</strong> {result.author}</p>
              {result.published && (
                <p><strong>Published:</strong> {new Date(result.published).toLocaleDateString()}</p>
              )}
            </>
          )}
          {type === 'author' && (
            <h2>{result.family_name}, {result.given_name}</h2>
          )}
          {type === 'comment' && (
            <>
              <p><strong>Comment:</strong> {result.body}</p>
              <div className="comment-metrics">
                <span>Rating: {Math.round(result.rating * 100)}%</span>
                <span>Votes: {result.votes}</span>
              </div>
            </>
          )}
        </div>
      </li>
    );
  };

  return (
    <div className="search-container">
      <h1>Book Search</h1>
      <div className="search-controls">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Enter book title, author, or genre"
          className="search-input"
        />
        <button onClick={handleSearch} disabled={loading}>
          {loading ? 'Searching...' : 'Search'}
        </button>
      </div>

      {error && <p className="error-message">Error: {error}</p>}

      <div className="search-config">
        <div className="config-group">
          <label>Search in:</label>
          <select
            multiple
            value={indices}
            onChange={(e) => setIndices(Array.from(e.target.selectedOptions, (option) => option.value))}
          >
            <option value="comments">Comments</option>
            <option value="booktitle">Book Titles</option>
            <option value="authorname">Author Names</option>
            <option value="isbn">ISBN</option>
          </select>
        </div>

        <div className="config-group">
          <label>Results per page:</label>
          <input
            type="number"
            value={resultsPerPage}
            onChange={(e) => setResultsPerPage(Math.min(100, Math.max(1, e.target.value)))}
            min="1"
            max="100"
          />
        </div>
      </div>

      <ul className="search-results">
        {results.map(renderResult)}
      </ul>
    </div>
  );
}

export default Search;
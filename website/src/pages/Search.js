import React, { useState } from 'react';
import '../styles/Search.css';

function Search({ jwt }) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const [indices, setIndices] = useState(['booktitle']); // Default to book titles
  const [resultsPerPage, setResultsPerPage] = useState(25);

  const handleSearch = async (retries = 3) => {
    const offset = 0; // Always start from the beginning
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
      if (!response.ok) {
        throw new Error('Search failed');
      }
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

  return (
    <div className="search-container">
      <h1>Book Search</h1>
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Enter book title, author, or genre"
        className="search-input"
      />
      <button onClick={handleSearch} disabled={loading}>
        Search
      </button>
      {loading && <p>Loading...</p>}
      {error && <p className="error-message">Error: {error}</p>}
      <div className="search-filters">
        <label>
          Search in:
          <select
            multiple
            value={indices}
            onChange={(e) =>
              setIndices(Array.from(e.target.selectedOptions, (option) => option.value))
            }
          >
            <option value="comments">Comments</option>
            <option value="booktitle">Book Titles</option>
            <option value="authorname">Author Names</option>
            <option value="isbn">ISBN</option> {/* Add ISBN as a search index */}
          </select>
        </label>
        <label>
          Results per page:
          <input
            type="number"
            value={resultsPerPage}
            onChange={(e) => setResultsPerPage(Number(e.target.value))}
            min="1"
            max="100"
          />
        </label>
      </div>
      <ul className="search-results">
        {results.map((result) => (
          <li key={result.uuid}>
            <h2>{result.title}</h2>
            <p><strong>Author:</strong> {result.author}</p>
            <p><strong>Genre:</strong> {result.genre}</p>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default Search;

import React, { useState } from 'react';
import '../styles/Search.css'; // Adjusted import path

function Search() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);

  const handleSearch = async () => {
    try {
      const response = await fetch(`/api/books?title=${encodeURIComponent(query)}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      if (!response.ok) {
        throw new Error('Search failed');
      }
      const data = await response.json();
      setResults(data);
    } catch (error) {
      console.error('Error during search:', error);
    }
  };

  return (
    <div className="search-container">
      <h1>Book Search</h1>
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Enter book title"
        className="search-input"
      />
      <button onClick={handleSearch} className="search-button">Search</button>
      <pre className="search-results">{JSON.stringify(results, null, 2)}</pre>
    </div>
  );
}

export default Search;

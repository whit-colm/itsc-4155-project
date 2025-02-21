import React, { useState } from 'react';
import '../styles/Search.css'; // Adjusted import path

function Search() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);

  const handleSearch = async () => {
    const response = await fetch('/api/books', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
      //body: JSON.stringify({ search: query })
    });
    const data = await response.json();
    setResults(data);
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

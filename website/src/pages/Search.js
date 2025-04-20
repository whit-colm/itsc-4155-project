import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import '../styles/Search.css';

// Use the exact apiVersion strings from Go model constants
const BOOK_SUMMARY_API_VERSION = "booksummary.itsc-4155-group-project.edu.whits.io/v1alpha2"; // Updated to v1alpha2
const AUTHOR_API_VERSION = "author.itsc-4155-group-project.edu.whits.io/v1alpha2";         // Updated to v1alpha2
const COMMENT_API_VERSION = "comment.itsc-4155-group-project.edu.whits.io/v1alpha1"; // Remains v1alpha1

function Search() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [page, setPage] = useState(1);
  const [hasNextPage, setHasNextPage] = useState(false);
  // State for selected search domains/indices
  const [indices, setIndices] = useState(['booktitle', 'authorname', 'comments']); // Default indices

  const limit = 10;

  // Function to handle checkbox changes for indices
  const handleIndexChange = (event) => {
    const { value, checked } = event.target;
    setIndices(prevIndices =>
      checked
        ? [...prevIndices, value] // Add index if checked
        : prevIndices.filter(index => index !== value) // Remove index if unchecked
    );
    setPage(1); // Reset to page 1 when indices change
  };

  useEffect(() => {
    const fetchResults = async () => {
      // Reset results if query or indices are empty
      if (!query.trim() || indices.length === 0) {
        setResults([]);
        setHasNextPage(false);
        if (page !== 1) setPage(1);
        // Optionally set an error if no indices are selected
        if (indices.length === 0) {
            setError("Please select at least one field to search.");
        } else {
            setError(null); // Clear error if query is just empty
        }
        return;
      }
      setLoading(true);
      setError(null);
      try {
        const offset = (page - 1) * limit;
        // Use the dynamic indices state joined by a comma
        const domainsToSearch = indices.join(',');
        const response = await fetch(
          `/api/search?q=${encodeURIComponent(query)}` +
          `&d=${domainsToSearch}&r=${limit}&o=${offset}` // Use dynamic domains
        );

        if (!response.ok) {
          // Attempt to parse error JSON, fallback to status text
          let errorSummary = `Error: ${response.statusText}`;
          try {
              const errorData = await response.json();
              errorSummary = errorData.summary || errorSummary;
          } catch (jsonError) {
              // Ignore JSON parsing error if response body is not JSON
          }
          throw new Error(errorSummary);
        }
        // Backend returns a flat array []map[string]any
        const data = await response.json();

        // Ensure data is an array
        const searchResults = Array.isArray(data) ? data : [];

        setResults(searchResults);
        // Determine if there might be a next page
        setHasNextPage(searchResults.length === limit);

      } catch (err) {
        setError(err.message);
        setResults([]);
        setHasNextPage(false); // No next page on error
        // Optionally reset page to 1 on error, or leave it
        // setPage(1);
      } finally {
        setLoading(false);
      }
    };

    // Debounce search
    const debounceTimeout = setTimeout(() => {
      fetchResults();
    }, 300);

    return () => clearTimeout(debounceTimeout);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query, page, indices]); // Add indices to dependency array

  const handleInputChange = (e) => {
    setQuery(e.target.value);
    setPage(1); // Reset to page 1 on new query
  };

  const handlePageChange = (newPage) => {
    // Basic validation
    if (newPage < 1) return;
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

      {/* Checkboxes for selecting indices */}
      <div className="search-config">
        <div className="config-group index-group">
            <label>Search In:</label>
            <div className="checkbox-group">
                <label>
                    <input
                        type="checkbox"
                        value="booktitle"
                        checked={indices.includes('booktitle')}
                        onChange={handleIndexChange}
                    /> Book Titles
                </label>
                <label>
                    <input
                        type="checkbox"
                        value="authorname"
                        checked={indices.includes('authorname')}
                        onChange={handleIndexChange}
                    /> Author Names
                </label>
                <label>
                    <input
                        type="checkbox"
                        value="comments"
                        checked={indices.includes('comments')}
                        onChange={handleIndexChange}
                    /> Comments
                </label>
            </div>
        </div>
        {/* Maybe add limit/results per page config later */}
      </div>

      {loading && <p className="loading-message">Loading...</p>}
      {error && <p className="error-message">{error}</p>}
      <ul className="search-results-list">
        {/* Filter out null/undefined results before mapping */}
        {results.filter(result => result != null).map((result, index) => { // Add index for fallback key
          // Default values
          let linkTo = '/';
          let title = 'Unknown Result Type';
          let details = JSON.stringify(result); // Fallback details
          // Use index as a fallback key if id is missing
          let key = result.id || `result-${index}`;

          // Use apiVersion to determine type and render accordingly
          switch (result.apiVersion) {
            case BOOK_SUMMARY_API_VERSION: // Now correctly matches v1alpha2
              // Ensure result.id exists before using it
              if (result.id) {
                  key = result.id;
                  linkTo = `/books/${result.id}`;
              } else {
                  console.warn("BookSummary result missing ID:", result);
                  title = "Invalid Book Data";
                  details = "Missing book identifier.";
                  break; // Prevent further processing if ID is missing
              }
              title = result.title || 'Untitled Book';
              details = result.authors?.map(a => `${a.givenname || ''} ${a.familyname || ''}`.trim()).join(', ') || 'Unknown Author';
              details = `by ${details}`;
              break;
            case AUTHOR_API_VERSION: // Now correctly matches v1alpha2
              // Ensure result.id exists before using it
              if (result.id) {
                  key = result.id;
                  linkTo = `/user/${result.id}`; // Link to user profile page
              } else {
                  console.warn("Author result missing ID:", result);
                  title = "Invalid Author Data";
                  details = "Missing author identifier.";
                  break; // Prevent further processing if ID is missing
              }
              title = `${result.givenname || ''} ${result.familyname || ''}`.trim() || 'Unknown Author';
              details = `Author`;
              break;
            case COMMENT_API_VERSION: // Still correctly matches v1alpha1
               // Ensure result.id and result.bookID exist before using them
               if (result.id && result.bookID) {
                   key = result.id;
                   linkTo = `/books/${result.bookID}#comment-${result.id}`;
               } else {
                   console.warn("Comment result missing ID or BookID:", result);
                   title = "Invalid Comment Data";
                   details = "Missing comment or book identifier.";
                   break; // Prevent further processing if IDs are missing
               }
              title = `Comment by ${result.poster?.name || result.poster?.username || 'Unknown User'}`;
              details = result.body?.substring(0, 150) + (result.body?.length > 150 ? '...' : '');
              break;
            default:
              title = 'Unrecognized Search Result';
              console.warn("Unrecognized search result type:", result);
              // Use index for key if ID is missing and type is unknown
              key = result.id || `unknown-${index}`;
              break;
          }

          return (
            <li key={key} className="search-result-item">
              <Link to={linkTo}>
                <h2>{title}</h2>
                <p>{details}</p>
              </Link>
            </li>
          );
        })}
        {!loading && results.length === 0 && query.trim() && indices.length > 0 && <p className="no-results">No results found for "{query}".</p>}
      </ul>
      {/* Update pagination controls */}
      <div>
        <div className="pagination-controls"></div>
        <button onClick={() => handlePageChange(page - 1)} disabled={page === 1 || loading}>
          Previous
        </button>
        <span>Page {page}</span>
        <button onClick={() => handlePageChange(page + 1)} disabled={!hasNextPage || loading}>
          Next
        </button>
      </div>
    </div>
  );
}

export default Search;
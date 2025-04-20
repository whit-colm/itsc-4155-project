import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import ReactMarkdown from 'react-markdown';
// import '../styles/AuthorProfile.css'; // Consider creating a CSS file for styling

function AuthorProfile({ jwt }) {
  const { authorId } = useParams();
  const [author, setAuthor] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchAuthor = async () => {
      setLoading(true);
      setError(null);
      try {
        const response = await fetch(`/api/authors/${authorId}`, {
          headers: {
            ...(jwt && { Authorization: `Bearer ${jwt}` }), // Include auth if needed by endpoint
          },
        });
        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.summary || `Failed to fetch author: ${response.statusText}`);
        }
        const data = await response.json(); // Expects model.Author structure
        setAuthor(data);
      } catch (err) {
        setError(err.message);
        setAuthor(null);
      } finally {
        setLoading(false);
      }
    };

    if (authorId) {
      fetchAuthor();
    }
  }, [authorId, jwt]);

  if (loading) {
    return <div className="loading-message">Loading author details...</div>;
  }

  if (error) {
    return <div className="error-message">Error: {error}</div>;
  }

  if (!author) {
    return <div className="error-message">Author not found.</div>;
  }

  // Construct full name from model.Author fields
  const fullName = `${author.given_name || ''} ${author.family_name || ''}`.trim();

  return (
    <div className="author-profile-container" style={{ padding: '20px', maxWidth: '800px', margin: '20px auto' }}>
      <h1>{fullName || 'Author Profile'}</h1>
      {/* Display other author details */}
      {author.bio && (
        <div className="author-bio">
          <h2>Biography</h2>
          <ReactMarkdown>{author.bio}</ReactMarkdown>
        </div>
      )}
      {author.ext_ids && author.ext_ids.length > 0 && (
        <div className="author-external-ids">
          <h2>External Identifiers</h2>
          <ul>
            {author.ext_ids.map((extId, index) => (
              <li key={index}>
                <strong>{extId.type}:</strong> {extId.id}
              </li>
            ))}
          </ul>
        </div>
      )}
      {/* TODO: Add section to list books by this author? Requires another API call */}
    </div>
  );
}

export default AuthorProfile;

import React, { useState, useEffect } from 'react';

function Health() {
  const [status, setStatus] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchHealth = async () => {
      try {
        const response = await fetch('/api/health');
        if (!response.ok) {
          throw new Error(`Health check failed: ${response.statusText}`);
        }
        const data = await response.json();
        setStatus(data);
      } catch (err) {
        setError(err.message);
      }
    };

    fetchHealth();
  }, []);

  if (error) {
    return <div>Error: {error}</div>;
  }

  if (!status) {
    return <div>Loading...</div>;
  }

  return (
    <div>
      <h1>Health Check</h1>
      <pre>{JSON.stringify(status, null, 2)}</pre>
    </div>
  );
}

export default Health;

import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

function GitHubCallback({ setJwt }) {
  const navigate = useNavigate();

  useEffect(() => {
    const fetchToken = async () => {
      try {
        const urlParams = new URLSearchParams(window.location.search);
        const code = urlParams.get('code');
        const state = urlParams.get('state');

        if (code && state) {
          const response = await fetch(`/api/auth/github/callback?code=${code}&state=${state}`);
          if (!response.ok) {
            throw new Error(`Failed to fetch token: ${response.statusText}`);
          }

          const data = await response.json();
          const token = data.token;

          if (token) {
            document.cookie = `jwt=${token}; path=/; secure; SameSite=Strict`;
            setJwt(token);
            navigate('/'); // Redirect to the homepage
          } else {
            throw new Error('Token not found in response');
          }
        }
      } catch (error) {
        console.error('Error during token exchange:', error);
        navigate('/login'); // Redirect to login on error
      }
    };

    fetchToken();
  }, [setJwt, navigate]);

  return <div>Processing GitHub login...</div>;
}

export default GitHubCallback;

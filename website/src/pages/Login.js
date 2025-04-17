import React, { useEffect } from 'react';
import '../styles/Login.css';

function Login() {
  const handleGitHubLogin = () => {
    window.location.href = '/api/auth/github/login';
  };

  useEffect(() => {
    const fetchToken = async () => {
      try {
        const urlParams = new URLSearchParams(window.location.search);
        const code = urlParams.get('code'); // GitHub sends a `code` parameter after login
        const state = urlParams.get('state'); // GitHub sends a `state` parameter for validation

        if (code && state) {
          const response = await fetch(`/api/auth/github/callback?code=${code}&state=${state}`);
          if (!response.ok) {
            throw new Error(`Failed to fetch token: ${response.statusText}`);
          }

          const data = await response.json();
          const token = data.token;

          if (token) {
            document.cookie = `jwt=${token}; path=/; secure; SameSite=Strict`; // Store token in cookie
            window.location.href = '/';
          } else {
            throw new Error('Token not found in response');
          }
        }
      } catch (error) {
        console.error('Error during token exchange:', error);
      }
    };

    fetchToken();
  }, []);

  return (
    <div className="login-container">
      <h1>Sign in with GitHub</h1>
      <button onClick={handleGitHubLogin} className="github-login-button">
        <img
          src="https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
          alt="GitHub Octocat"
          className="github-logo"
        />
        Sign in with GitHub
      </button>
    </div>
  );
}

export default Login;
import React, { useEffect } from 'react';
import '../styles/Login.css';

function Login() {
  const handleGitHubLogin = () => {
    window.location.href = '/api/auth/github/login';
  };

  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get('token');
    if (token) {
      document.cookie = `jwt=${token}; path=/; secure; HttpOnly; SameSite=Strict`; // Store token in a secure cookie
      window.location.href = '/'; // Redirect to the homepage or desired page
    }
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

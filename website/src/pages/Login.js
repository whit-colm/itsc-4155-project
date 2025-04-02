import React from 'react';
import '../styles/Login.css';

function Login() {
  const handleGitHubLogin = () => {
    window.location.href = '/api/auth/github/login';
  };

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

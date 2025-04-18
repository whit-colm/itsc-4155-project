import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Link, useParams } from 'react-router-dom';
import Home from './pages/Home';
import Search from './pages/Search';
import Login from './pages/Login';
import Profile from './pages/Profile';
import CreateBook from './pages/CreateBook';
import BookDetails from './pages/BookDetails';
import UserProfile from './pages/UserProfile';
import Books from './pages/Books';
import GitHubCallback from './pages/GitHubCallback';
import './App.css';
import Footer from './components/Footer';
import logo from './logo.png';

function App() {
  const [jwt, setJwt] = useState(null);

  const getCookie = (name) => {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
  };

  const validateToken = async () => {
    const token = getCookie('jwt');
    if (token) {
      try {
        const response = await fetch('/api/user/me', {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        if (response.ok) {
          const userData = await response.json();
          console.log('Authenticated user:', userData);
          setJwt(token);
        } else if (response.status === 401) {
          document.cookie = 'jwt=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;';
          setJwt(null);
          alert('Your session has expired. Please log in again.');
          window.location.href = '/login';
        } else {
          console.error(`Failed to validate token: ${response.statusText}`);
          alert('An error occurred while validating your session. Please try again.');
        }
      } catch (error) {
        console.error('Error validating token:', error);
        alert('A network error occurred while validating your session. Please check your connection.');
      }
    }
  };

  useEffect(() => {
    validateToken();
    const interval = setInterval(validateToken, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

  const handleLogout = () => {
    document.cookie = 'jwt=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;';
    setJwt(null);
    window.location.href = '/login';
  };

  return (
    <div className="App">
      <Router>
        <header>
          <nav className="App-nav">
            <Link to="/" className="App-logo-link">
              <img src={logo} alt="Home" className="App-logo" />
              <span className="App-logo-text">Jaws</span>
            </Link>
            <ul className="App-nav-links">
              <li>
                <Link to="/">Home</Link>
              </li>
              <li>
                <Link to="/search">Search</Link>
              </li>
              <li>
                <Link to="/books">Books</Link>
              </li>
              {!jwt && (
                <li>
                  <Link to="/login">Login</Link>
                </li>
              )}
              {jwt && (
                <li>
                  <Link to="/profile">Profile</Link>
                </li>
              )}
              {jwt && (
                <li>
                  <Link to="/create-book">Create Book</Link>
                </li>
              )}
              {jwt && (
                <li>
                  <button onClick={handleLogout}>Logout</button>
                </li>
              )}
            </ul>
          </nav>
        </header>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/search" element={<Search />} />
          <Route path="/login" element={<Login />} />
          <Route path="/profile" element={<Profile jwt={jwt} setJwt={setJwt} />} />
          <Route path="/create-book" element={<CreateBook />} />
          <Route path="/books/:bookId" element={<BookDetails jwt={jwt} />} />
          <Route path="/user/:userId" element={<UserProfile jwt={jwt} />} />
          <Route path="/books" element={<Books jwt={jwt} />} />
          <Route path="/auth/github/callback" element={<GitHubCallback setJwt={setJwt} />} />
        </Routes>
        <Footer />
      </Router>
    </div>
  );
}

export default App;
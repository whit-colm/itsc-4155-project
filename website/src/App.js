import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Home from './pages/Home';
import Search from './pages/Search';
import Login from './pages/Login';
import Profile from './pages/Profile';
import CreateBook from './pages/CreateBook';
import BookDetails from './pages/BookDetails';
import './App.css';
import Footer from './components/Footer';
import logo from './logo.png';

function App() {
  const [jwt, setJwt] = useState(null);

  useEffect(() => {
    const fetchUserData = async () => {
      const token = localStorage.getItem('jwt');
      if (token) {
        setJwt(token);
        try {
          const response = await fetch('/api/users/me', {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });
          if (response.status === 401) {
            localStorage.removeItem('jwt');
            setJwt(null);
          }
        } catch (error) {
          console.error('Error fetching user data:', error);
        }
      }
    };

    fetchUserData();
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('jwt');
    setJwt(null);
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
          <Route path="/profile" element={<Profile jwt={jwt} />} />
          <Route path="/create-book" element={<CreateBook />} />
          <Route path="/books/:uuid" element={<BookDetails />} />
        </Routes>
        <Footer />
      </Router>
    </div>
  );
}

export default App;
import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Home from './pages/Home';
import Recommendations from './pages/Recommendations';
import Search from './pages/Search';
import CreateAccount from './pages/CreateAccount';
import Login from './pages/Login';
import Profile from './pages/Profile';
import './App.css';
import Footer from './components/Footer';
import logo from './logo.png';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);

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
                <Link to="/recommendations">Recommendations</Link>
              </li>
              <li>
                <Link to="/search">Search</Link>
              </li>
              {!isLoggedIn && (
                <li>
                  <Link to="/create-account">Create Account</Link>
                </li>
              )}
              {!isLoggedIn && (
                <li>
                  <Link to="/login">Login</Link>
                </li>
              )}
              {isLoggedIn && (
                <li>
                  <Link to="/profile">Profile</Link>
                </li>
              )}
            </ul>
          </nav>
        </header>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/recommendations" element={<Recommendations />} />
          <Route path="/search" element={<Search />} />
          <Route path="/create-account" element={<CreateAccount />} />
          <Route path="/login" element={<Login setIsLoggedIn={setIsLoggedIn} />} />
          <Route path="/profile" element={<Profile />} />
        </Routes>
        <Footer />
      </Router>
    </div>
  );
}

export default App;
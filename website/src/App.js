import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Home from './pages/Home'; // Adjusted import path
import Recommendations from './pages/Recommendations'; // Adjusted import path
import './App.css'; // Corrected import path
import Footer from './components/Footer';
import logo from './logo.png'; // Import the logo

function App() {
  return (
    <div className="App">
      <Router>
        <header>
          <nav className="App-nav">
            <Link to="/">
              <img src={logo} alt="Home" className="App-logo" /> {/* Add logo */}
            </Link>
            <ul className="App-nav-links">
              <li>
                <Link to="/">Home</Link>
              </li>
              <li>
                <Link to="/recommendations">Recommendations</Link>
              </li>
            </ul>
          </nav>
        </header>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/recommendations" element={<Recommendations />} />
        </Routes>
        <Footer />
      </Router>
    </div>
  );
}

export default App;
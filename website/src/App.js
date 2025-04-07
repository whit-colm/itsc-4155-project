import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Home from './pages/Home';
import Recommendations from './pages/Recommendations';
import Search from './pages/Search';
import BookDetailPage from './pages/BookDetails';
import './App.css';
import Footer from './components/Footer';
import logo from './logo.png';

function App() {
  return (
    <div className="App">
      <Router>
        <header>
          <nav className="App-nav">
            <Link to="/">
              <img src={logo} alt="Home" className="App-logo" />
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
            </ul>
          </nav>
        </header>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/recommendations" element={<Recommendations />} />
          <Route path="/search" element={<Search />} />
          <Route path="/bookdetail/:id" element={<BookDetailPage />} />
        </Routes>
        <Footer />
      </Router>
    </div>
  );
}

export default App;

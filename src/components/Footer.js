import React from 'react';
import '../styles/Footer.css'; // Adjust the import path

function Footer() {
  return (
    <footer className="footer">
      <p>Follow us on social media:</p>
      <div className="social-icons">
        <a href="https://facebook.com" target="_blank" rel="noopener noreferrer">
          <i className="fab fa-facebook"></i>
        </a>
        <a href="https://twitter.com" target="_blank" rel="noopener noreferrer">
          <i className="fab fa-twitter"></i>
        </a>
        <a href="https://instagram.com" target="_blank" rel="noopener noreferrer">
          <i className="fab fa-instagram"></i>
        </a>
      </div>
    </footer>
  );
}

export default Footer;

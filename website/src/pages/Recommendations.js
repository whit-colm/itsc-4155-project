import React from 'react';
import '../styles/Recommendations.css'; // Adjusted import path

function Recommendations() {
  return (
    <div>
      <h1>Book Recommendations</h1>
      <p>Here are some books you might like:</p>
      <ul className="recommendations-list">
        <li>
          <img src={`${process.env.PUBLIC_URL}/1984.png`} alt="1984" className="book-image" />
          <p>1984</p>
        </li>
        <li>
          <img src={`${process.env.PUBLIC_URL}/tkam.png`} alt="To Kill a Mockingbird" className="book-image" />
          <p>To Kill a Mockingbird</p>
        </li>
        <li>
          <img src={`${process.env.PUBLIC_URL}/lotr.png`} alt="The Lord of the Rings" className="book-image" />
          <p>The Lord of the Rings</p>
        </li>
      </ul>
    </div>
  );
}

export default Recommendations;

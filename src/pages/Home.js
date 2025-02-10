import React from 'react';
import '../styles/Home.css'; // Adjusted import path

function Home() {
  return (
    <div>
      <h1>Welcome to Jaws</h1>
      <p>Your personalized book recommendation system.</p>
      <div className="image-row">
        <img src={`${process.env.PUBLIC_URL}/books.png`} alt="Description of books" className="home-image" />
        <img src={`${process.env.PUBLIC_URL}/books2.png`} alt="Description of books" className="home-image" />
      </div>
    </div>
  );
}

export default Home;

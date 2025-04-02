import React, { useState, useEffect } from 'react';
import '../styles/Profile.css';

function Profile({ jwt }) {
  const [name, setName] = useState('');
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [favoriteGenres, setFavoriteGenres] = useState([]);
  const [isEditing, setIsEditing] = useState(false);

  useEffect(() => {
    const fetchUserData = async () => {
      try {
        const response = await fetch('/api/users/me', {
          headers: {
            Authorization: `Bearer ${jwt}`,
          },
        });
        const userData = await response.json();
        setName(userData.name);
        setUsername(userData.username);
        setEmail(userData.email);
        setFavoriteGenres(userData.favoriteGenres || []);
      } catch (error) {
        console.error('Error fetching user data:', error);
      }
    };

    if (jwt) {
      fetchUserData();
    }
  }, [jwt]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    // Implement profile update logic here
    console.log('Profile updated:', { name, username, email, favoriteGenres });
    setIsEditing(false);
  };

  const handleGenreChange = (e) => {
    const { value, checked } = e.target;
    setFavoriteGenres((prev) =>
      checked ? [...prev, value] : prev.filter((genre) => genre !== value)
    );
  };

  return (
    <div className="profile-container">
      <h1>Profile Settings</h1>
      {isEditing ? (
        <form onSubmit={handleSubmit}>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Name"
            required
          />
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="Username"
            required
          />
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Email"
            required
          />
          <div>
            <label>Favorite Genres:</label>
            <div>
              <label>
                <input
                  type="checkbox"
                  value="Fiction"
                  checked={favoriteGenres.includes('Fiction')}
                  onChange={handleGenreChange}
                />
                Fiction
              </label>
              <label>
                <input
                  type="checkbox"
                  value="Mystery"
                  checked={favoriteGenres.includes('Mystery')}
                  onChange={handleGenreChange}
                />
                Mystery
              </label>
              <label>
                <input
                  type="checkbox"
                  value="Science Fiction"
                  checked={favoriteGenres.includes('Science Fiction')}
                  onChange={handleGenreChange}
                />
                Science Fiction
              </label>
              {/* Add more genres as needed */}
            </div>
          </div>
          <button type="submit">Save Changes</button>
        </form>
      ) : (
        <div>
          <p><strong>Name:</strong> {name}</p>
          <p><strong>Username:</strong> {username}</p>
          <p><strong>Email:</strong> {email}</p>
          <p><strong>Favorite Genres:</strong> {favoriteGenres.join(', ')}</p>
          <button onClick={() => setIsEditing(true)}>Edit</button>
        </div>
      )}
    </div>
  );
}

export default Profile;

import React, { useState, useEffect } from 'react';
import '../styles/Profile.css';
import base32 from 'hi-base32';
import hmacSHA1 from 'crypto-js/hmac-sha1';

function Profile({ jwt, setJwt }) { // Accept setJwt as a prop
  const [name, setName] = useState('');
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [favoriteGenres, setFavoriteGenres] = useState([]);
  const [isEditing, setIsEditing] = useState(false);
  const [userId, setUserId] = useState(''); // Add state for user ID
  const [error, setError] = useState(''); // Add state for error message

  useEffect(() => {
    const fetchUserData = async () => {
      const token = document.cookie
        .split('; ')
        .find((row) => row.startsWith('jwt='))
        ?.split('=')[1]; // Read JWT from cookie

      if (token) {
        try {
          const response = await fetch('/api/user/me', {
            headers: {
              Authorization: `Bearer ${token}`, // Send JWT as Bearer token
            },
          });

          if (!response.ok) {
            throw new Error(`Failed to fetch user data: ${response.statusText}`);
          }

          const userData = await response.json();
          setName(userData.name);
          setUsername(userData.username);
          setEmail(userData.email);
          setFavoriteGenres(userData.favoriteGenres || []);
          setUserId(userData.id); // Set user ID from the response
        } catch (error) {
          console.error('Error fetching user data:', error);
          setError('Failed to fetch user data. Please try again later.');
        }
      }
    };

    fetchUserData();
  }, [jwt]);

  const handleSaveJwt = (token) => {
    localStorage.setItem('jwt', token);
    setJwt(token); // Use setJwt from props
  };

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

  function generateDeleteTOTP(userId) {
    const base32Alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
    const secret = userId.toUpperCase().split('').filter((char) => base32Alphabet.includes(char)).join('');
    const key = base32.decode.asBytes(secret);
  
    const now = Math.floor(Date.now() / 1000);
    const count = Math.floor(now / 30);
  
    const countBytes = new Array(8).fill(0).map((_, i) => (count >> ((7 - i) * 8)) & 0xff);
    const hash = hmacSHA1(countBytes, key);
    const offset = hash.words[hash.words.length - 1] & 0xf;
  
    const code = ((hash.words[offset] & 0x7f) << 24 |
                  (hash.words[offset + 1] & 0xff) << 16 |
                  (hash.words[offset + 2] & 0xff) << 8 |
                  (hash.words[offset + 3] & 0xff)) % 1000000;
  
    return code.toString().padStart(6, '0');
  }
  
  const handleDeleteAccount = async () => {
    const code = generateDeleteTOTP(userId); // Use the fetched user ID
  
    try {
      const response = await fetch(`/api/user/me?code=${code}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${jwt}`,
        },
      });
  
      if (response.ok) {
        alert('Account deleted successfully.');
        document.cookie = 'jwt=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;';
        localStorage.removeItem('jwt'); // Clear JWT from localStorage
        window.location.href = '/';
      } else {
        alert('Failed to delete account.');
      }
    } catch (error) {
      console.error('Error deleting account:', error);
      alert('An error occurred while deleting the account.');
    }
  };

  return (
    <div className="profile-container">
      <h1>Profile Settings</h1>
      {error && <p className="error-message">{error}</p>} {/* Display error message */}
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
          <button onClick={handleDeleteAccount}>Delete Account</button>
        </div>
      )}
    </div>
  );
}

export default Profile;

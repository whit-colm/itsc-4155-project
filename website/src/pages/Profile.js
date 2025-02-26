import React, { useState, useEffect } from 'react';
import '../styles/Profile.css';

function Profile() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [favoriteGenres, setFavoriteGenres] = useState([]);
  const [readingPreferences, setReadingPreferences] = useState([]);
  const [isEditing, setIsEditing] = useState(false);

  useEffect(() => {
    // Fetch user data when the component mounts
    const fetchUserData = async () => {
      // Replace with actual API call
      const userData = {
        name: 'John Doe',
        email: 'john.doe@example.com',
        favoriteGenres: ['Fiction', 'Mystery'],
        readingPreferences: ['E-books', 'Printed books']
      };
      setName(userData.name);
      setEmail(userData.email);
      setFavoriteGenres(userData.favoriteGenres);
      setReadingPreferences(userData.readingPreferences);
    };

    fetchUserData();
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    // Implement profile update logic here
    console.log('Profile updated:', { name, email, favoriteGenres, readingPreferences });
    setIsEditing(false);
  };

  const handleGenreChange = (e) => {
    const { value, checked } = e.target;
    setFavoriteGenres((prev) =>
      checked ? [...prev, value] : prev.filter((genre) => genre !== value)
    );
  };

  const handlePreferenceChange = (e) => {
    const { value, checked } = e.target;
    setReadingPreferences((prev) =>
      checked ? [...prev, value] : prev.filter((preference) => preference !== value)
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
          <div>
            <label>Reading Preferences:</label>
            <div>
              <label>
                <input
                  type="checkbox"
                  value="E-books"
                  checked={readingPreferences.includes('E-books')}
                  onChange={handlePreferenceChange}
                />
                E-books
              </label>
              <label>
                <input
                  type="checkbox"
                  value="Printed books"
                  checked={readingPreferences.includes('Printed books')}
                  onChange={handlePreferenceChange}
                />
                Printed books
              </label>
              <label>
                <input
                  type="checkbox"
                  value="Audiobooks"
                  checked={readingPreferences.includes('Audiobooks')}
                  onChange={handlePreferenceChange}
                />
                Audiobooks
              </label>
              {/* Add more preferences as needed */}
            </div>
          </div>
          <button type="submit">Save Changes</button>
        </form>
      ) : (
        <div>
          <p><strong>Name:</strong> {name}</p>
          <p><strong>Email:</strong> {email}</p>
          <p><strong>Favorite Genres:</strong> {favoriteGenres.join(', ')}</p>
          <p><strong>Reading Preferences:</strong> {readingPreferences.join(', ')}</p>
          <button onClick={() => setIsEditing(true)}>Edit</button>
        </div>
      )}
    </div>
  );
}

export default Profile;

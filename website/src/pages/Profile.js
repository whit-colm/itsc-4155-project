import React, { useState, useEffect } from 'react';
import '../styles/Profile.css';
import base32 from 'hi-base32';
import hmacSHA1 from 'crypto-js/hmac-sha1';

function Profile({ jwt, setJwt }) {
  const [name, setName] = useState('');
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [isEditing, setIsEditing] = useState(false);
  const [userId, setUserId] = useState('');
  const [error, setError] = useState('');
  const [pronouns, setPronouns] = useState('');
  const [avatar, setAvatar] = useState('');
  const [brefAvatar, setBrefAvatar] = useState('');

  useEffect(() => {
    const fetchUserData = async () => {
      const token = document.cookie
        .split('; ')
        .find((row) => row.startsWith('jwt='))
        ?.split('=')[1];

      if (token) {
        try {
          const response = await fetch('/api/user/me', {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });

          if (!response.ok) {
            throw new Error(`Failed to fetch user data: ${response.statusText}`);
          }

          const userData = await response.json();
          setName(userData.name);
          setUsername(userData.username);
          setEmail(userData.email);
          setUserId(userData.id);
          setPronouns(userData.pronouns);
          setAvatar(`/api/avatars/${userData.avatar}`);
          setBrefAvatar(userData.bref_avatar);
        } catch (error) {
          console.error('Error fetching user data:', error);
          setError('Failed to fetch user data. Please try again later.');
        }
      }
    };

    fetchUserData();
  }, [jwt]);

  const handleSaveJwt = (token) => {
    document.cookie = `jwt=${token}; path=/; secure; SameSite=Strict`; // Save JWT in cookie
    setJwt(token);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    console.log('Profile updated:', { name, username, email, pronouns });
    setIsEditing(false);
  };

  const handleDeleteAccount = async () => {
    const code = generateDeleteTOTP(userId);

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
        window.location.href = '/';
      } else {
        alert('Failed to delete account.');
      }
    } catch (error) {
      console.error('Error deleting account:', error);
      alert('An error occurred while deleting the account.');
    }
  };

  const handleAvatarUpload = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await fetch('/api/user/me/avatar', {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${jwt}`, // Include JWT for authorization
        },
        body: formData, // Do not set Content-Type manually
      });

      if (!response.ok) {
        throw new Error('Failed to upload avatar.');
      }

      const data = await response.json();
      setBrefAvatar(data.bref_avatar); // Update avatar reference
    } catch (error) {
      console.error('Error uploading avatar:', error);
      setError('Failed to upload avatar. Please try again.');
    }
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

  return (
    <div className="profile-container">
      <h1>Profile Settings</h1>
      {error && <p className="error-message">{error}</p>}
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
          <input
            type="text"
            value={pronouns}
            onChange={(e) => setPronouns(e.target.value)}
            placeholder="Pronouns"
          />
          <button type="submit">Save Changes</button>
        </form>
      ) : (
        <div>
          <img
            alt={`${name}'s avatar`}
            className="profile-avatar"
            src={brefAvatar ? `/api/blob/${brefAvatar}` : null}
          />
          <input type="file" onChange={handleAvatarUpload} accept="image/*" />
          <p><strong>Name:</strong> {name}</p>
          <p><strong>Pronouns:</strong> {pronouns}</p>
          <p><strong>Username:</strong> {username}</p>
          <p><strong>Email:</strong> {email}</p>
          <button onClick={() => setIsEditing(true)}>Edit</button>
          <button onClick={handleDeleteAccount}>Delete Account</button>
        </div>
      )}
    </div>
  );
}

export default Profile;

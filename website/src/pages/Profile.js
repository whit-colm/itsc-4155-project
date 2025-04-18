import React, { useState, useEffect } from 'react';
import '../styles/Profile.css';

// Define defaultAvatar path assuming it's in the public folder
const defaultAvatar = '/logo192.png';

function Profile({ jwt }) {
  const [name, setName] = useState('');
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [pronouns, setPronouns] = useState('');
  const [isEditing, setIsEditing] = useState(false);
  const [userId, setUserId] = useState('');
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [brefAvatar, setBrefAvatar] = useState('');
  const [avatarUrl, setAvatarUrl] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isUploading, setIsUploading] = useState(false);

  const getJwt = () => document.cookie
    .split('; ')
    .find((row) => row.startsWith('jwt='))
    ?.split('=')[1];

  const fetchAvatar = async (blobRef) => {
    if (!blobRef) {
      setAvatarUrl(null);
      return;
    }
    const token = getJwt();
    if (!token) return;

    try {
      const response = await fetch(`/api/blob/${blobRef}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (response.ok) {
        const blob = await response.blob();
        setAvatarUrl(URL.createObjectURL(blob));
      } else {
        console.error('Failed to fetch avatar blob');
        setAvatarUrl(null);
      }
    } catch (err) {
      console.error('Error fetching avatar blob:', err);
      setAvatarUrl(null);
    }
  };

  useEffect(() => {
    const fetchUserData = async () => {
      const token = getJwt();

      if (token) {
        setError('');
        try {
          const response = await fetch('/api/user/me', {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });

          if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.summary || `Failed to fetch user data: ${response.statusText}`);
          }

          const userData = await response.json();
          setName(userData.name || '');
          setUsername(userData.username || '');
          setEmail(userData.email || '');
          setUserId(userData.uuid);
          setPronouns(userData.pronouns || '');
          setBrefAvatar(userData.bref_avatar || '');
          fetchAvatar(userData.bref_avatar);

        } catch (error) {
          console.error('Error fetching user data:', error);
          setError(error.message || 'Failed to fetch user data. Please try again later.');
          setName('');
          setUsername('');
          setEmail('');
          setUserId('');
          setPronouns('');
          setBrefAvatar('');
          setAvatarUrl(null);
        }
      } else {
        setError('Not logged in.');
      }
    };

    fetchUserData();
  }, [jwt]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError('');
    setSuccessMessage('');
    const token = getJwt();
    if (!token) {
      setError('Authentication error. Please log in.');
      setIsSubmitting(false);
      return;
    }

    const updatedData = { name, username, email, pronouns };

    try {
      const response = await fetch('/api/user/me', {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(updatedData),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.summary || 'Failed to update profile');
      }

      setSuccessMessage('Profile updated successfully!');
      setIsEditing(false);

    } catch (error) {
      console.error('Error updating profile:', error);
      setError(error.message || 'An error occurred while updating the profile.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteAccount = async () => {
    if (!window.confirm('Are you sure you want to delete your account? This action cannot be undone.')) {
      return;
    }

    const token = getJwt();
    if (!token) {
      setError('Authentication error. Please log in.');
      return;
    }

    setIsSubmitting(true);
    setError('');

    try {
      const response = await fetch(`/api/user/me`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.ok) {
        alert('Account deleted successfully.');
        document.cookie = 'jwt=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC; SameSite=Strict; secure';
        window.location.href = '/';
      } else {
        const errorData = await response.json();
        throw new Error(errorData.summary || 'Failed to delete account.');
      }
    } catch (error) {
      console.error('Error deleting account:', error);
      setError(error.message || 'An error occurred while deleting the account.');
      setIsSubmitting(false);
    }
  };

  const handleAvatarUpload = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    const token = getJwt();
    if (!token) {
      setError('Authentication error. Please log in.');
      return;
    }

    const formData = new FormData();
    formData.append('avatar', file);

    setIsUploading(true);
    setError('');
    setSuccessMessage('');

    try {
      const response = await fetch('/api/user/me/avatar', {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.summary || 'Failed to upload avatar.');
      }

      const data = await response.json();
      setBrefAvatar(data.bref_avatar);
      fetchAvatar(data.bref_avatar);
      setSuccessMessage('Avatar updated successfully!');

    } catch (error) {
      console.error('Error uploading avatar:', error);
      setError(error.message || 'Failed to upload avatar. Please try again.');
    } finally {
      setIsUploading(false);
      e.target.value = null;
    }
  };

  return (
    <div className="profile-container">
      <h1>Profile Settings</h1>
      {error && <p className="error-message">{error}</p>}
      {successMessage && <p className="success-message">{successMessage}</p>}

      <div className="profile-view">
        <div className="avatar-section">
          <img
            alt={name ? `${name}'s avatar` : 'User avatar'}
            className="profile-avatar"
            src={avatarUrl || defaultAvatar}
            onError={(e) => { e.target.onerror = null; e.target.src = defaultAvatar; }}
          />
          <label htmlFor="avatar-upload-input" className={`avatar-upload-label ${isUploading ? 'disabled' : ''}`}>
            {isUploading ? 'Uploading...' : 'Change Avatar'}
          </label>
          <input
            id="avatar-upload-input"
            type="file"
            onChange={handleAvatarUpload}
            accept="image/*"
            disabled={isUploading}
            className="avatar-upload-input"
          />
        </div>

        {isEditing ? (
          <form onSubmit={handleSubmit} className="profile-form">
            <label htmlFor="profile-name">Name:</label>
            <input
              id="profile-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Name"
              required
              disabled={isSubmitting}
            />
            <label htmlFor="profile-username">Username:</label>
            <input
              id="profile-username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Username"
              required
              disabled={isSubmitting}
            />
            <label htmlFor="profile-email">Email:</label>
            <input
              id="profile-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="Email"
              required
              disabled={isSubmitting}
            />
            <label htmlFor="profile-pronouns">Pronouns:</label>
            <input
              id="profile-pronouns"
              type="text"
              value={pronouns}
              onChange={(e) => setPronouns(e.target.value)}
              placeholder="Pronouns (e.g., she/her, they/them)"
              disabled={isSubmitting}
            />
            <div className="form-actions">
              <button type="submit" disabled={isSubmitting}>
                {isSubmitting ? 'Saving...' : 'Save Changes'}
              </button>
              <button type="button" onClick={() => setIsEditing(false)} disabled={isSubmitting} className="cancel-button">
                Cancel
              </button>
            </div>
          </form>
        ) : (
          <div className="profile-display">
            <p><strong>Name:</strong> {name || 'N/A'}</p>
            <p><strong>Pronouns:</strong> {pronouns || 'N/A'}</p>
            <p><strong>Username:</strong> {username || 'N/A'}</p>
            <p><strong>Email:</strong> {email || 'N/A'}</p>
            <div className="profile-actions">
              <button onClick={() => { setIsEditing(true); setSuccessMessage(''); setError(''); }} disabled={isSubmitting}>Edit Profile</button>
              <button onClick={handleDeleteAccount} disabled={isSubmitting} className="delete-button">Delete Account</button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default Profile;

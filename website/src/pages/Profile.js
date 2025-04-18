import React, { useState, useEffect } from 'react';
import '../styles/Profile.css';

// Define defaultAvatar path assuming it's in the public folder
const defaultAvatar = '/logo192.png';

// --- TOTP Helper Functions ---
function uuidToBase32Secret(uuid) {
  const base32chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
  return uuid.toUpperCase().split('').filter(c => base32chars.includes(c)).join('');
}

function base32Decode(str) {
  const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
  let bits = '';
  let value = 0;
  let output = [];
  for (let i = 0; i < str.length; i++) {
    value = alphabet.indexOf(str[i]);
    if (value === -1) continue;
    bits += value.toString(2).padStart(5, '0');
  }
  for (let i = 0; i + 8 <= bits.length; i += 8) {
    output.push(parseInt(bits.slice(i, i + 8), 2));
  }
  return new Uint8Array(output);
}

async function genDeleteTOTP(uuid, deltaSeconds = 0) {
  if (!uuid) throw new Error("User ID is required to generate TOTP.");
  const secret = uuidToBase32Secret(uuid);
  if (!secret || secret.length < 16) { // Basic check for a plausible secret length
      throw new Error("Invalid User ID format for TOTP generation.");
  }
  const key = base32Decode(secret);

  const now = Math.floor(Date.now() / 1000) + deltaSeconds;
  let count = Math.floor(now / 30);

  const countBytes = new Uint8Array(8);
  for (let i = 7; i >= 0; i--) {
    countBytes[i] = count & 0xff;
    count = count >> 8;
  }

  try {
    const cryptoKey = await window.crypto.subtle.importKey(
      'raw',
      key,
      { name: 'HMAC', hash: 'SHA-1' },
      false,
      ['sign']
    );
    const hashBuffer = await window.crypto.subtle.sign('HMAC', cryptoKey, countBytes);
    const hash = new Uint8Array(hashBuffer);

    const offset = hash[hash.length - 1] & 0x0f;
    const code = ((hash[offset] & 0x7f) << 24) |
                 ((hash[offset + 1] & 0xff) << 16) |
                 ((hash[offset + 2] & 0xff) << 8) |
                 (hash[offset + 3] & 0xff);

    const otp = code % 1000000;
    return otp.toString().padStart(6, '0');
  } catch (error) {
      console.error("Error generating TOTP:", error);
      throw new Error("Could not generate deletion code. Ensure your browser supports Web Crypto API.");
  }
}
// --- End TOTP Helper Functions ---

function Profile({ jwt }) {
  const [name, setName] = useState('');
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [pronouns, setPronouns] = useState('');
  const [isEditing, setIsEditing] = useState(false);
  const [userId, setUserId] = useState('');
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [avatarUuid, setAvatarUuid] = useState('');
  const [avatarUrl, setAvatarUrl] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [selectedFile, setSelectedFile] = useState(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deletionCodeInput, setDeletionCodeInput] = useState('');
  const [generatedDeletionCode, setGeneratedDeletionCode] = useState('');
  const [isGeneratingCode, setIsGeneratingCode] = useState(false);

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
          setUserId(userData.id);
          setPronouns(userData.pronouns || '');
          setAvatarUuid(userData.bref_avatar || '');
          fetchAvatar(userData.bref_avatar);

        } catch (error) {
          console.error('Error fetching user data:', error);
          setError(error.message || 'Failed to fetch user data. Please try again later.');
          setName('');
          setUsername('');
          setEmail('');
          setUserId('');
          setPronouns('');
          setAvatarUuid('');
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

    try {
      const fetchResponse = await fetch('/api/user/me', {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!fetchResponse.ok) {
        const errorData = await fetchResponse.json();
        throw new Error(errorData.summary || `Failed to fetch current user data: ${fetchResponse.statusText}`);
      }
      const currentUserData = await fetchResponse.json();

      const updatedData = {
        ...currentUserData,
        id: currentUserData.id || userId,
        name: name,
        username: username,
        email: email,
        pronouns: pronouns,
      };

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

      const updatedUser = await response.json();
      setName(updatedUser.name || '');
      setUsername(updatedUser.username || '');
      setEmail(updatedUser.email || '');
      setPronouns(updatedUser.pronouns || '');
      setAvatarUuid(updatedUser.bref_avatar || '');
      fetchAvatar(updatedUser.bref_avatar);

      setSuccessMessage('Profile updated successfully!');
      setIsEditing(false);

    } catch (error) {
      console.error('Error updating profile:', error);
      setError(error.message || 'An error occurred while updating the profile.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteAccountClick = () => {
    setError('');
    setSuccessMessage('');
    setShowDeleteConfirm(true);
    setDeletionCodeInput('');
    setGeneratedDeletionCode('');
  };

  const handleGenerateCode = async () => {
    setIsGeneratingCode(true);
    setError('');
    try {
      const code = await genDeleteTOTP(userId);
      setGeneratedDeletionCode(code);
    } catch (error) {
      console.error("Error in handleGenerateCode:", error);
      setError(error.message || "Failed to generate deletion code.");
      setGeneratedDeletionCode('');
    } finally {
      setIsGeneratingCode(false);
    }
  };

  const handleConfirmDelete = async () => {
    if (!deletionCodeInput.trim()) {
      setError('Please enter the deletion code.');
      return;
    }

    const token = getJwt();
    if (!token) {
      setError('Authentication error. Please log in.');
      setShowDeleteConfirm(false);
      return;
    }

    setIsSubmitting(true);
    setError('');

    try {
      const response = await fetch(`/api/user/me?code=${deletionCodeInput}`, {
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
        if (response.status === 403) {
          throw new Error(errorData.summary || 'Invalid deletion code or permission denied.');
        }
        throw new Error(errorData.summary || 'Failed to delete account.');
      }
    } catch (error) {
      console.error('Error deleting account:', error);
      setError(error.message || 'An error occurred while deleting the account.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleFileChange = (e) => {
    setSelectedFile(e.target.files[0]);
  };

  const handleAvatarUpload = async () => {
    if (!selectedFile) {
      setError('Please select an image file first.');
      return;
    }

    const token = getJwt();
    if (!token) {
      setError('Authentication error. Please log in.');
      return;
    }

    setIsUploading(true);
    setError('');
    setSuccessMessage('');

    try {
      const response = await fetch('/api/user/me/avatar', {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': selectedFile.type,
        },
        body: selectedFile,
      });

      if (!response.ok) {
        const contentType = response.headers.get("content-type");
        let errorData;
        if (contentType && contentType.indexOf("application/json") !== -1) {
          errorData = await response.json();
        } else {
          const errorText = await response.text();
          throw new Error(errorText || `Failed to upload avatar: ${response.statusText}`);
        }
        throw new Error(errorData.summary || 'Failed to upload avatar.');
      }

      const data = await response.json();
      setAvatarUuid(data.bref_avatar);
      fetchAvatar(data.bref_avatar);
      setSuccessMessage('Avatar updated successfully!');
      setSelectedFile(null);

    } catch (error) {
      console.error('Error uploading avatar:', error);
      setError(error.message || 'Failed to upload avatar. Please try again.');
    } finally {
      setIsUploading(false);
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
            Select Image
          </label>
          <input
            id="avatar-upload-input"
            type="file"
            onChange={handleFileChange}
            accept="image/*"
            disabled={isUploading}
            className="avatar-upload-input"
          />
          {selectedFile && <p>Selected: {selectedFile.name}</p>}
          <button onClick={handleAvatarUpload} disabled={!selectedFile || isUploading} className="avatar-upload-button">
            {isUploading ? 'Uploading...' : 'Upload Avatar'}
          </button>
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
              <button onClick={handleDeleteAccountClick} disabled={isSubmitting} className="delete-button">Delete Account</button>
            </div>
          </div>
        )}

        {showDeleteConfirm && (
          <div className="delete-confirmation">
            <h3>Confirm Account Deletion</h3>
            <p>Click the button below to generate a time-sensitive deletion code. Enter the generated code in the input field to confirm.</p>

            <button onClick={handleGenerateCode} disabled={isGeneratingCode || isSubmitting} className="generate-code-button">
              {isGeneratingCode ? 'Generating...' : 'Generate Deletion Code'}
            </button>

            {generatedDeletionCode && (
              <div className="generated-code-display">
                <p>Enter this code:</p>
                <strong>{generatedDeletionCode}</strong>
              </div>
            )}

            {error && <p className="error-message">{error}</p>}

            <label htmlFor="deletion-code">Enter Code:</label>
            <input
              id="deletion-code"
              type="text"
              value={deletionCodeInput}
              onChange={(e) => setDeletionCodeInput(e.target.value)}
              placeholder="Enter 6-digit code"
              maxLength="6"
              disabled={isSubmitting}
            />
            <div className="form-actions">
              <button onClick={handleConfirmDelete} disabled={isSubmitting || !deletionCodeInput.trim() || deletionCodeInput.length !== 6} className="delete-button">
                {isSubmitting ? 'Deleting...' : 'Confirm Delete'}
              </button>
              <button type="button" onClick={() => setShowDeleteConfirm(false)} disabled={isSubmitting} className="cancel-button">
                Cancel
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default Profile;

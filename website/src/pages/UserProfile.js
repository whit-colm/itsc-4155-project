import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import '../styles/UserProfile.css'; // Import the CSS file

// Define the fallback path directly
const defaultAvatarPath = '/logo192.png';

function UserProfile({ jwt }) {
  const { userId } = useParams(); // Corresponds to 'id' in the API path /api/user/:id
  const [user, setUser] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const [avatarUrl, setAvatarUrl] = useState(null);

  // Function to get JWT from cookie (if needed, though passed as prop)
  const getJwt = () => document.cookie
    .split('; ')
    .find((row) => row.startsWith('jwt='))
    ?.split('=')[1] || jwt; // Use prop or cookie

  // Function to fetch avatar blob and create URL
  const fetchAvatar = async (blobRef) => {
    if (!blobRef) {
      setAvatarUrl(null); // Or set a default avatar URL
      return;
    }
    const token = getJwt();
    if (!token) return; // Need token for blob endpoint

    try {
      const response = await fetch(`/api/blob/${blobRef}`, {
        headers: { Authorization: `Bearer ${token}` }, // Blob endpoint requires auth
      });
      if (response.ok) {
        const blob = await response.blob();
        setAvatarUrl(URL.createObjectURL(blob));
      } else {
        console.error('Failed to fetch avatar blob:', response.statusText);
        setAvatarUrl(null); // Or set a default/error avatar
      }
    } catch (err) {
      console.error('Error fetching avatar blob:', err);
      setAvatarUrl(null);
    }
  };

  useEffect(() => {
    const fetchUserProfile = async (retries = 3) => {
      setLoading(true);
      setError(null);
      const token = getJwt();

      try {
        // Use userId which corresponds to the :id param in the route
        const response = await fetch(`/api/user/${userId}`, {
          headers: {
            // Include token if needed by endpoint/middleware
            ...(token && { Authorization: `Bearer ${token}` }),
          },
        });
        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.summary || `Failed to fetch user profile: ${response.statusText}`);
        }
        const data = await response.json();
        setUser(data);

        // Fetch avatar using the 'bref_avatar' field
        if (data.bref_avatar) { // Use 'bref_avatar'
          fetchAvatar(data.bref_avatar); // Use 'bref_avatar'
        } else {
          setAvatarUrl(null); // No avatar reference found
        }
      } catch (err) {
        if (retries > 0) {
          console.warn(`Retrying fetchUserProfile... (${3 - retries + 1})`);
          setTimeout(() => fetchUserProfile(retries - 1), 2000); // Retry with delay
        } else {
          setError(err.message || 'Failed to load user profile. Please try again later.');
          setUser(null); // Clear user data on final error
        }
      } finally {
        // Only set loading to false after the final attempt or success
        if (retries === 0 || !error) {
          setLoading(false);
        }
      }
    };

    if (userId) {
      fetchUserProfile();
    } else {
      setError("User ID not provided.");
      setLoading(false);
    }

    // Cleanup function for the avatar URL object
    return () => {
      if (avatarUrl) {
        URL.revokeObjectURL(avatarUrl);
      }
    };
  }, [userId, jwt]); // Rerun effect if userId or jwt changes

  if (loading) {
    return <div className="loading-message">Loading user profile...</div>;
  }

  if (error) {
    return <div className="error-message">Error: {error}</div>;
  }

  if (!user) {
    return <div className="error-message">User not found.</div>;
  }

  // Display the publicly available user info using correct field names
  return (
    <div className="user-profile-container">
      {/* Use 'name' if available, fallback to 'username' */}
      <h1>{user.name || user.username || 'User Profile'}</h1>
      <img
        // Use the defined path as fallback
        src={avatarUrl || defaultAvatarPath}
        // Use 'name' for alt text
        alt={user.name ? `${user.name}'s avatar` : 'User avatar'}
        className="user-avatar"
        // Update onError fallback as well
        onError={(e) => { e.target.onerror = null; e.target.src = defaultAvatarPath; }}
      />
      {/* Display available fields */}
      {user.username && <p><strong>Username:</strong> {user.username}</p>}
      {user.pronouns && <p><strong>Pronouns:</strong> {user.pronouns}</p>}
      {/* Display admin status if desired */}
      {/* <p><strong>Admin:</strong> {user.admin ? 'Yes' : 'No'}</p> */}
    </div>
  );
}

export default UserProfile;
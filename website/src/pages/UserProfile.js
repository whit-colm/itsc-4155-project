import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';

function UserProfile({ jwt }) {
  const { userId } = useParams();
  const [user, setUser] = useState(null);
  const [error, setError] = useState(null);
  const [avatarUrl, setAvatarUrl] = useState(null);

  useEffect(() => {
    const fetchUserProfile = async () => {
      try {
        const response = await fetch(`/api/users/${userId}`, {
          headers: {
            Authorization: `Bearer ${jwt}`,
          },
        });
        if (!response.ok) {
          throw new Error(`Failed to fetch user profile: ${response.statusText}`);
        }
        const data = await response.json();
        setUser(data);

        if (data.bref_avatar) {
          const avatarResponse = await fetch(`/api/blob/${data.bref_avatar}`);
          if (!avatarResponse.ok) {
            throw new Error(`Failed to fetch avatar: ${avatarResponse.statusText}`);
          }
          const avatarBlob = await avatarResponse.blob();
          setAvatarUrl(URL.createObjectURL(avatarBlob));
        }
      } catch (err) {
        setError(err.message);
      }
    };

    fetchUserProfile();
  }, [userId, jwt]);

  if (error) {
    return <div>Error: {error}</div>;
  }

  if (!user) {
    return <div>Loading...</div>;
  }

  return (
    <div>
      <h1>{user.displayName}</h1>
      <p><strong>Username:</strong> {user.username}</p>
      <p><strong>Pronouns:</strong> {user.pronouns}</p>
      <img src={avatarUrl} alt={`${user.displayName}'s avatar`} />
    </div>
  );
}

export default UserProfile;
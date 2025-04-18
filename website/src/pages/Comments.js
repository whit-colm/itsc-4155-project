import React, { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import { v4 as uuidv4 } from 'uuid';
import '../styles/Comments.css';

// Function to get JWT from cookie
const getJwt = () => document.cookie
    .split('; ')
    .find((row) => row.startsWith('jwt='))
    ?.split('=')[1];

function Comments({ bookId, jwt: propJwt }) { // Use propJwt to avoid conflict
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState('');
  const [editingComment, setEditingComment] = useState(null); // Stores comment ID being edited
  const [editedText, setEditedText] = useState('');
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const [pendingCommentId, setPendingCommentId] = useState(null); // Track pending optimistic comment
  const [userVoteMap, setUserVoteMap] = useState({}); // Store user votes separately
  const [totalVoteMap, setTotalVoteMap] = useState({}); // Store total votes separately

  // Fetch initial comments
  useEffect(() => {
    const fetchComments = async () => {
      setLoading(true);
      setError(null);
      const token = getJwt() || propJwt; // Use token from cookie or prop

      try {
        const response = await fetch(`/api/books/${bookId}/reviews`, {
          // Conditionally add Authorization header
          headers: { ...(token && { Authorization: `Bearer ${token}` }) }
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.summary || `Failed to fetch comments: ${response.statusText}`);
        }

        let data = await response.json();
        data = Array.isArray(data) ? data : [];

        setComments(data); // Set the raw comment data

        // Initialize total votes map
        const initialTotalVotes = {};
        data.forEach(comment => {
            initialTotalVotes[comment.ID] = comment.Votes || 0; // Use ID and Votes from Go model
        });
        setTotalVoteMap(initialTotalVotes);


        // Fetch vote status for all comments if logged in
        if (token && data.length > 0) {
          fetchVoteStatuses(data.map(c => c.ID)); // Use ID from Go model
        } else {
          setUserVoteMap({}); // Clear votes if not logged in
        }

      } catch (err) {
        setError(err.message);
        setComments([]);
        setUserVoteMap({});
        setTotalVoteMap({});
      } finally {
        setLoading(false);
      }
    };

    if (bookId) {
        fetchComments();
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [bookId, propJwt]); // Rerun if bookId or jwt changes

  // Fetch vote status for multiple comments (using Voted endpoint)
  const fetchVoteStatuses = async (commentIds) => {
      const token = getJwt() || propJwt;
      if (!token || commentIds.length === 0) return;

      // Use the batch endpoint if available, otherwise call individually
      // Assuming individual calls for now based on previous logic
      const newVoteMap = { ...userVoteMap }; // Start with existing votes
      let fetchError = false;

      for (const commentId of commentIds) {
          try {
              const response = await fetch(`/api/comments/${commentId}/vote`, { // GET request to Voted endpoint
                  headers: { Authorization: `Bearer ${token}` }
              });

              if (response.status === 404) { // No vote found
                  newVoteMap[commentId] = 0;
                  continue;
              }
              if (!response.ok) throw new Error(`Failed vote status fetch for ${commentId}`);

              // Voted endpoint returns map[uuid.UUID]int8
              const voteMapResponse = await response.json();
              newVoteMap[commentId] = voteMapResponse[commentId] ?? 0; // Extract vote, default to 0

          } catch (error) {
              console.error('Error fetching vote status for comment', commentId, ':', error);
              newVoteMap[commentId] = undefined; // Mark as unknown on error
              fetchError = true;
          }
      }
      setUserVoteMap(newVoteMap); // Update state once after all fetches
      if (fetchError) {
          // Optionally set a non-blocking error message
          // setError("Could not fetch all vote statuses.");
      }
  };


  const handleAddComment = async () => {
    if (!newComment.trim()) {
      setError('Comment text cannot be empty.');
      return;
    }
    setError(null);
    const token = getJwt() || propJwt;
    if (!token) {
        setError('You must be logged in to comment.');
        return;
    }

    const tempId = uuidv4(); // Temporary ID for optimistic update
    // Optimistic comment structure matching Go model as much as possible
    const optimisticComment = {
      ID: tempId, // Use ID
      Body: newComment, // Use Body
      Poster: { DisplayName: 'You' }, // Placeholder user info - Go model uses CommentUser { ID, DisplayName, Pronouns, Username, Avatar }
      CreatedAt: new Date().toISOString(), // Use CreatedAt
      // Votes handled separately
      pending: true // Mark as pending
    };

    setComments(prev => [...prev, optimisticComment]);
    setUserVoteMap(prev => ({ ...prev, [tempId]: 0 })); // Optimistic vote state
    setTotalVoteMap(prev => ({ ...prev, [tempId]: 0 })); // Optimistic total votes
    setPendingCommentId(tempId);
    setNewComment('');

    try {
      const response = await fetch(`/api/books/${bookId}/reviews`, { // POST to create comment
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({ Body: newComment }) // Send Body matching Go model
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.summary || 'Failed to add comment');
      }

      const newCommentData = await response.json(); // Get the actual comment data from backend
      // Replace optimistic comment with real data
      setComments(prev => prev.map(c =>
          c.ID === tempId ? { ...newCommentData, pending: false } : c
      ));
      // Update vote maps with actual ID and initial votes
      setUserVoteMap(prev => {
          const { [tempId]: _, ...rest } = prev; // Remove temp ID entry
          return { ...rest, [newCommentData.ID]: 0 };
      });
      setTotalVoteMap(prev => {
          const { [tempId]: _, ...rest } = prev; // Remove temp ID entry
          return { ...rest, [newCommentData.ID]: newCommentData.Votes || 0 };
      });

    } catch (err) {
      setError(err.message);
      // Remove optimistic comment and its votes on failure
      setComments(prev => prev.filter(c => c.ID !== tempId));
      setUserVoteMap(prev => {
          const { [tempId]: _, ...rest } = prev;
          return rest;
       });
       setTotalVoteMap(prev => {
           const { [tempId]: _, ...rest } = prev;
           return rest;
       });
    } finally {
      setPendingCommentId(null); // Clear pending state
    }
  };

  const handleDeleteComment = async (commentId) => {
    const originalComments = [...comments];
    const originalUserVotes = { ...userVoteMap };
    const originalTotalVotes = { ...totalVoteMap };

    // Optimistically remove
    setComments(prev => prev.filter(c => c.ID !== commentId));
    setUserVoteMap(prev => {
        const { [commentId]: _, ...rest } = prev;
        return rest;
     });
     setTotalVoteMap(prev => {
         const { [commentId]: _, ...rest } = prev;
         return rest;
     });
    setError(null);
    const token = getJwt() || propJwt;
     if (!token) {
        setError('You must be logged in to delete comments.');
        setComments(originalComments); // Revert optimistic removal
        setUserVoteMap(originalUserVotes);
        setTotalVoteMap(originalTotalVotes);
        return;
    }

    try {
      const response = await fetch(`/api/comments/${commentId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` }
      });

      if (!response.ok) {
        // Handle specific errors like 403 Forbidden
        if (response.status === 403) {
             throw new Error("You don't have permission to delete this comment.");
        }
        const errorData = await response.json();
        throw new Error(errorData.summary || 'Failed to delete comment');
      }
      // No need to do anything on success, comment is already removed optimistically
    } catch (err) {
      setError(err.message);
      setComments(originalComments); // Revert on error
      setUserVoteMap(originalUserVotes);
      setTotalVoteMap(originalTotalVotes);
    }
  };

  const handleEditComment = (comment) => {
    setEditingComment(comment.ID); // Use ID
    setEditedText(comment.Body); // Use Body
    setError(null);
  };

  const handleSaveEdit = async (commentId) => {
    if (!editedText.trim()) {
      setError('Comment text cannot be empty.');
      return;
    }
    setError(null);
    const token = getJwt() || propJwt;
     if (!token) {
        setError('You must be logged in to edit comments.');
        return;
    }

    const originalComments = [...comments];
    // Optimistically update
    setComments(prev => prev.map(c =>
      c.ID === commentId ? { ...c, Body: editedText } : c // Use ID and Body
    ));
    setEditingComment(null); // Exit editing mode

    try {
      const response = await fetch(`/api/comments/${commentId}`, {
        method: 'PATCH', // Use PATCH for update
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        // Send only the updated field(s) matching Go model
        body: JSON.stringify({ Body: editedText })
      });

      if (!response.ok) {
         if (response.status === 403) {
             throw new Error("You don't have permission to edit this comment.");
         }
        const errorData = await response.json();
        throw new Error(errorData.summary || 'Failed to update comment');
      }
      // Optionally update comment data from response if backend returns it
      const updatedCommentData = await response.json();
      setComments(prev => prev.map(c => c.ID === commentId ? updatedCommentData : c));

    } catch (err) {
      setError(err.message);
      setComments(originalComments); // Revert on error
      setEditingComment(commentId); // Re-enter editing mode on error
    }
  };

  const handleVote = async (commentId, voteDirection) => {
      const token = getJwt() || propJwt;
      if (!token) {
          setError('You must be logged in to vote.');
          return;
      }
      setError(null);

      const originalUserVote = userVoteMap[commentId];
      const originalTotalVote = totalVoteMap[commentId];

      let optimisticVote = 0;
      let optimisticTotalChange = 0;

      const currentVote = originalUserVote ?? 0;

      // Determine new vote and change in total votes
      if (voteDirection === currentVote) { // User clicks same vote again (undo)
          optimisticVote = 0;
          optimisticTotalChange = -voteDirection;
      } else { // User clicks new vote (up or down)
          optimisticVote = voteDirection;
          optimisticTotalChange = voteDirection - currentVote;
      }
      const optimisticTotal = (originalTotalVote ?? 0) + optimisticTotalChange;

      // Optimistically update UI state maps
      setUserVoteMap(prev => ({ ...prev, [commentId]: optimisticVote }));
      setTotalVoteMap(prev => ({ ...prev, [commentId]: optimisticTotal }));


      try {
          // Call the Vote endpoint (POST)
          // Go endpoint expects vote value in query param for POST /api/comments/{id}/vote
          const response = await fetch(`/api/comments/${commentId}/vote?vote=${optimisticVote}`, {
              method: 'POST',
              headers: {
                  // 'Content-Type': 'application/json', // May not be needed if no body
                  Authorization: `Bearer ${token}`
              },
              // No body needed if vote is in query param
          });

          if (!response.ok) {
              const errorData = await response.json();
              throw new Error(errorData.summary || 'Failed to process vote');
          }

          // Vote endpoint returns map[uuid.UUID]int with updated total votes
          const voteResponseMap = await response.json();
          const finalTotalVotes = voteResponseMap[commentId];

          // Update total votes with the confirmed value from the backend
          setTotalVoteMap(prev => ({ ...prev, [commentId]: finalTotalVotes }));
          // User vote already optimistically set, assume it's correct unless error

      } catch (err) {
          setError(err.message);
          // Revert UI on error
          setUserVoteMap(prev => ({ ...prev, [commentId]: originalUserVote }));
          setTotalVoteMap(prev => ({ ...prev, [commentId]: originalTotalVote }));
      }
  };


  if (loading) return <div className="loading-message">Loading comments...</div>;

  return (
    <div className="comments-container">
      <h2>Comments</h2>
      {error && <div className="error-message">{error}</div>}

      {/* Add Comment Form - only if logged in */}
      {(getJwt() || propJwt) && ( // Check if user is logged in
          <div className="add-comment-section">
            <textarea
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
              placeholder="Add a comment (Markdown supported)"
              rows="4"
              disabled={!!pendingCommentId} // Disable while posting
            />
            <button onClick={handleAddComment} disabled={!newComment.trim() || !!pendingCommentId}>
              {pendingCommentId ? 'Posting...' : 'Post Comment'}
            </button>
          </div>
      )}
      {!(getJwt() || propJwt) && <p>Please log in to post comments.</p>}


      <ul>
        {comments.map(comment => {
          const userVote = userVoteMap[comment.ID];
          const totalVotes = totalVoteMap[comment.ID] ?? 0; // Use map or default to 0
          const isLoggedIn = !!(getJwt() || propJwt);
          // TODO: Get current user ID to compare with comment.Poster.ID for edit/delete permissions
          // const currentUserId = getCurrentUserId(); // Function to get current user ID from JWT
          // const canModify = isLoggedIn && currentUserId === comment.Poster?.ID;

          return (
              // Use ID from Go model
              <li key={comment.ID} id={`comment-${comment.ID}`} className={`comment-item ${comment.pending ? 'pending-comment' : ''}`}>
                {editingComment === comment.ID ? (
                  <div className="edit-comment-section">
                    <textarea
                      value={editedText}
                      onChange={(e) => setEditedText(e.target.value)}
                      rows="4"
                    />
                    <div className="edit-actions">
                      <button onClick={() => handleSaveEdit(comment.ID)} disabled={!editedText.trim()}>Save</button>
                      <button onClick={() => setEditingComment(null)} className="cancel-button">Cancel</button>
                    </div>
                  </div>
                ) : (
                  <div>
                    <div className="comment-meta">
                      {/* Use Poster.DisplayName from Go model */}
                      <span className="comment-author">{comment.Poster?.DisplayName || 'Anonymous'}</span>
                      {/* Use CreatedAt from Go model */}
                      <span className="comment-date">{comment.CreatedAt ? new Date(comment.CreatedAt).toLocaleString() : ''}</span>
                    </div>
                    <div className="comment-body">
                      {/* Use Body from Go model */}
                      <ReactMarkdown>{comment.Body}</ReactMarkdown>
                    </div>
                    {/* Show actions only if logged in */}
                    {isLoggedIn && (
                        <div className="comment-actions">
                          {/* Add check if user is the poster for edit/delete */}
                          {/* Example: {canModify && (...)} */}
                          <button onClick={() => handleEditComment(comment)} className="action-button">Edit</button>
                          <button onClick={() => handleDeleteComment(comment.ID)} className="action-button delete-button">Delete</button>
                          <div className="voting">
                            <button
                              onClick={() => handleVote(comment.ID, 1)}
                              className={`vote-button upvote ${userVote === 1 ? 'active' : ''}`}
                              aria-label="Upvote"
                              disabled={comment.pending || typeof userVote === 'undefined'} // Disable if pending or vote status not loaded
                            >
                              ↑
                            </button>
                            <span className="vote-count">{totalVotes}</span>
                            <button
                              onClick={() => handleVote(comment.ID, -1)}
                              className={`vote-button downvote ${userVote === -1 ? 'active' : ''}`}
                              aria-label="Downvote"
                              disabled={comment.pending || typeof userVote === 'undefined'} // Disable if pending or vote status not loaded
                            >
                              ↓
                            </button>
                          </div>
                        </div>
                    )}
                     {/* Show only voting count if not logged in */}
                     {!isLoggedIn && (
                         <div className="comment-actions">
                             <div className="voting">
                                 <span className="vote-count">{totalVotes} Votes</span>
                             </div>
                         </div>
                     )}
                  </div>
                )}
              </li>
          );
        })}
        {!loading && comments.length === 0 && <p className="no-comments">No comments yet. Be the first!</p>}
      </ul>
    </div>
  );
}

export default Comments;
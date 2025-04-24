import React, { useState, useEffect, useMemo } from 'react'; // Import useMemo
import ReactMarkdown from 'react-markdown';
import { v4 as uuidv4 } from 'uuid';
import { jwtDecode } from 'jwt-decode'; // Correct: Use named import
import '../styles/Comments.css';

// Function to get JWT from cookie
const getJwt = () => document.cookie
    .split('; ')
    .find((row) => row.startsWith('jwt='))
    ?.split('=')[1];

// Function to get user ID from JWT
const getCurrentUserId = (token) => {
    if (!token) return null;
    try {
        const decoded = jwtDecode(token);
        // Assuming the user ID is stored in the 'sub' claim (standard)
        return decoded?.sub || null;
    } catch (error) {
        console.error("Failed to decode JWT:", error);
        return null;
    }
};

function Comments({ bookId, jwt: propJwt }) { // Use propJwt to avoid conflict
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState('');
  const [editingComment, setEditingComment] = useState(null); // Stores comment ID being edited
  const [editedText, setEditedText] = useState('');
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const [pendingCommentId, setPendingCommentId] = useState(null); // Track pending optimistic comment
  const [userVoteMap, setUserVoteMap] = useState({}); // Store user votes separately: map[commentId]int8 (1, -1, 0)
  const [totalVoteMap, setTotalVoteMap] = useState({}); // Store total votes separately: map[commentId]int

  // Get the current user ID using useMemo to avoid recalculating on every render
  const currentUserId = useMemo(() => getCurrentUserId(getJwt() || propJwt), [propJwt]);

  // Fetch vote status for multiple comments using the batch endpoint
  const fetchVoteStatuses = async () => {
      const token = getJwt() || propJwt;
      if (!token || !bookId) return; // Need token and bookId

      try {
          // Use GET /api/books/:id/reviews/votes endpoint
          const response = await fetch(`/api/books/${bookId}/reviews/votes`, {
              headers: { Authorization: `Bearer ${token}` }
          });

          if (!response.ok) {
              if (response.status === 401) {
                  console.warn("Unauthorized to fetch vote statuses.");
                  setUserVoteMap({}); // Clear votes if unauthorized
                  return;
              }
              const errorData = await response.json();
              throw new Error(errorData.summary || `Failed vote status fetch for book ${bookId}`);
          }

          // Backend returns map[uuid.UUID]int8
          const voteMapResponse = await response.json();

          if (typeof voteMapResponse === 'object' && voteMapResponse !== null) {
              const processedVoteMap = {};
              for (const commentId in voteMapResponse) {
                  processedVoteMap[commentId] = Number(voteMapResponse[commentId]); // Ensure it's a number
              }
              setUserVoteMap(processedVoteMap);
          } else {
              console.error("Unexpected format for vote statuses response:", voteMapResponse);
              setUserVoteMap({}); // Reset on unexpected format
          }

      } catch (error) {
          console.error('Error fetching batch vote statuses:', error);
          setUserVoteMap({}); // Clear votes on error
      }
  };

  // Fetch initial comments and then vote statuses
  useEffect(() => {
    const fetchCommentsAndVotes = async () => {
      setLoading(true);
      setError(null);
      const token = getJwt() || propJwt;

      try {
        // Use GET /api/books/:id/reviews
        const response = await fetch(`/api/books/${bookId}/reviews`, {
          headers: { ...(token && { Authorization: `Bearer ${token}` }) }
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.summary || `Failed to fetch comments: ${response.statusText}`);
        }

        // Response is []model.Comment
        let data = await response.json();
        data = Array.isArray(data) ? data : [];

        // Process data to ensure 'date' field exists if backend sends 'CreatedAt' or similar
        const processedData = data.map(comment => ({
            ...comment,
            // Use 'date' from model.Comment, fallback if needed (though backend should send 'date')
            date: comment.date || comment.CreatedAt || new Date().toISOString()
        }));

        setComments(processedData);

        // Initialize total votes map using 'votes' field from model.Comment
        const initialTotalVotes = {};
        processedData.forEach(comment => {
            initialTotalVotes[comment.id] = comment.votes || 0; // Use lowercase 'id' and 'votes'
        });
        setTotalVoteMap(initialTotalVotes);

        // Fetch vote statuses if logged in
        if (token && processedData.length > 0) {
          await fetchVoteStatuses();
        } else {
          setUserVoteMap({});
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
        fetchCommentsAndVotes();
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [bookId, propJwt]);

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

    const tempId = uuidv4();
    // Optimistic comment structure matching model.Comment
    const optimisticComment = {
      id: tempId, // Use lowercase 'id'
      body: newComment, // Use lowercase 'body'
      poster: { name: 'You', id: currentUserId }, // Use 'poster', 'name', 'id' (matching CommentUser JSON tags)
      date: new Date().toISOString(), // Use lowercase 'date'
      votes: 0, // Use lowercase 'votes'
      pending: true
    };

    setComments(prev => [...prev, optimisticComment]);
    setUserVoteMap(prev => ({ ...prev, [tempId]: 0 })); // User hasn't voted yet
    setTotalVoteMap(prev => ({ ...prev, [tempId]: 1 })); // Poster automatically upvotes on create (based on db repo)
    setPendingCommentId(tempId);
    setNewComment('');

    try {
      // Use POST /api/books/:id/reviews
      const response = await fetch(`/api/books/${bookId}/reviews`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({ Body: newComment }) // Backend expects uppercase 'Body' in request JSON
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.summary || 'Failed to add comment');
      }

      // Response is model.Comment
      const newCommentData = await response.json();
      // Ensure the received data also uses 'date' field for consistency
      const processedNewComment = {
          ...newCommentData,
          date: newCommentData.date || newCommentData.CreatedAt || new Date().toISOString(), // Use 'date' field
          pending: false
      };

      // Replace optimistic comment with real data
      setComments(prev => prev.map(c =>
          c.id === tempId ? processedNewComment : c // Match lowercase 'id'
      ));
      // Update vote maps with actual ID and initial votes
      setUserVoteMap(prev => {
          const { [tempId]: _, ...rest } = prev;
          // User's initial vote is 1 because they created it (based on db repo)
          return { ...rest, [processedNewComment.id]: 1 };
      });
      setTotalVoteMap(prev => {
          const { [tempId]: _, ...rest } = prev;
          // Use 'votes' from the response
          return { ...rest, [processedNewComment.id]: processedNewComment.votes || 1 };
      });

    } catch (err) {
      setError(err.message);
      setComments(prev => prev.filter(c => c.id !== tempId));
      setUserVoteMap(prev => {
          const { [tempId]: _, ...rest } = prev;
          return rest;
       });
       setTotalVoteMap(prev => {
           const { [tempId]: _, ...rest } = prev;
           return rest;
       });
    } finally {
      setPendingCommentId(null);
    }
  };

  const handleDeleteComment = async (commentId) => {
    const originalComments = [...comments];
    const originalUserVotes = { ...userVoteMap };
    const originalTotalVotes = { ...totalVoteMap };

    // Optimistically remove
    setComments(prev => prev.filter(c => c.id !== commentId));
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
      // Use DELETE /api/comments/:id
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
    setEditingComment(comment.id); // Use lowercase 'id'
    setEditedText(comment.body); // Use lowercase 'body'
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
    // Optimistically update using lowercase 'id' and 'body'
    setComments(prev => prev.map(c =>
      c.id === commentId ? { ...c, body: editedText } : c
    ));
    setEditingComment(null);

    try {
      // Use PATCH /api/comments/:id
      const response = await fetch(`/api/comments/${commentId}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({ Body: editedText }) // Backend expects uppercase 'Body'
      });

      if (!response.ok) {
         if (response.status === 403) {
             throw new Error("You don't have permission to edit this comment.");
         }
        const errorData = await response.json();
        throw new Error(errorData.summary || 'Failed to update comment');
      }
      // Response is updated model.Comment
      const updatedCommentData = await response.json();
      // Ensure date consistency
      const processedUpdatedComment = {
          ...updatedCommentData,
          date: updatedCommentData.date || updatedCommentData.CreatedAt || new Date().toISOString(),
      };
      // Update state with confirmed data
      setComments(prev => prev.map(c => c.id === commentId ? processedUpdatedComment : c));

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
          // Use POST /api/comments/:id/vote?vote=<value>
          const response = await fetch(`/api/comments/${commentId}/vote?vote=${optimisticVote}`, {
              method: 'POST',
              headers: {
                  Authorization: `Bearer ${token}`
              },
          });

          if (!response.ok) {
              const errorData = await response.json();
              throw new Error(errorData.summary || 'Failed to process vote');
          }

          // Response is map[uuid.UUID]int containing the new total vote count
          const voteResponseMap = await response.json();
          const finalTotalVotes = voteResponseMap[commentId]; // Get total for the specific comment

          // Update total votes with the confirmed value from the backend
          setTotalVoteMap(prev => ({ ...prev, [commentId]: finalTotalVotes }));
          // User vote already optimistically set

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
          const userVote = userVoteMap[comment.id]; // Use lowercase 'id'
          const totalVotes = totalVoteMap[comment.id] ?? 0; // Use lowercase 'id'
          const isLoggedIn = !!(getJwt() || propJwt);
          // Check if the current user is the poster using lowercase 'id'
          const canModify = isLoggedIn && currentUserId === comment.poster?.id;

          return (
              <li key={comment.id} id={`comment-${comment.id}`} className={`comment-item ${comment.pending ? 'pending-comment' : ''}`}>
                {editingComment === comment.id ? ( // Use lowercase 'id'
                  <div className="edit-comment-section">
                    <textarea
                      value={editedText}
                      onChange={(e) => setEditedText(e.target.value)}
                      rows="4"
                    />
                    <div className="edit-actions">
                      <button onClick={() => handleSaveEdit(comment.id)} disabled={!editedText.trim()}>Save</button>
                      <button onClick={() => setEditingComment(null)} className="cancel-button">Cancel</button>
                    </div>
                  </div>
                ) : (
                  <div>
                    <div className="comment-meta">
                      <span className="comment-author">{comment.poster?.name || 'Anonymous'}</span>
                      <span className="comment-date">{comment.date ? new Date(comment.date).toLocaleString() : ''}</span>
                      {comment.edited && <span className="comment-edited">(edited {new Date(comment.edited).toLocaleString()})</span>}
                    </div>
                    <div className="comment-body">
                      <ReactMarkdown>{comment.body}</ReactMarkdown>
                    </div>
                    {isLoggedIn && (
                        <div className="comment-actions">
                          {canModify && (
                              <>
                                <button onClick={() => handleEditComment(comment)} className="action-button">Edit</button>
                                <button onClick={() => handleDeleteComment(comment.id)} className="action-button delete-button">Delete</button>
                              </>
                          )}
                          <div className="voting">
                            <button
                              onClick={() => handleVote(comment.id, 1)} // Use lowercase 'id'
                              className={`vote-button upvote ${userVote === 1 ? 'active' : ''}`}
                              aria-label="Upvote"
                              disabled={comment.pending || typeof userVote === 'undefined'}
                            >
                              ↑
                            </button>
                            <span className="vote-count">{totalVotes}</span>
                            <button
                              onClick={() => handleVote(comment.id, -1)} // Use lowercase 'id'
                              className={`vote-button downvote ${userVote === -1 ? 'active' : ''}`}
                              aria-label="Downvote"
                              disabled={comment.pending || typeof userVote === 'undefined'}
                            >
                              ↓
                            </button>
                          </div>
                        </div>
                    )}
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
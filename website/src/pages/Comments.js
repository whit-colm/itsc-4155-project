import React, { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import { v4 as uuidv4 } from 'uuid'; // Import uuidv4
import '../styles/Comments.css';

function Comments({ bookId, jwt }) {
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState('');
  const [editingComment, setEditingComment] = useState(null);
  const [editedText, setEditedText] = useState('');
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const [pendingCommentId, setPendingCommentId] = useState(null); // Track pending comment

  useEffect(() => {
    const fetchComments = async () => {
      setLoading(true);
      setError(null); // Clear previous errors
      try {
        const response = await fetch(`/api/books/${bookId}/reviews`, {
          headers: { Authorization: `Bearer ${jwt}` }
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.summary || `Failed to fetch comments: ${response.statusText}`);
        }

        let data = await response.json();
        data = Array.isArray(data) ? data : [];
        setComments(data);

        data.forEach(comment => {
          fetchVoteStatus(comment.uuid);
        });

      } catch (err) {
        setError(err.message);
        setComments([]);
      } finally {
        setLoading(false);
      }
    };
    fetchComments();
  }, [bookId, jwt]);

  const fetchVoteStatus = async (commentId) => {
    try {
      const response = await fetch(`/api/comments/${commentId}/vote`, {
        headers: { Authorization: `Bearer ${jwt}` }
      });

      if (response.status === 404) {
        setComments(prev => prev.map(comment =>
          comment.uuid === commentId ? { ...comment, userVote: 0, totalVotes: comment.votes || 0 } : comment
        ));
        return;
      }
      if (!response.ok) throw new Error('Failed to fetch vote status');

      const { vote, votes } = await response.json();
      setComments(prev => prev.map(comment =>
        comment.uuid === commentId ? { ...comment, userVote: vote, totalVotes: votes } : comment
      ));
    } catch (error) {
      console.error('Error fetching vote status for comment', commentId, ':', error);
      setComments(prev => prev.map(comment =>
        comment.uuid === commentId ? { ...comment, userVote: undefined, totalVotes: comment.votes || 0 } : comment
      ));
    }
  };

  const handleAddComment = async () => {
    if (!newComment.trim()) {
      setError('Comment text cannot be empty.');
      return;
    }
    setError(null);

    const tempId = uuidv4();
    const optimisticComment = {
      uuid: tempId,
      body: newComment,
      user: { username: 'You' },
      created_at: new Date().toISOString(),
      userVote: 0,
      votes: 0,
      pending: true
    };

    setComments(prev => [...prev, optimisticComment]);
    setPendingCommentId(tempId);
    setNewComment('');

    try {
      const response = await fetch(`/api/books/${bookId}/reviews`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${jwt}`
        },
        body: JSON.stringify({ body: newComment })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.summary || 'Failed to add comment');
      }

      const newCommentData = await response.json();
      setComments(prev => prev.map(c => c.uuid === tempId ? { ...newCommentData, pending: false } : c));
      fetchVoteStatus(newCommentData.uuid);
    } catch (err) {
      setError(err.message);
      setComments(prev => prev.filter(c => c.uuid !== tempId));
    } finally {
      setPendingCommentId(null);
    }
  };

  const handleDeleteComment = async (commentId) => {
    const originalComments = [...comments];
    setComments(prev => prev.filter(c => c.uuid !== commentId));
    setError(null);

    try {
      const response = await fetch(`/api/comments/${commentId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${jwt}` }
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.summary || 'Failed to delete comment');
      }
    } catch (err) {
      setError(err.message);
      setComments(originalComments);
    }
  };

  const handleEditComment = (comment) => {
    setEditingComment(comment.uuid);
    setEditedText(comment.body);
    setError(null);
  };

  const handleSaveEdit = async (commentId) => {
    if (!editedText.trim()) {
      setError('Comment text cannot be empty.');
      return;
    }
    setError(null);

    const originalComments = [...comments];
    setComments(prev => prev.map(c =>
      c.uuid === commentId ? { ...c, body: editedText } : c
    ));
    setEditingComment(null);

    try {
      const response = await fetch(`/api/comments/${commentId}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${jwt}`
        },
        body: JSON.stringify({ body: editedText })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.summary || 'Failed to update comment');
      }
    } catch (err) {
      setError(err.message);
      setComments(originalComments);
    }
  };

  const handleVote = async (commentId, voteDirection) => {
    const originalComments = [...comments];
    setError(null);

    setComments(prev => prev.map(c => {
      if (c.uuid === commentId) {
        const currentVote = c.userVote || 0;
        let voteChange = 0;
        if (voteDirection === currentVote) {
          voteChange = -voteDirection;
          voteDirection = 0;
        } else {
          voteChange = voteDirection - currentVote;
        }
        return {
          ...c,
          userVote: voteDirection,
          totalVotes: (c.totalVotes || c.votes || 0) + voteChange
        };
      }
      return c;
    }));

    try {
      const response = await fetch(`/api/comments/${commentId}/vote`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${jwt}`
        },
        body: JSON.stringify({ vote: voteDirection })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.summary || 'Failed to process vote');
      }

      const { votes } = await response.json();
      setComments(prev => prev.map(c =>
        c.uuid === commentId ? { ...c, totalVotes: votes } : c
      ));
    } catch (err) {
      setError(err.message);
      setComments(originalComments);
    }
  };

  if (loading) return <div className="loading-message">Loading comments...</div>;

  return (
    <div className="comments-container">
      <h2>Comments</h2>
      {error && <div className="error-message">{error}</div>}

      <div className="add-comment-section">
        <textarea
          value={newComment}
          onChange={(e) => setNewComment(e.target.value)}
          placeholder="Add a comment (Markdown supported)"
          rows="4"
        />
        <button onClick={handleAddComment} disabled={!newComment.trim() || !!pendingCommentId}>
          {pendingCommentId ? 'Posting...' : 'Post Comment'}
        </button>
      </div>

      <ul>
        {comments.map(comment => (
          <li key={comment.uuid} className={`comment-item ${comment.pending ? 'pending-comment' : ''}`}>
            {editingComment === comment.uuid ? (
              <div className="edit-comment-section">
                <textarea
                  value={editedText}
                  onChange={(e) => setEditedText(e.target.value)}
                  rows="4"
                />
                <div className="edit-actions">
                  <button onClick={() => handleSaveEdit(comment.uuid)} disabled={!editedText.trim()}>Save</button>
                  <button onClick={() => setEditingComment(null)} className="cancel-button">Cancel</button>
                </div>
              </div>
            ) : (
              <div>
                <div className="comment-meta">
                  <span className="comment-author">{comment.user?.username || 'Anonymous'}</span>
                  <span className="comment-date">{comment.created_at ? new Date(comment.created_at).toLocaleString() : ''}</span>
                </div>
                <div className="comment-body">
                  <ReactMarkdown>{comment.body}</ReactMarkdown>
                </div>
                <div className="comment-actions">
                  <button onClick={() => handleEditComment(comment)} className="action-button">Edit</button>
                  <button onClick={() => handleDeleteComment(comment.uuid)} className="action-button delete-button">Delete</button>
                  <div className="voting">
                    <button
                      onClick={() => handleVote(comment.uuid, 1)}
                      className={`vote-button upvote ${comment.userVote === 1 ? 'active' : ''}`}
                      aria-label="Upvote"
                      disabled={comment.pending}
                    >
                      ↑
                    </button>
                    <span className="vote-count">{comment.totalVotes ?? comment.votes ?? 0}</span>
                    <button
                      onClick={() => handleVote(comment.uuid, -1)}
                      className={`vote-button downvote ${comment.userVote === -1 ? 'active' : ''}`}
                      aria-label="Downvote"
                      disabled={comment.pending}
                    >
                      ↓
                    </button>
                  </div>
                </div>
              </div>
            )}
          </li>
        ))}
        {!loading && comments.length === 0 && <p className="no-comments">No comments yet. Be the first!</p>}
      </ul>
    </div>
  );
}

export default Comments;
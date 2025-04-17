import React, { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';

function Comments({ bookId, jwt }) {
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState('');
  const [editingComment, setEditingComment] = useState(null);
  const [editedText, setEditedText] = useState('');
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchComments = async () => {
      setLoading(true);
      try {
        const response = await fetch(`/api/books/${bookId}/comments`, {
          headers: {
            Authorization: `Bearer ${jwt}`,
          },
        });
        if (!response.ok) {
          throw new Error(`Failed to fetch comments: ${response.statusText}`);
        }
        const data = await response.json();
        setComments(data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    fetchComments();
  }, [bookId, jwt]);

  const fetchVoteStatus = async (commentId) => {
    try {
      const response = await fetch(`/api/comments/${commentId}/vote`, {
        headers: { Authorization: `Bearer ${jwt}` },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch vote status.');
      }

      const { userVote, totalVotes } = await response.json();
      setComments((prev) =>
        prev.map((comment) =>
          comment.id === commentId ? { ...comment, userVote, totalVotes } : comment
        )
      );
    } catch (error) {
      console.error('Error fetching vote status:', error);
    }
  };

  useEffect(() => {
    comments.forEach((comment) => fetchVoteStatus(comment.id));
  }, [comments, jwt]);

  const handleAddComment = async () => {
    if (!newComment.trim()) {
      setError('Comment text cannot be empty.');
      return;
    }
    setError(null);
    const newCommentObj = { text: newComment };
    setComments([...comments, { ...newCommentObj, id: Date.now(), userVote: 0, totalVotes: 0 }]); // Optimistic update
    setNewComment('');
    try {
      const response = await fetch(`/api/books/${bookId}/comments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${jwt}`,
        },
        body: JSON.stringify(newCommentObj),
      });
      if (!response.ok) {
        throw new Error('Failed to add comment.');
      }
      const comment = await response.json();
      setComments((prev) => prev.map((c) => (c.id === newCommentObj.id ? comment : c))); // Replace optimistic comment
    } catch (err) {
      setError(err.message);
      setComments((prev) => prev.filter((c) => c.id !== newCommentObj.id)); // Revert optimistic update
    }
  };

  const handleDeleteComment = async (commentId) => {
    const originalComments = [...comments];
    setComments(comments.filter((comment) => comment.id !== commentId)); // Optimistic update
    try {
      const response = await fetch(`/api/comments/${commentId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${jwt}` },
      });
      if (!response.ok) {
        throw new Error('Failed to delete comment.');
      }
    } catch (err) {
      setError(err.message);
      setComments(originalComments); // Revert optimistic update
    }
  };

  const handleSaveEdit = async (commentId) => {
    if (!editedText.trim()) {
      setError('Comment text cannot be empty.');
      return;
    }
    setError(null);
    const originalComments = [...comments];
    setComments(
      comments.map((comment) =>
        comment.id === commentId ? { ...comment, text: editedText } : comment
      )
    ); // Optimistic update
    setEditingComment(null);
    setEditedText('');
    try {
      const response = await fetch(`/api/comments/${commentId}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${jwt}`,
        },
        body: JSON.stringify({ text: editedText }),
      });
      if (!response.ok) {
        throw new Error('Failed to edit comment.');
      }
    } catch (err) {
      setError(err.message);
      setComments(originalComments); // Revert optimistic update
    }
  };

  const handleVote = async (commentId, vote) => {
    const originalComments = [...comments];
    setComments(
      comments.map((comment) =>
        comment.id === commentId
          ? { ...comment, userVote: vote, totalVotes: comment.totalVotes + (vote - comment.userVote) }
          : comment
      )
    ); // Optimistic update
    try {
      const response = await fetch(`/api/comments/${commentId}/vote?vote=${vote}`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${jwt}` },
      });
      if (!response.ok) {
        throw new Error('Failed to vote on comment.');
      }
      const { totalVotes } = await response.json();
      setComments(
        comments.map((comment) =>
          comment.id === commentId ? { ...comment, totalVotes } : comment
        )
      );
    } catch (err) {
      setError(err.message);
      setComments(originalComments); // Revert optimistic update
    }
  };

  if (loading) {
    return <div>Loading comments...</div>;
  }

  if (error) {
    return <div className="error-message">Error: {error}</div>;
  }

  return (
    <div>
      <h2>Comments</h2>
      <ul>
        {comments.map((comment) => (
          <li key={comment.id}>
            {editingComment === comment.id ? (
              <div>
                <textarea
                  value={editedText}
                  onChange={(e) => setEditedText(e.target.value)}
                  placeholder="Edit your comment"
                />
                <button onClick={() => handleSaveEdit(comment.id)}>Save</button>
                <button onClick={() => setEditingComment(null)}>Cancel</button>
              </div>
            ) : (
              <div>
                <ReactMarkdown>{comment.text}</ReactMarkdown> {/* Render markdown */}
                <button onClick={() => setEditingComment(comment.id)}>Edit</button>
                <button onClick={() => handleDeleteComment(comment.id)}>Delete</button>
                <div>
                  <button onClick={() => handleVote(comment.id, 1)} disabled={comment.userVote === 1}>
                    Upvote
                  </button>
                  <button onClick={() => handleVote(comment.id, -1)} disabled={comment.userVote === -1}>
                    Downvote
                  </button>
                  <button onClick={() => handleVote(comment.id, 0)} disabled={comment.userVote === 0}>
                    Remove Vote
                  </button>
                  <span>Total Votes: {comment.totalVotes}</span>
                </div>
              </div>
            )}
          </li>
        ))}
      </ul>
      <textarea
        value={newComment}
        onChange={(e) => setNewComment(e.target.value)}
        placeholder="Add a comment"
      />
      <button onClick={handleAddComment}>Submit</button>
    </div>
  );
}

export default Comments;

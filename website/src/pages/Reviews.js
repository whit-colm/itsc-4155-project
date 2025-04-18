import React, { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import { v4 as uuidv4 } from 'uuid'; // Import uuidv4 for optimistic updates
import '../styles/Reviews.css'; // Import corresponding CSS

function Reviews({ bookId, jwt }) {
  const [reviews, setReviews] = useState([]);
  const [newReview, setNewReview] = useState('');
  // Edit state
  const [editingComment, setEditingComment] = useState(null); // { uuid: string, isReply: boolean, parentUuid?: string }
  const [editedText, setEditedText] = useState('');
  // Reply state
  const [replyingTo, setReplyingTo] = useState(null); // UUID of the comment being replied to
  const [replyText, setReplyText] = useState('');
  // General state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [pendingAction, setPendingAction] = useState(null); // Track pending 'add' or 'reply'

  // Fetch initial reviews (comments)
  useEffect(() => {
    const fetchReviews = async () => {
      setLoading(true);
      setError(null);
      try {
        const response = await fetch(`/api/books/${bookId}/reviews`, {
          headers: { Authorization: `Bearer ${jwt}` },
        });
        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.summary || `Failed to fetch reviews: ${response.statusText}`);
        }
        let data = await response.json();
        data = Array.isArray(data) ? data : [];
        // Initialize vote status - fetch individually
        const reviewsWithVotes = data.map(r => ({ ...r, userVote: undefined, totalVotes: r.votes || 0 }));
        setReviews(reviewsWithVotes);
        // Fetch vote status for each review/reply
        reviewsWithVotes.forEach(review => {
            fetchVoteStatus(review.uuid);
            (review.replies || []).forEach(reply => fetchVoteStatus(reply.uuid));
        });
      } catch (err) {
        setError(err.message);
        setReviews([]);
      } finally {
        setLoading(false);
      }
    };
    fetchReviews();
  }, [bookId, jwt]);

  // Fetch vote status for a single comment/reply
   const fetchVoteStatus = async (commentUuid) => {
     try {
       const response = await fetch(`/api/comments/${commentUuid}/vote`, { // GET endpoint
         headers: { Authorization: `Bearer ${jwt}` }
       });
       if (response.status === 404) {
         updateCommentVoteState(commentUuid, 0, null); // Assume 0 if not found, keep existing total
         return;
       }
       if (!response.ok) throw new Error('Failed to fetch vote status');
       const { vote, votes } = await response.json();
       updateCommentVoteState(commentUuid, vote, votes);
     } catch (error) {
       console.error('Error fetching vote status for', commentUuid, ':', error);
       updateCommentVoteState(commentUuid, undefined, null); // Mark as unknown
     }
   };

   // Helper to update vote state immutably for reviews or replies
   const updateCommentVoteState = (commentUuid, userVote, totalVotes) => {
       setReviews(prevReviews =>
           prevReviews.map(review => {
               if (review.uuid === commentUuid) {
                   return { ...review, userVote, totalVotes: totalVotes ?? review.totalVotes };
               }
               if (review.replies) {
                   return {
                       ...review,
                       replies: review.replies.map(reply =>
                           reply.uuid === commentUuid
                               ? { ...reply, userVote, totalVotes: totalVotes ?? reply.totalVotes }
                               : reply
                       ),
                   };
               }
               return review;
           })
       );
   };


  // Add a new top-level review
  const handleAddReview = async () => {
    if (!newReview.trim()) return;
    setError(null);
    const tempId = uuidv4();
    const optimisticReview = {
      uuid: tempId, body: newReview, user: { username: 'You' }, created_at: new Date().toISOString(),
      userVote: 0, votes: 0, replies: [], pending: true
    };
    setReviews(prev => [...prev, optimisticReview]);
    setPendingAction(tempId);
    setNewReview('');

    try {
      const response = await fetch(`/api/books/${bookId}/reviews`, { // POST to book reviews
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${jwt}` },
        body: JSON.stringify({ body: newReview }), // Use 'body' field
      });
      if (!response.ok) {
        const errorData = await response.json(); throw new Error(errorData.summary || 'Failed to add review');
      }
      const actualReview = await response.json();
      setReviews(prev => prev.map(r => r.uuid === tempId ? { ...actualReview, pending: false } : r));
      fetchVoteStatus(actualReview.uuid); // Fetch vote status for new review
    } catch (err) {
      setError(err.message);
      setReviews(prev => prev.filter(r => r.uuid !== tempId)); // Revert optimistic add
    } finally {
      setPendingAction(null);
    }
  };

  // Add a reply to an existing review/comment
  const handleAddReply = async (parentUuid) => {
    if (!replyText.trim()) return;
    setError(null);
    const tempId = uuidv4();
     const optimisticReply = {
       uuid: tempId, body: replyText, parent_uuid: parentUuid, user: { username: 'You' }, created_at: new Date().toISOString(),
       userVote: 0, votes: 0, pending: true
     };

    // Optimistically add reply
    setReviews(prevReviews => prevReviews.map(review => {
        if (review.uuid === parentUuid) {
            return { ...review, replies: [...(review.replies || []), optimisticReply] };
        }
        // Add logic here if replies can be nested deeper
        return review;
    }));
    setPendingAction(tempId);
    setReplyText('');
    setReplyingTo(null);

    try {
      const response = await fetch(`/api/comments`, { // POST to comments endpoint for replies
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${jwt}` },
        body: JSON.stringify({ body: replyText, parent_uuid: parentUuid }), // Use 'body' and 'parent_uuid'
      });
      if (!response.ok) {
        const errorData = await response.json(); throw new Error(errorData.summary || 'Failed to add reply');
      }
      const actualReply = await response.json();

      // Replace optimistic reply with actual one
      setReviews(prevReviews => prevReviews.map(review => {
          if (review.uuid === parentUuid) {
              return {
                  ...review,
                  replies: (review.replies || []).map(r => r.uuid === tempId ? { ...actualReply, pending: false } : r)
              };
          }
          // Add logic here if replies can be nested deeper
          return review;
      }));
       fetchVoteStatus(actualReply.uuid); // Fetch vote status for new reply
    } catch (err) {
      setError(err.message);
      // Revert optimistic add
       setReviews(prevReviews => prevReviews.map(review => {
           if (review.uuid === parentUuid) {
               return { ...review, replies: (review.replies || []).filter(r => r.uuid !== tempId) };
           }
           return review;
       }));
    } finally {
       setPendingAction(null);
    }
  };

  // Delete a comment or reply
  const handleDeleteComment = async (commentUuid) => {
    const originalReviews = JSON.parse(JSON.stringify(reviews)); // Deep copy for revert
    setError(null);

    // Optimistic delete
    setReviews(prevReviews =>
        prevReviews.map(review => ({
            ...review,
            replies: (review.replies || []).filter(reply => reply.uuid !== commentUuid) // Filter out reply
        })).filter(review => review.uuid !== commentUuid) // Filter out top-level review
    );

    try {
      const response = await fetch(`/api/comments/${commentUuid}`, { // DELETE endpoint
        method: 'DELETE',
        headers: { Authorization: `Bearer ${jwt}` },
      });
      if (!response.ok) {
        const errorData = await response.json(); throw new Error(errorData.summary || 'Failed to delete comment');
      }
      // Delete successful, state already updated
    } catch (err) {
      setError(err.message);
      setReviews(originalReviews); // Revert on error
    }
  };

  // Handle voting
  const handleVote = async (commentUuid, voteDirection) => {
    const originalReviews = JSON.parse(JSON.stringify(reviews)); // Deep copy for revert
    setError(null);

    // Optimistic update
    setReviews(prevReviews => prevReviews.map(review => {
        const processVote = (comment) => {
            if (comment.uuid === commentUuid) {
                const currentVote = comment.userVote || 0;
                let voteChange = 0;
                let newVoteDirection = voteDirection;
                if (voteDirection === currentVote) { // Undo vote
                    voteChange = -voteDirection;
                    newVoteDirection = 0;
                } else { // New or change vote
                    voteChange = voteDirection - currentVote;
                }
                return { ...comment, userVote: newVoteDirection, totalVotes: (comment.totalVotes || comment.votes || 0) + voteChange };
            }
            return comment;
        };

        const updatedReview = processVote(review);
        if (updatedReview.replies) {
            updatedReview.replies = updatedReview.replies.map(processVote);
        }
        return updatedReview;
    }));


    try {
      const response = await fetch(`/api/comments/${commentUuid}/vote`, { // POST endpoint
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${jwt}` },
        body: JSON.stringify({ vote: voteDirection }), // Send vote in body
      });
      if (!response.ok) {
        const errorData = await response.json(); throw new Error(errorData.summary || 'Failed to process vote');
      }
      const { votes } = await response.json(); // Get actual total votes from backend
      updateCommentVoteState(commentUuid, voteDirection, votes); // Update state with confirmed total

    } catch (err) {
      setError(err.message);
      setReviews(originalReviews); // Revert on error
    }
  };

  // Start editing a comment/reply
  const handleEditComment = (comment) => {
    setEditingComment(comment.uuid);
    setEditedText(comment.body);
    setError(null);
  };

  // Save edited comment/reply
  const handleSaveEdit = async (commentUuid) => {
    if (!editedText.trim()) return;
    const originalReviews = JSON.parse(JSON.stringify(reviews));
    setError(null);

    // Optimistic update
    setReviews(prevReviews => prevReviews.map(review => {
        const updateBody = (comment) => comment.uuid === commentUuid ? { ...comment, body: editedText } : comment;
        const updatedReview = updateBody(review);
        if (updatedReview.replies) {
            updatedReview.replies = updatedReview.replies.map(updateBody);
        }
        return updatedReview;
    }));
    setEditingComment(null);

    try {
      const response = await fetch(`/api/comments/${commentUuid}`, { // PATCH endpoint
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${jwt}` },
        body: JSON.stringify({ body: editedText }),
      });
      if (!response.ok) {
        const errorData = await response.json(); throw new Error(errorData.summary || 'Failed to update comment');
      }
      // Update successful
    } catch (err) {
      setError(err.message);
      setReviews(originalReviews); // Revert on error
    }
  };

  // Render a single comment/review (and its replies recursively, though only one level is handled here)
  const renderComment = (comment, isReply = false) => (
    <li key={comment.uuid} className={`review-item ${isReply ? 'reply-item' : ''} ${comment.pending ? 'pending-comment' : ''}`}>
       {editingComment === comment.uuid ? (
         <div className="edit-comment-section">
           <textarea value={editedText} onChange={(e) => setEditedText(e.target.value)} rows="3" />
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
           <div className="comment-body"><ReactMarkdown>{comment.body}</ReactMarkdown></div>
           <div className="comment-actions">
             {/* Add logic for conditional rendering based on user permissions if available */}
             <button onClick={() => handleEditComment(comment)} className="action-button">Edit</button>
             <button onClick={() => handleDeleteComment(comment.uuid)} className="action-button delete-button">Delete</button>
             {!isReply && <button onClick={() => setReplyingTo(comment.uuid)} className="action-button">Reply</button>}
             <div className="voting">
               <button onClick={() => handleVote(comment.uuid, 1)} className={`vote-button upvote ${comment.userVote === 1 ? 'active' : ''}`} disabled={comment.pending}>↑</button>
               <span className="vote-count">{comment.totalVotes ?? comment.votes ?? 0}</span>
               <button onClick={() => handleVote(comment.uuid, -1)} className={`vote-button downvote ${comment.userVote === -1 ? 'active' : ''}`} disabled={comment.pending}>↓</button>
             </div>
           </div>
         </div>
       )}

       {/* Reply Input Area */}
       {replyingTo === comment.uuid && !isReply && (
         <div className="reply-section">
           <textarea
             value={replyText}
             onChange={(e) => setReplyText(e.target.value)}
             placeholder={`Reply to ${comment.user?.username || 'Anonymous'}...`}
             rows="2"
           />
           <div className="reply-actions">
                <button onClick={() => handleAddReply(comment.uuid)} disabled={!replyText.trim() || pendingAction}>
                    {pendingAction ? 'Replying...' : 'Post Reply'}
                </button>
                <button onClick={() => setReplyingTo(null)} className="cancel-button">Cancel</button>
           </div>
         </div>
       )}

       {/* Render Replies */}
       {!isReply && comment.replies && comment.replies.length > 0 && (
         <ul className="replies-list">
           {comment.replies.map(reply => renderComment(reply, true))}
         </ul>
       )}
    </li>
  );

  if (loading) return <div className="loading-message">Loading reviews...</div>;

  return (
    <div className="reviews-container">
      <h2>Reviews</h2>
      {error && <div className="error-message">{error}</div>}

      {/* Add Top-Level Review Form */}
      <div className="add-review-section">
        <textarea
          value={newReview}
          onChange={(e) => setNewReview(e.target.value)}
          placeholder="Write a review (Markdown supported)"
          rows="4"
        />
        <button onClick={handleAddReview} disabled={!newReview.trim() || pendingAction}>
          {pendingAction ? 'Posting...' : 'Submit Review'}
        </button>
      </div>

      {/* List of Reviews and Replies */}
      <ul className="reviews-list">
        {reviews.map(review => renderComment(review))}
        {!loading && reviews.length === 0 && <p className="no-reviews">No reviews yet. Be the first!</p>}
      </ul>
    </div>
  );
}

export default Reviews;

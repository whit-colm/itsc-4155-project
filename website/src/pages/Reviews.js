import React, { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';

function Reviews({ bookId, jwt }) {
  const [reviews, setReviews] = useState([]);
  const [newReview, setNewReview] = useState('');
  const [newReply, setNewReply] = useState({ parentId: null, text: '' });

  useEffect(() => {
    const fetchReviews = async () => {
      const response = await fetch(`/api/books/${bookId}/reviews`, {
        headers: { Authorization: `Bearer ${jwt}` },
      });
      const data = await response.json();
      setReviews(data);
    };
    fetchReviews();
  }, [bookId, jwt]);

  const handleAddReview = async () => {
    const response = await fetch(`/api/books/${bookId}/reviews`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${jwt}`,
      },
      body: JSON.stringify({ text: newReview }),
    });
    if (response.ok) {
      const review = await response.json();
      setReviews([...reviews, review]);
      setNewReview('');
    }
  };

  const handleAddReply = async (parentId) => {
    const response = await fetch(`/api/comments`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${jwt}`,
      },
      body: JSON.stringify({ text: newReply.text, parentId }),
    });
    if (response.ok) {
      const reply = await response.json();
      setReviews(
        reviews.map((review) =>
          review.id === parentId
            ? { ...review, replies: [...(review.replies || []), reply] }
            : review
        )
      );
      setNewReply({ parentId: null, text: '' });
    }
  };

  const handleDeleteComment = async (commentId) => {
    await fetch(`/comments/${commentId}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${jwt}` },
    });
    setReviews(reviews.filter((review) => review.id !== commentId));
  };

  const handleVote = async (commentId, vote) => {
    const response = await fetch(`/api/comments/${commentId}/vote?vote=${vote}`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${jwt}` },
    });

    if (response.ok) {
      const { totalVotes } = await response.json(); // Assume response includes totalVotes
      const updatedReviews = reviews.map((review) => {
        if (review.id === commentId) {
          return { ...review, userVote: vote, totalVotes };
        }
        if (review.replies) {
          return {
            ...review,
            replies: review.replies.map((reply) =>
              reply.id === commentId ? { ...reply, userVote: vote, totalVotes } : reply
            ),
          };
        }
        return review;
      });
      setReviews(updatedReviews);
    }
  };

  const fetchVotes = async () => {
    const response = await fetch(`/api/books/${bookId}/votes`, {
      headers: { Authorization: `Bearer ${jwt}` },
    });

    if (response.ok) {
      const votes = await response.json();

      const updatedReviews = reviews.map((review) => ({
        ...review,
        userVote: votes[review.id]?.userVote || 0,
        totalVotes: votes[review.id]?.totalVotes || 0,
        replies: review.replies
          ? review.replies.map((reply) => ({
              ...reply,
              userVote: votes[reply.id]?.userVote || 0,
              totalVotes: votes[reply.id]?.totalVotes || 0,
            }))
          : [],
      }));
      setReviews(updatedReviews);
    }
  };

  useEffect(() => {
    fetchVotes();
  }, [bookId, jwt]);

  return (
    <div>
      <h2>Reviews</h2>
      <ul>
        {reviews.map((review) => (
          <li key={review.id}>
            <ReactMarkdown>{review.text}</ReactMarkdown> {/* Render markdown */}
            <div>
              <button onClick={() => handleVote(review.id, 1)} disabled={review.userVote === 1}>
                Upvote
              </button>
              <button onClick={() => handleVote(review.id, -1)} disabled={review.userVote === -1}>
                Downvote
              </button>
              <button onClick={() => handleVote(review.id, 0)} disabled={review.userVote === 0}>
                Remove Vote
              </button>
              <span>Total Votes: {review.totalVotes}</span>
            </div>
            <button onClick={() => handleDeleteComment(review.id)}>Delete</button>
            <ul>
              {review.replies &&
                review.replies.map((reply) => (
                  <li key={reply.id}>
                    <ReactMarkdown>{reply.text}</ReactMarkdown> {/* Render markdown */}
                    <div>
                      <button onClick={() => handleVote(reply.id, 1)} disabled={reply.userVote === 1}>
                        Upvote
                      </button>
                      <button onClick={() => handleVote(reply.id, -1)} disabled={reply.userVote === -1}>
                        Downvote
                      </button>
                      <button onClick={() => handleVote(reply.id, 0)} disabled={reply.userVote === 0}>
                        Remove Vote
                      </button>
                      <span>Total Votes: {reply.totalVotes}</span>
                    </div>
                    <button onClick={() => handleDeleteComment(reply.id)}>Delete</button>
                  </li>
                ))}
            </ul>
            <textarea
              value={newReply.parentId === review.id ? newReply.text : ''}
              onChange={(e) =>
                setNewReply({ parentId: review.id, text: e.target.value })
              }
              placeholder="Reply to this review"
            />
            <button onClick={() => handleAddReply(review.id)}>Reply</button>
          </li>
        ))}
      </ul>
      <textarea
        value={newReview}
        onChange={(e) => setNewReview(e.target.value)}
        placeholder="Write a review"
      />
      <button onClick={handleAddReview}>Submit Review</button>
    </div>
  );
}

export default Reviews;

import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import '../styles/BookDetails.css';

const bookData = {
  '1984': {
    title: '1984',
    cover: `${process.env.PUBLIC_URL}/1984.png`,
    description:
      'A dystopian novel set in a totalitarian society ruled by Big Brother. The story explores themes of surveillance, censorship, and individualism.',
    author: 'George Orwell',
    publicationDate: 'June 8, 1949',
    averageRating: 4.5,
    reviews: [
      { user: 'Alice', comment: 'A chilling portrayal of the future.' },
      { user: 'Bob', comment: 'A masterpiece that remains relevant.' }
    ],
    similarBooks: ['Brave New World', 'Fahrenheit 451']
  },
  tkam: {
    title: 'To Kill a Mockingbird',
    cover: `${process.env.PUBLIC_URL}/tkam.png`,
    description:
      'A novel about the serious issues of rape and racial inequality, told through the eyes of a child in a small Southern town.',
    author: 'Harper Lee',
    publicationDate: 'July 11, 1960',
    averageRating: 4.8,
    reviews: [
      { user: 'Charlie', comment: 'A moving tale of morality and human compassion.' },
      { user: 'Diana', comment: 'A timeless classic that speaks to all generations.' }
    ],
    similarBooks: ['The Help', 'A Time to Kill']
  },
  lotr: {
    title: 'The Lord of the Rings',
    cover: `${process.env.PUBLIC_URL}/lotr.png`,
    description:
      'An epic fantasy adventure in Middle-earth, filled with magic, heroic quests, and epic battles that define the struggle between good and evil.',
    author: 'J.R.R. Tolkien',
    publicationDate: 'July 29, 1954',
    averageRating: 4.9,
    reviews: [
      { user: 'Eve', comment: 'An extraordinary journey into a fantastical world.' },
      { user: 'Frank', comment: 'A monumental work of literature and adventure.' }
    ],
    similarBooks: ['The Hobbit', 'The Silmarillion']
  }
};

function BookDetailPage() {
  const { id } = useParams();
  const book = id ? bookData[id] : undefined;

  // State for the review form (for current user)
  const [isReviewFormOpen, setIsReviewFormOpen] = useState(false);
  const [newRating, setNewRating] = useState('');
  const [newReview, setNewReview] = useState('');

  // State for response forms: track which review's response is open, and its text.
  const [activeResponseIndex, setActiveResponseIndex] = useState(null);
  const [newResponse, setNewResponse] = useState('');

  const handleReviewSubmit = (e) => {
    e.preventDefault();
    alert(`Review submitted!\nRating: ${newRating}\nReview: ${newReview}`);
    setNewRating('');
    setNewReview('');
    setIsReviewFormOpen(false);
  };

  const handleResponseSubmit = (reviewIndex, e) => {
    e.preventDefault();
    alert(`Response submitted for review #${reviewIndex + 1}:\n${newResponse}`);
    setNewResponse('');
    setActiveResponseIndex(null);
  };
  
  if (!book) {
    return <div>Book not found.</div>;
  }

  return (
    <div className="book-detail-container">
      <div className="book-main-details">
        <img src={book.cover} alt={book.title} className="book-cover" />
        <div className="book-info">
          <h1>{book.title}</h1>
          <p>
            <strong>Author:</strong> {book.author}
          </p>
          <p>
            <strong>Publication Date:</strong> {book.publicationDate}
          </p>
          <p>
            <strong>Average Rating:</strong> {book.averageRating}
          </p>
          <p>{book.description}</p>
          <button
            onClick={() => setIsReviewFormOpen(!isReviewFormOpen)}
            className="review-btn"
          >
            {isReviewFormOpen ? 'Cancel' : 'Add Review / Rate Book'}
          </button>
          {isReviewFormOpen && (
            <form onSubmit={handleReviewSubmit} className="review-form">
              <label>
                Rating (1-5):
                <input
                  type="number"
                  min="1"
                  max="5"
                  value={newRating}
                  onChange={(e) => setNewRating(e.target.value)}
                  required
                />
              </label>
              <label>
                Review:
                <textarea
                  value={newReview}
                  onChange={(e) => setNewReview(e.target.value)}
                  required
                ></textarea>
              </label>
              <button type="submit">Submit Review</button>
            </form>
          )}
        </div>
      </div>
      <div className="additional-details">
        <h2>User Reviews</h2>
        <ul className="reviews-list">
          {book.reviews.map((review, index) => (
            <li key={index}>
              <div className="review-content">
                <strong>{review.user}:</strong> {review.comment}
              </div>
              <button
                onClick={() =>
                  setActiveResponseIndex(
                    activeResponseIndex === index ? null : index
                  )
                }
                className="respond-btn"
              >
                {activeResponseIndex === index ? 'Cancel Response' : 'Respond'}
              </button>
              {activeResponseIndex === index && (
                <form
                  onSubmit={(e) => handleResponseSubmit(index, e)}
                  className="response-form"
                >
                  <textarea
                    value={newResponse}
                    onChange={(e) => setNewResponse(e.target.value)}
                    placeholder="Write your response..."
                    required
                  ></textarea>
                  <button type="submit">Submit Response</button>
                </form>
              )}
            </li>
          ))}
        </ul>
        <h2>Similar Books</h2>
        <ul className="similar-books-list">
          {book.similarBooks.map((similar, index) => (
            <li key={index}>{similar}</li>
          ))}
        </ul>
      </div>
    </div>
  );
}

export default BookDetailPage;

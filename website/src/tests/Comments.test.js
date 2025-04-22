import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import Comments from '../pages/Comments';
import { jwtDecode } from 'jwt-decode'; // Import jwtDecode

const mockJwt = 'mock-jwt-token';
const mockBookId = 'mock-book-id';
const mockUserId = 'mock-user-id'; // Assume this is the logged-in user

// Mock jwtDecode
jest.mock('jwt-decode', () => ({
  jwtDecode: jest.fn(() => ({ sub: mockUserId })), // Mock to return user ID
}));

// Mock document.cookie
Object.defineProperty(document, 'cookie', {
    writable: true,
    value: `jwt=${mockJwt}`,
});

// --- Mock Data matching Go models ---
const mockCommentsData = [
    {
        id: 'comment-1',
        body: 'This is the first comment.',
        bookID: mockBookId,
        date: new Date(Date.now() - 100000).toISOString(),
        poster: { id: 'user-other', name: 'Other User', username: 'other#1111', avatar: 'avatar-other-id' }, // Match model.CommentUser
        votes: 5, // Match model.Comment
        edited: null, // Match model.Comment
        deleted: false, // Match model.Comment
        rating: 0, // Match model.Comment
        parent: '00000000-0000-0000-0000-000000000000', // Match model.Comment (use nil UUID)
    },
    {
        id: 'comment-2',
        body: 'This is the second comment, by me.',
        bookID: mockBookId,
        date: new Date().toISOString(),
        poster: { id: mockUserId, name: 'Me User', username: 'me#2222', avatar: 'avatar-me-id' }, // Match model.CommentUser
        votes: 3, // Match model.Comment
        edited: null, // Match model.Comment
        deleted: false, // Match model.Comment
        rating: 0, // Match model.Comment
        parent: '00000000-0000-0000-0000-000000000000', // Match model.Comment (use nil UUID)
    },
];

// Mock vote status map: map[commentId]int8
const mockVotesData = {
    'comment-1': 0, // Current user hasn't voted on comment 1
    'comment-2': 1, // Current user has upvoted comment 2
};
// --- End Mock Data ---


beforeEach(() => {
  // Reset mocks
  jwtDecode.mockClear();
  global.fetch = jest.fn((url, options) => {
    console.log('Fetch called:', url, options?.method);

    // GET /api/books/:id/reviews
    if (url === `/api/books/${mockBookId}/reviews` && (!options || options.method === 'GET')) {
      console.log('Mocking GET /api/books/:id/reviews');
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve(mockCommentsData),
      });
    }
    // GET /api/books/:id/reviews/votes
    else if (url === `/api/books/${mockBookId}/reviews/votes`) {
        console.log('Mocking GET /api/books/:id/reviews/votes');
        return Promise.resolve({
            ok: true,
            json: () => Promise.resolve(mockVotesData),
        });
    }
    // POST /api/books/:id/reviews (Add comment)
    else if (url === `/api/books/${mockBookId}/reviews` && options?.method === 'POST') {
        console.log('Mocking POST /api/books/:id/reviews');
        const newCommentId = 'comment-new';
        const requestBody = JSON.parse(options.body);
        // Return structure matching model.Comment
        return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
                id: newCommentId,
                body: requestBody.Body, // Use Body from request
                bookID: mockBookId,
                date: new Date().toISOString(),
                poster: { id: mockUserId, name: 'Me User', username: 'me#2222', avatar: 'avatar-me-id' }, // Match model.CommentUser
                votes: 1, // Starts with 1 vote from poster (based on db repo)
                edited: null,
                deleted: false,
                rating: 0,
                parent: '00000000-0000-0000-0000-000000000000',
            }),
        });
    }
    // PATCH /api/comments/:id (Edit comment)
    else if (url.startsWith('/api/comments/') && options?.method === 'PATCH') {
        const commentId = url.split('/')[3];
        console.log(`Mocking PATCH /api/comments/${commentId}`);
        const requestBody = JSON.parse(options.body);
        const originalComment = mockCommentsData.find(c => c.id === commentId);
        // Return structure matching model.Comment
        return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
                ...originalComment, // Keep original data
                body: requestBody.Body, // Update body
                edited: new Date().toISOString(), // Set edited timestamp
            }),
        });
    }
    // DELETE /api/comments/:id
    else if (url.startsWith('/api/comments/') && options?.method === 'DELETE') {
        const commentId = url.split('/')[3];
        console.log(`Mocking DELETE /api/comments/${commentId}`);
        // Check if user has permission (simple check based on mockUserId)
        const commentToDelete = mockCommentsData.find(c => c.id === commentId);
        if (commentToDelete?.poster?.id !== mockUserId) {
            return Promise.resolve({
                ok: false,
                status: 403,
                json: () => Promise.resolve({ summary: "Forbidden" }),
            });
        }
        return Promise.resolve({ ok: true }); // Simple success
    }
    // POST /api/comments/:id/vote
    else if (url.includes('/vote') && options?.method === 'POST') {
        const commentId = url.split('/')[3];
        const voteValue = parseInt(new URLSearchParams(url.split('?')[1]).get('vote'), 10);
        console.log(`Mocking POST /api/comments/${commentId}/vote?vote=${voteValue}`);
        const originalComment = mockCommentsData.find(c => c.id === commentId);
        const currentTotal = originalComment?.votes || 0;
        const currentUserVote = mockVotesData[commentId] || 0;
        let newTotal = currentTotal;

        if (voteValue === currentUserVote) { // Undoing vote
            newTotal -= voteValue;
        } else { // Applying new vote (or changing vote)
            newTotal += (voteValue - currentUserVote);
        }
        // Update mock data for subsequent calls in the same test
        mockVotesData[commentId] = voteValue;
        if (originalComment) originalComment.votes = newTotal;

        // Return the new total vote count for this comment in the expected map format
        return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ [commentId]: newTotal }),
        });
    }

    console.error('Unhandled fetch mock:', url, options);
    return Promise.reject(new Error(`Unknown URL/Method: ${url} ${options?.method}`));
  });
});

afterEach(() => {
  jest.restoreAllMocks();
  // Reset mock data changes if necessary between tests
  mockCommentsData[0].votes = 5;
  mockCommentsData[1].votes = 3;
  mockVotesData['comment-1'] = 0;
  mockVotesData['comment-2'] = 1;
});

test('renders Comments heading and initial comments with votes', async () => {
  render(<Comments bookId={mockBookId} jwt={mockJwt} />);

  // Check heading
  expect(screen.getByRole('heading', { name: /Comments/i })).toBeInTheDocument();

  // Wait for comments to load and display
  await waitFor(() => {
    expect(screen.getByText('This is the first comment.')).toBeInTheDocument();
    expect(screen.getByText('Other User')).toBeInTheDocument(); // Check author name
    expect(screen.getByText('This is the second comment, by me.')).toBeInTheDocument();
    expect(screen.getByText('Me User')).toBeInTheDocument(); // Check author name
  });

  // Check initial vote counts and user vote status
  const comment1Item = screen.getByText('This is the first comment.').closest('li');
  const comment2Item = screen.getByText('This is the second comment, by me.').closest('li');

  const comment1Votes = within(comment1Item).getByText(/5/); // Find vote count within the item
  const comment2Votes = within(comment2Item).getByText(/3/);
  expect(comment1Votes).toBeInTheDocument();
  expect(comment2Votes).toBeInTheDocument();

  const comment1UpvoteButton = within(comment1Item).getByRole('button', { name: /Upvote/i });
  const comment2UpvoteButton = within(comment2Item).getByRole('button', { name: /Upvote/i });
  expect(comment1UpvoteButton).not.toHaveClass('active'); // User hasn't voted on comment 1
  expect(comment2UpvoteButton).toHaveClass('active'); // User has upvoted comment 2

  // Verify initial fetch calls
  expect(fetch).toHaveBeenCalledWith(`/api/books/${mockBookId}/reviews`, expect.any(Object));
  expect(fetch).toHaveBeenCalledWith(`/api/books/${mockBookId}/reviews/votes`, expect.any(Object));
});

test('allows adding a new comment', async () => {
  render(<Comments bookId={mockBookId} jwt={mockJwt} />);

  // Wait for initial load just in case
  await screen.findByText('This is the first comment.');

  const textarea = screen.getByPlaceholderText(/Add a comment/i);
  const postButton = screen.getByRole('button', { name: /Post Comment/i });

  // Type comment and click post
  fireEvent.change(textarea, { target: { value: 'A new comment!' } });
  fireEvent.click(postButton);

  // Check for optimistic update (shows 'You' and the text)
  await waitFor(() => {
    expect(screen.getByText('A new comment!')).toBeInTheDocument();
    expect(screen.getByText('You')).toBeInTheDocument(); // Optimistic author
  });

  // Wait for fetch to complete and final state
  await waitFor(() => {
    expect(fetch).toHaveBeenCalledWith(`/api/books/${mockBookId}/reviews`, expect.objectContaining({ method: 'POST' }));
    // Check that the author name updated from 'You' to the actual name from response
    expect(screen.getByText('Me User')).toBeInTheDocument();
    // Check new comment vote count (should be 1 based on mock response)
    const newCommentItem = screen.getByText('A new comment!').closest('li');
    expect(within(newCommentItem).getByText(/1/)).toBeInTheDocument(); // Vote count is 1
    // Check textarea is cleared
    expect(textarea.value).toBe('');
  });
});

test('allows editing own comment', async () => {
  render(<Comments bookId={mockBookId} jwt={mockJwt} />);

  // Wait for comments to load
  const comment2Text = await screen.findByText('This is the second comment, by me.');
  const comment2Item = comment2Text.closest('li');

  // Find the edit button for the user's own comment
  const editButton = within(comment2Item).getByRole('button', { name: /Edit/i });
  fireEvent.click(editButton);

  // Edit the text
  const editTextarea = within(comment2Item).getByRole('textbox');
  fireEvent.change(editTextarea, { target: { value: 'Updated comment text.' } });

  // Save changes
  const saveButton = within(comment2Item).getByRole('button', { name: /Save/i });
  fireEvent.click(saveButton);

  // Wait for fetch and UI update
  await waitFor(() => {
    expect(fetch).toHaveBeenCalledWith('/api/comments/comment-2', expect.objectContaining({ method: 'PATCH' }));
    expect(screen.getByText('Updated comment text.')).toBeInTheDocument();
    // Check for (edited) marker - depends on timing logic in component/mock
    expect(within(comment2Item).getByText(/\(edited .*\)/i)).toBeInTheDocument();
  });
});

test('allows deleting own comment', async () => {
    render(<Comments bookId={mockBookId} jwt={mockJwt} />);

    // Wait for comments to load
    const comment2Text = await screen.findByText('This is the second comment, by me.');
    const comment2Item = comment2Text.closest('li');

    // Find the delete button for the user's own comment
    const deleteButton = within(comment2Item).getByRole('button', { name: /Delete/i });
    fireEvent.click(deleteButton);

    // Wait for fetch and UI update (comment should disappear)
    await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith('/api/comments/comment-2', expect.objectContaining({ method: 'DELETE' }));
        expect(screen.queryByText('This is the second comment, by me.')).not.toBeInTheDocument();
    });
});

test('prevents deleting others comment', async () => {
    render(<Comments bookId={mockBookId} jwt={mockJwt} />);

    // Wait for comments to load
    const comment1Text = await screen.findByText('This is the first comment.');
    const comment1Item = comment1Text.closest('li');

    // Check that delete button is NOT present for other user's comment
    expect(within(comment1Item).queryByRole('button', { name: /Delete/i })).not.toBeInTheDocument();
});


test('allows voting on comments and undoing votes', async () => {
    render(<Comments bookId={mockBookId} jwt={mockJwt} />);

    // Wait for comments to load
    const comment1Text = await screen.findByText('This is the first comment.');
    const comment1Item = comment1Text.closest('li');

    const upvoteButton = within(comment1Item).getByRole('button', { name: /Upvote/i });
    const downvoteButton = within(comment1Item).getByRole('button', { name: /Downvote/i });
    const voteCount = within(comment1Item).querySelector('.vote-count');

    // --- Initial state: 5 votes, user hasn't voted ---
    expect(voteCount).toHaveTextContent('5');
    expect(upvoteButton).not.toHaveClass('active');
    expect(downvoteButton).not.toHaveClass('active');

    // --- 1. Upvote ---
    fireEvent.click(upvoteButton);
    // Check optimistic update
    expect(voteCount).toHaveTextContent('6'); // 5 + 1
    expect(upvoteButton).toHaveClass('active');
    expect(downvoteButton).not.toHaveClass('active');
    // Wait for fetch confirmation
    await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith('/api/comments/comment-1/vote?vote=1', expect.objectContaining({ method: 'POST' }));
        // Vote count should remain 6 based on mock response
        expect(voteCount).toHaveTextContent('6');
    });

    // --- 2. Click Upvote again (Undo) ---
    fireEvent.click(upvoteButton);
    // Check optimistic update
    expect(voteCount).toHaveTextContent('5'); // 6 - 1
    expect(upvoteButton).not.toHaveClass('active');
    expect(downvoteButton).not.toHaveClass('active');
    // Wait for fetch confirmation
     await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith('/api/comments/comment-1/vote?vote=0', expect.objectContaining({ method: 'POST' }));
        // Vote count should remain 5 based on mock response
        expect(voteCount).toHaveTextContent('5');
    });

    // --- 3. Downvote ---
    fireEvent.click(downvoteButton);
    // Check optimistic update
    expect(voteCount).toHaveTextContent('4'); // 5 - 1
    expect(upvoteButton).not.toHaveClass('active');
    expect(downvoteButton).toHaveClass('active');
    // Wait for fetch confirmation
    await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith('/api/comments/comment-1/vote?vote=-1', expect.objectContaining({ method: 'POST' }));
        // Vote count should remain 4 based on mock response
        expect(voteCount).toHaveTextContent('4');
    });

     // --- 4. Click Downvote again (Undo) ---
    fireEvent.click(downvoteButton);
    // Check optimistic update
    expect(voteCount).toHaveTextContent('5'); // 4 + 1
    expect(upvoteButton).not.toHaveClass('active');
    expect(downvoteButton).not.toHaveClass('active');
    // Wait for fetch confirmation
    await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith('/api/comments/comment-1/vote?vote=0', expect.objectContaining({ method: 'POST' }));
        // Vote count should remain 5 based on mock response
        expect(voteCount).toHaveTextContent('5');
    });

    // --- 5. Upvote from neutral ---
    fireEvent.click(upvoteButton);
    // Check optimistic update
    expect(voteCount).toHaveTextContent('6'); // 5 + 1
    expect(upvoteButton).toHaveClass('active');
    expect(downvoteButton).not.toHaveClass('active');
    // Wait for fetch confirmation
    await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith('/api/comments/comment-1/vote?vote=1', expect.objectContaining({ method: 'POST' }));
        // Vote count should remain 6 based on mock response
        expect(voteCount).toHaveTextContent('6');
    });

    // --- 6. Downvote from upvoted state ---
    fireEvent.click(downvoteButton);
    // Check optimistic update
    expect(voteCount).toHaveTextContent('4'); // 6 - 2 (change is -1 - 1 = -2)
    expect(upvoteButton).not.toHaveClass('active');
    expect(downvoteButton).toHaveClass('active');
    // Wait for fetch confirmation
    await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith('/api/comments/comment-1/vote?vote=-1', expect.objectContaining({ method: 'POST' }));
        // Vote count should remain 4 based on mock response
        expect(voteCount).toHaveTextContent('4');
    });
    expect(true).toBe(true);
});
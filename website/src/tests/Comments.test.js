import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import Comments from '../pages/Comments';

const mockJwt = 'mock-jwt-token';
const mockBookId = 'mock-book-id';

beforeEach(() => {
  global.fetch = jest.fn((url) => {
    if (url.includes('/comments')) {
      return Promise.resolve({
        json: () =>
          Promise.resolve([
            { id: '1', text: 'Comment 1', totalVotes: 5, userVote: 0 },
            { id: '2', text: 'Comment 2', totalVotes: 3, userVote: 1 },
          ]),
      });
    }
    return Promise.reject(new Error('Unknown URL'));
  });
});

afterEach(() => {
  jest.restoreAllMocks();
});

test('renders Comments heading', () => {
  render(<Comments bookId={mockBookId} jwt={mockJwt} />);
  const headingElement = screen.getByText(/Comments/i);
  expect(headingElement).toBeInTheDocument();
});

test('fetches and displays comments', async () => {
  render(<Comments bookId={mockBookId} jwt={mockJwt} />);

  await waitFor(() => {
    expect(screen.getByText('Comment 1')).toBeInTheDocument();
    expect(screen.getByText('Comment 2')).toBeInTheDocument();
  });
});
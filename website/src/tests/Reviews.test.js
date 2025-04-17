import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import Reviews from '../pages/Reviews';

const mockJwt = 'mock-jwt-token';
const mockBookId = 'mock-book-id';

beforeEach(() => {
  global.fetch = jest.fn(() =>
    Promise.resolve({
      json: () =>
        Promise.resolve([
          { id: '1', text: 'Review 1', totalVotes: 10 },
          { id: '2', text: 'Review 2', totalVotes: 7 },
        ]),
    })
  );
});

afterEach(() => {
  jest.restoreAllMocks();
});

test('renders Reviews heading', () => {
  render(<Reviews bookId={mockBookId} jwt={mockJwt} />);
  const headingElement = screen.getByText(/Reviews/i);
  expect(headingElement).toBeInTheDocument();
});

test('fetches and displays reviews', async () => {
  render(<Reviews bookId={mockBookId} jwt={mockJwt} />);

  await waitFor(() => {
    expect(screen.getByText('Review 1')).toBeInTheDocument();
    expect(screen.getByText('Review 2')).toBeInTheDocument();
  });
});
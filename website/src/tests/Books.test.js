import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import Books from '../pages/Books';

const mockJwt = 'mock-jwt-token';

beforeEach(() => {
  global.fetch = jest.fn(() =>
    Promise.resolve({
      json: () =>
        Promise.resolve([
          { uuid: '1', title: 'Book 1', author: 'Author 1', genre: 'Fiction' },
          { uuid: '2', title: 'Book 2', author: 'Author 2', genre: 'Mystery' },
        ]),
    })
  );
});

afterEach(() => {
  jest.restoreAllMocks();
});

test('renders Books heading', () => {
  render(<Books jwt={mockJwt} />);
  const headingElement = screen.getByText(/Books/i);
  expect(headingElement).toBeInTheDocument();
});

test('fetches and displays books', async () => {
  render(<Books jwt={mockJwt} />);

  await waitFor(() => {
    expect(screen.getByText('Book 1')).toBeInTheDocument();
    expect(screen.getByText('Book 2')).toBeInTheDocument();
  });
});
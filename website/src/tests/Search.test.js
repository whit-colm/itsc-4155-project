import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import Search from '../pages/Search';

const mockJwt = 'mock-jwt-token';

beforeEach(() => {
  global.fetch = jest.fn((url) => {
    if (url.includes('fail')) {
      return Promise.resolve({
        ok: false,
      });
    }
    return Promise.resolve({
      ok: true,
      json: () =>
        Promise.resolve({
          results: [
            { uuid: '1', title: 'Book 1', author: 'Author 1', genre: 'Genre 1' },
            { uuid: '2', title: 'Book 2', author: 'Author 2', genre: 'Genre 2' },
          ],
        }),
    });
  });
});

afterEach(() => {
  jest.restoreAllMocks();
});

test('renders Book Search heading', () => {
  render(<Search jwt={mockJwt} />);
  const headingElement = screen.getByText(/Book Search/i);
  expect(headingElement).toBeInTheDocument();
});

test('renders search input and updates value', () => {
  render(<Search jwt={mockJwt} />);
  const inputElement = screen.getByPlaceholderText(/Enter book title/i);
  fireEvent.change(inputElement, { target: { value: '1984' } });
  expect(inputElement.value).toBe('1984');
});

test('fetches and displays search results', async () => {
  render(<Search jwt={mockJwt} />);

  const searchInput = screen.getByPlaceholderText(/Enter book title, author, or genre/i);
  fireEvent.change(searchInput, { target: { value: 'Book' } });

  const searchButton = screen.getByRole('button', { name: /Search/i });
  fireEvent.click(searchButton);

  await waitFor(() => {
    expect(global.fetch).toHaveBeenCalledWith(
      '/api/search?idx=booktitle&q=Book&r=25&o=0',
      {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${mockJwt}`,
          'Content-Type': 'application/json',
        },
      }
    );
  });

  await waitFor(() => {
    expect(screen.queryByText('Error: Search failed')).not.toBeInTheDocument();
    expect(screen.getByText('Book 1')).toBeInTheDocument();
    expect(screen.getByText('Book 2')).toBeInTheDocument();
  });
});

test('displays error message on fetch failure', async () => {
  render(<Search jwt={mockJwt} />);

  const searchInput = screen.getByPlaceholderText(/Enter book title, author, or genre/i);
  fireEvent.change(searchInput, { target: { value: 'fail' } });

  const searchButton = screen.getByRole('button', { name: /Search/i });
  fireEvent.click(searchButton);

  await waitFor(() => {
    expect(screen.getByText('Error: Search failed')).toBeInTheDocument();
  });
});

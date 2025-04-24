import { render, screen, waitFor } from '@testing-library/react';
import BookDetails from '../pages/BookDetails';
import { MemoryRouter, useParams } from 'react-router-dom'; // Import MemoryRouter and useParams

// Mock useParams
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'), // use actual for all non-hook parts
  useParams: jest.fn(),
}));

const mockJwt = 'mock-jwt-token';
const mockBookId = '01953a93-21e7-73da-8a27-fc22aa66a95e'; // Use bookId consistent with useParams
const mockCoverId = 'mock-cover-id';

beforeEach(() => {
  // Reset mocks before each test
  useParams.mockReturnValue({ bookId: mockBookId }); // Mock useParams to return the bookId

  global.fetch = jest.fn((url) => {
    console.log('Fetch called with URL:', url); // Debugging fetch calls

    // Mock for fetching book data based on bookId from params
    if (url.includes(`/api/books/${mockBookId}`)) {
      console.log('Mocking GET /api/books/:bookId');
      return Promise.resolve({
        ok: true,
        json: () =>
          Promise.resolve({
            // Match model.Book structure (fields displayed by BookDetails)
            id: mockBookId,
            title: 'Mock Book Title',
            // NOTE: Component uses book.author, but backend model has AuthorIDs.
            // Mocking what the component expects for now.
            author: 'Mock Author Name',
            published: '2023-01-15', // Use YYYY-MM-DD format
            isbns: [ // Include ISBNs even if not directly displayed by BookDetails itself
              { type: 'isbn10', value: '123456789X' },
              { type: 'isbn13', value: '9781234567897' },
            ],
            // Use bref_cover_image matching model.Book JSON tag
            bref_cover_image: mockCoverId,
            description: 'This is a mock book description.', // Add description
          }),
      });
    }
    // Mock for fetching the cover image blob
    else if (url.includes(`/api/blob/${mockCoverId}`)) {
      console.log('Mocking GET /api/blob/:coverId');
      // Mock URL.createObjectURL
      global.URL.createObjectURL = jest.fn(() => 'mock-blob-url');
      return Promise.resolve({
        ok: true,
        blob: () => Promise.resolve(new Blob(['mock-image-content'], { type: 'image/png' })),
      });
    }
    // Mock for fetching comments (keep simple for BookDetails test)
    else if (url.includes(`/api/books/${mockBookId}/reviews`)) {
      console.log('Mocking GET /api/books/:bookId/reviews');
      return Promise.resolve({ ok: true, json: () => Promise.resolve([]) }); // Return empty array
    }
    // Mock for fetching comment votes (keep simple)
    else if (url.includes(`/api/books/${mockBookId}/reviews/votes`)) {
        console.log('Mocking GET /api/books/:bookId/reviews/votes');
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) }); // Return empty map
    }

    console.error('Unknown URL in fetch mock:', url);
    return Promise.reject(new Error(`Unknown URL: ${url}`));
  });

  // Mock URL.revokeObjectURL if needed
  global.URL.revokeObjectURL = jest.fn();
});

afterEach(() => {
  jest.restoreAllMocks();
  // Clean up URL mocks
  if (global.URL.createObjectURL) delete global.URL.createObjectURL;
  if (global.URL.revokeObjectURL) delete global.URL.revokeObjectURL;
});

test('renders Book title after loading', async () => {
  render(
    <MemoryRouter> {/* Comments component uses Link */}
      {/* Pass jwt, BookDetails gets ID from mocked useParams */}
      <BookDetails jwt={mockJwt} />
    </MemoryRouter>
  );

  // Use findBy* to wait for the element
  const headingElement = await screen.findByRole('heading', { name: /Mock Book Title/i });
  expect(headingElement).toBeInTheDocument();
});

test('renders book details', async () => {
  render(
    <MemoryRouter>
      <BookDetails jwt={mockJwt} />
    </MemoryRouter>
  );

  // Wait for details using findBy* or within a waitFor block
  expect(await screen.findByText('Mock Book Title')).toBeInTheDocument();
  expect(screen.getByText('Mock Author Name')).toBeInTheDocument(); // Component uses book.author
  expect(screen.getByText('This is a mock book description.')).toBeInTheDocument();

  // Check for formatted date - adjust format based on how component displays it
  // Using a flexible check:
  expect(screen.getByText(/Published:/)).toHaveTextContent('1/15/2023'); // Example: MM/DD/YYYY format

  // Verify fetch calls
  expect(fetch).toHaveBeenCalledWith(expect.stringContaining(`/api/books/${mockBookId}`), expect.any(Object));
  expect(true).toBe(true);
});

test('renders cover image', async () => {
  render(
    <MemoryRouter>
      <BookDetails jwt={mockJwt} />
    </MemoryRouter>
  );

  // Wait for the image using findBy*
  const imageElement = await screen.findByAltText(/Mock Book Title cover/i);
  expect(imageElement).toBeInTheDocument();
  expect(imageElement).toHaveAttribute('src', 'mock-blob-url'); // Check the mocked blob URL

  // Verify fetch calls for book and blob
  expect(fetch).toHaveBeenCalledWith(expect.stringContaining(`/api/books/${mockBookId}`), expect.any(Object));
  expect(fetch).toHaveBeenCalledWith(expect.stringContaining(`/api/blob/${mockCoverId}`), expect.any(Object));
  expect(global.URL.createObjectURL).toHaveBeenCalled();
  expect(true).toBe(true);
});

test('renders Comments component', async () => {
    render(
      <MemoryRouter>
        <BookDetails jwt={mockJwt} />
      </MemoryRouter>
    );

    // Wait for the BookDetails to load its data first
    await screen.findByRole('heading', { name: /Mock Book Title/i });

    // Check if the Comments section heading is rendered (indicating Comments component loaded)
    expect(screen.getByRole('heading', { name: /Comments/i })).toBeInTheDocument();

    // Verify the fetch call for comments was made
    expect(fetch).toHaveBeenCalledWith(expect.stringContaining(`/api/books/${mockBookId}/reviews`), expect.any(Object));
    expect(true).toBe(true);
});

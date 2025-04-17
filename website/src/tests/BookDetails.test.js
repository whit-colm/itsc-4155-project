import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import BookDetails from '../pages/BookDetails';

const mockJwt = 'mock-jwt-token';

beforeEach(() => {
  global.fetch = jest.fn((url) => {
    if (url.includes('/api/books/')) {
      return Promise.resolve({
        json: () =>
          Promise.resolve({
            title: 'Mock Book',
            author: 'Mock Author',
            published: '2023-01-01',
            isbns: [
              { type: 'isbn10', value: '123456789X' },
              { type: 'isbn13', value: '9781234567897' },
            ],
            bref_cover: 'mock-cover-id',
          }),
      });
    } else if (url.includes('/api/blob/')) {
      return Promise.resolve({
        blob: () => Promise.resolve(new Blob(['mock-image'], { type: 'image/png' })),
      });
    }
    return Promise.reject(new Error('Unknown URL'));
  });
});

afterEach(() => {
  jest.restoreAllMocks();
});

test('renders Book Details heading', async () => {
  render(<BookDetails uuid="01953a93-21e7-73da-8a27-fc22aa66a95e" />);

  const headingElement = await screen.findByText(/Loading.../i);
  expect(headingElement).toBeInTheDocument();
});

test('renders book details', async () => {
  render(<BookDetails uuid="mock-uuid" jwt={mockJwt} />);

  await waitFor(() => {
    expect(screen.getByText('Mock Book')).toBeInTheDocument();
    expect(screen.getByText('Mock Author')).toBeInTheDocument();
    expect(screen.getByText('2023-01-01')).toBeInTheDocument();
    expect(screen.getByText('123456789X')).toBeInTheDocument();
    expect(screen.getByText('9781234567897')).toBeInTheDocument();
  });
});

test('renders cover image', async () => {
  render(<BookDetails uuid="mock-uuid" jwt={mockJwt} />);

  await waitFor(() => {
    const imageElement = screen.getByAltText(/Mock Book cover/i);
    expect(imageElement).toBeInTheDocument();
  });
});

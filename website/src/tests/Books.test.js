import { render, screen, waitFor } from '@testing-library/react';
import Books from '../pages/Books';
import { MemoryRouter } from 'react-router-dom'; // Import MemoryRouter

const mockJwt = 'mock-jwt-token';

beforeEach(() => {
  global.fetch = jest.fn((url) => {
    // Handle the initial empty query which returns 400
    if (url.includes('/api/search?') && url.includes('q=&')) {
      return Promise.resolve({
        ok: false,
        status: 400,
        json: () => Promise.resolve({ summary: 'Your query must not be empty' }),
      });
    }
    // Mock a successful response for other potential calls (though Books.js doesn't make others)
    // Use BookSummary structure matching the search endpoint response
    return Promise.resolve({
      ok: true,
      json: () =>
        Promise.resolve([
          {
            id: '1', // Use 'id' as per BookSummary
            title: 'Book 1',
            authors: [{ id: 'a1', givenname: 'Author', familyname: 'One' }], // Use authors array
            published: '2023-01-01',
            isbns: [{ type: 'isbn13', value: '978111' }],
            apiVersion: "booksummary.itsc-4155-group-project.edu.whits.io/v1alpha1" // Add apiVersion
          },
          {
            id: '2',
            title: 'Book 2',
            authors: [{ id: 'a2', givenname: 'Author', familyname: 'Two' }],
            published: '2023-02-01',
            isbns: [{ type: 'isbn13', value: '978222' }],
            apiVersion: "booksummary.itsc-4155-group-project.edu.whits.io/v1alpha1"
          },
        ]),
    });
  });
});

afterEach(() => {
  jest.restoreAllMocks();
});

test('renders Books heading and shows "No books found" initially', async () => {
  render(
    <MemoryRouter> {/* Wrap Books in MemoryRouter because it uses Link */}
      <Books jwt={mockJwt} />
    </MemoryRouter>
  );
  // Check for heading
  const headingElement = screen.getByRole('heading', { name: /Books/i });
  expect(headingElement).toBeInTheDocument();

  // Wait for the component to handle the initial 400 error
  await waitFor(() => {
    expect(screen.getByText(/No books found\./i)).toBeInTheDocument();
  });

  // Verify fetch was called for the empty query
  expect(global.fetch).toHaveBeenCalledWith(
    '/api/search?d=booktitle&q=&r=100&o=0',
    expect.any(Object)
  );
});

// Optional: Add a test simulating a successful fetch if the component logic were different
// test('fetches and displays books on success', async () => {
//   // Modify fetch mock specifically for this test if needed
//   global.fetch.mockImplementationOnce((url) => {
//     if (url.includes('/api/search?') && url.includes('q=&')) {
//       // Simulate successful fetch instead of 400
//       return Promise.resolve({
//         ok: true,
//         json: () => Promise.resolve([
//           { id: '1', title: 'Book 1', authors: [{ givenname: 'Author', familyname: 'One' }], published: '2023-01-01', isbns: [], apiVersion: "booksummary..." },
//           { id: '2', title: 'Book 2', authors: [{ givenname: 'Author', familyname: 'Two' }], published: '2023-02-01', isbns: [], apiVersion: "booksummary..." },
//         ]),
//       });
//     }
//     return Promise.reject(new Error('Unexpected fetch call'));
//   });

//   render(
//     <MemoryRouter>
//       <Books jwt={mockJwt} />
//     </MemoryRouter>
//   );

//   await waitFor(() => {
//     expect(screen.getByText('Book 1')).toBeInTheDocument();
//     expect(screen.getByText('Author One')).toBeInTheDocument(); // Check for author name
//     expect(screen.getByText('Book 2')).toBeInTheDocument();
//     expect(screen.getByText('Author Two')).toBeInTheDocument();
//   });
// });
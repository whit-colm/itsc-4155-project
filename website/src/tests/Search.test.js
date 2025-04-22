import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom'; // Import MemoryRouter
import Search from '../pages/Search';

// Mock react-router-dom's Link component
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'), // use actual for all non-hook parts
  Link: ({ children, to }) => <a href={to}>{children}</a>, // Simple anchor mock
}));

// --- Mock Data matching Go models and Search.js expectations ---
const mockBookResult = {
  apiVersion: "booksummary.itsc-4155-group-project.edu.whits.io/v1alpha2",
  id: 'book-123',
  title: 'Mock Book Title',
  authors: [{ given_name: 'Mock', family_name: 'Author' }],
  description: 'A mock book description.',
  // other fields as needed by component...
};
const mockAuthorResult = {
  apiVersion: "author.itsc-4155-group-project.edu.whits.io/v1alpha3",
  id: 'author-456',
  given_name: 'Another',
  family_name: 'Writer',
  bio: 'Bio of the writer.',
  // other fields...
};
const mockCommentResult = {
  apiVersion: "comment.itsc-4155-group-project.edu.whits.io/v1alpha1",
  id: 'comment-789',
  bookID: 'book-abc',
  body: 'This is a comment body preview.',
  poster: { name: 'Commenter Name', username: 'commenter#1111' },
  // other fields...
};

const mockResultsPage1 = [mockBookResult, mockAuthorResult];
const mockResultsPage2 = [mockCommentResult];
// --- End Mock Data ---


beforeEach(() => {
  global.fetch = jest.fn((url) => {
    console.log("Fetch called:", url);
    const urlParams = new URLSearchParams(url.split('?')[1]);
    const query = urlParams.get('q');
    const offset = parseInt(urlParams.get('o') || '0', 10);
    // const limit = parseInt(urlParams.get('r') || '10', 10); // Limit is used in component logic

    if (query === 'fail') {
      return Promise.resolve({
        ok: false,
        status: 500,
        json: () => Promise.resolve({ summary: 'Simulated search failure' }),
      });
    }
    if (query === 'empty') {
        return Promise.resolve({
            ok: true,
            json: () => Promise.resolve([]), // Empty array for no results
        });
    }

    // Simulate pagination
    if (offset === 0 && query === 'multi') {
        return Promise.resolve({
            ok: true,
            json: () => Promise.resolve(mockResultsPage1),
        });
    } else if (offset >= 10 && query === 'multi') { // Assuming limit is 10
         return Promise.resolve({
            ok: true,
            json: () => Promise.resolve(mockResultsPage2),
        });
    }

    // Default successful response
    return Promise.resolve({
      ok: true,
      // Return different results based on query/offset for testing
      json: () => Promise.resolve(mockResultsPage1),
    });
  });
});

afterEach(() => {
  jest.restoreAllMocks();
});

// Helper to render with Router context
const renderWithRouter = (ui) => {
    return render(ui, { wrapper: MemoryRouter });
};


test('renders Search heading', () => {
  renderWithRouter(<Search />);
  // Match heading from Search.js
  const headingElement = screen.getByRole('heading', { name: /Search Books, Authors, Comments/i });
  expect(headingElement).toBeInTheDocument();
});

test('renders search input and updates value', () => {
  renderWithRouter(<Search />);
  // Match placeholder from Search.js
  const inputElement = screen.getByPlaceholderText(/Enter your search query.../i);
  fireEvent.change(inputElement, { target: { value: 'Test Query' } });
  expect(inputElement.value).toBe('Test Query');
});

test('fetches and displays search results based on apiVersion', async () => {
  renderWithRouter(<Search />);

  const searchInput = screen.getByPlaceholderText(/Enter your search query.../i);
  fireEvent.change(searchInput, { target: { value: 'find stuff' } });

  // Wait for fetch debounce and execution
  await waitFor(() => {
    // Default indices are booktitle, authorname, comments
    expect(global.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/search?q=find%20stuff&d=booktitle,authorname,comments&r=10&o=0'),
      // No specific headers/method needed for GET
    );
  });

  // Wait for results to render based on mock data
  await waitFor(() => {
    // Check Book result
    expect(screen.getByRole('heading', { name: /Mock Book Title/i })).toBeInTheDocument();
    expect(screen.getByText(/by Mock Author/i)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Mock Book Title/i })).toHaveAttribute('href', '/books/book-123');

    // Check Author result
    expect(screen.getByRole('heading', { name: /Another Writer/i })).toBeInTheDocument();
    expect(screen.getByText(/Author/i)).toBeInTheDocument(); // Simple detail for author
    expect(screen.getByRole('link', { name: /Another Writer/i })).toHaveAttribute('href', '/authors/author-456');
  });
});

test('updates search when indices change', async () => {
    renderWithRouter(<Search />);

    const searchInput = screen.getByPlaceholderText(/Enter your search query.../i);
    fireEvent.change(searchInput, { target: { value: 'filter test' } });

    // Wait for initial fetch with default indices
    await waitFor(() => {
        expect(global.fetch).toHaveBeenCalledWith(
            expect.stringContaining('d=booktitle,authorname,comments'),
        );
    });

    // Uncheck 'comments'
    const commentsCheckbox = screen.getByLabelText(/Comments/i);
    fireEvent.click(commentsCheckbox);

    // Wait for fetch debounce and execution with updated indices
    await waitFor(() => {
        expect(global.fetch).toHaveBeenCalledWith(
            // Should now exclude 'comments'
            expect.stringContaining('/api/search?q=filter%20test&d=booktitle,authorname&r=10&o=0'),
        );
    });

    // Check 'Book Titles' checkbox (assuming it was checked by default)
    const bookCheckbox = screen.getByLabelText(/Book Titles/i);
    expect(bookCheckbox).toBeChecked();
    // Check 'Author Names' checkbox
    const authorCheckbox = screen.getByLabelText(/Author Names/i);
    expect(authorCheckbox).toBeChecked();
    // Check 'Comments' checkbox is now unchecked
    expect(commentsCheckbox).not.toBeChecked();
});


test('handles pagination correctly', async () => {
    renderWithRouter(<Search />);

    const searchInput = screen.getByPlaceholderText(/Enter your search query.../i);
    // Use 'multi' query to trigger pagination mock
    fireEvent.change(searchInput, { target: { value: 'multi' } });

    // Wait for initial fetch (page 1)
    await waitFor(() => {
        expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('o=0'));
        expect(screen.getByRole('heading', { name: /Mock Book Title/i })).toBeInTheDocument();
        expect(screen.getByRole('heading', { name: /Another Writer/i })).toBeInTheDocument();
    });

    const nextButton = screen.getByRole('button', { name: /Next/i });
    const prevButton = screen.getByRole('button', { name: /Previous/i });

    // Check initial state
    expect(prevButton).toBeDisabled();
    // Next button enabled depends on hasNextPage logic (mock returns 2 results, limit 10 -> hasNextPage=false initially)
    // Let's adjust mock to return 10 items for page 1 to enable 'Next'
    global.fetch.mockImplementation((url) => {
        const urlParams = new URLSearchParams(url.split('?')[1]);
        const offset = parseInt(urlParams.get('o') || '0', 10);
        if (offset === 0) return Promise.resolve({ ok: true, json: () => Promise.resolve(Array(10).fill(mockBookResult)) });
        if (offset >= 10) return Promise.resolve({ ok: true, json: () => Promise.resolve(mockResultsPage2) });
        return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
    });

    // Re-trigger search to get new mock response
    fireEvent.change(searchInput, { target: { value: 'multi ' } }); // Add space to re-trigger
    fireEvent.change(searchInput, { target: { value: 'multi' } });

    await waitFor(() => {
         expect(screen.getAllByRole('heading', { name: /Mock Book Title/i }).length).toBe(10);
         expect(nextButton).not.toBeDisabled(); // Now should be enabled
    });


    // Click Next
    fireEvent.click(nextButton);

    // Wait for fetch for page 2 (offset 10)
    await waitFor(() => {
        expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('o=10'));
        // Check for page 2 results (Comment)
        expect(screen.getByRole('heading', { name: /Comment by Commenter Name/i })).toBeInTheDocument();
        // Check page 1 results are gone
        expect(screen.queryByRole('heading', { name: /Mock Book Title/i })).not.toBeInTheDocument();
    });

    // Check button states on page 2
    expect(prevButton).not.toBeDisabled();
    // Next button disabled as mockResultsPage2 has only 1 item (< limit)
    expect(nextButton).toBeDisabled();

    // Click Previous
    fireEvent.click(prevButton);

     // Wait for fetch for page 1 (offset 0)
    await waitFor(() => {
        expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('o=0'));
        // Check for page 1 results again
        expect(screen.getAllByRole('heading', { name: /Mock Book Title/i }).length).toBe(10);
        // Check page 2 results are gone
        expect(screen.queryByRole('heading', { name: /Comment by Commenter Name/i })).not.toBeInTheDocument();
    });

    // Check button states back on page 1
    expect(prevButton).toBeDisabled();
    expect(nextButton).not.toBeDisabled();
});


test('displays error message on fetch failure', async () => {
  renderWithRouter(<Search />);

  const searchInput = screen.getByPlaceholderText(/Enter your search query.../i);
  fireEvent.change(searchInput, { target: { value: 'fail' } });

  // Wait for fetch debounce and execution
  await waitFor(() => {
    expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('q=fail'));
  });

  // Wait for error message from JSON response
  await waitFor(() => {
    expect(screen.getByText('Simulated search failure')).toBeInTheDocument();
  });
});

test('displays "No results found" message', async () => {
    renderWithRouter(<Search />);

    const searchInput = screen.getByPlaceholderText(/Enter your search query.../i);
    fireEvent.change(searchInput, { target: { value: 'empty' } }); // Use 'empty' query for mock

    // Wait for fetch debounce and execution
    await waitFor(() => {
        expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('q=empty'));
    });

    // Wait for the "No results" message
    await waitFor(() => {
        expect(screen.getByText(/No results found for "empty"./i)).toBeInTheDocument();
    });
});

test('clears results when query is empty', async () => {
    renderWithRouter(<Search />);

    const searchInput = screen.getByPlaceholderText(/Enter your search query.../i);
    fireEvent.change(searchInput, { target: { value: 'initial search' } });

    // Wait for initial results
    await waitFor(() => {
        expect(screen.getByRole('heading', { name: /Mock Book Title/i })).toBeInTheDocument();
    });

    // Clear the search input
    fireEvent.change(searchInput, { target: { value: '' } });

    // Wait for results to disappear
    await waitFor(() => {
        expect(screen.queryByRole('heading', { name: /Mock Book Title/i })).not.toBeInTheDocument();
        // Fetch should not be called again for empty query
        // Check the last call was for 'initial search'
        expect(global.fetch).toHaveBeenLastCalledWith(expect.stringContaining('q=initial%20search'));
    });
});

test('shows error if no indices are selected', async () => {
    renderWithRouter(<Search />);

    const searchInput = screen.getByPlaceholderText(/Enter your search query.../i);
    fireEvent.change(searchInput, { target: { value: 'test query' } });

    // Uncheck all default indices
    fireEvent.click(screen.getByLabelText(/Book Titles/i));
    fireEvent.click(screen.getByLabelText(/Author Names/i));
    fireEvent.click(screen.getByLabelText(/Comments/i));

    // Wait for error message to appear
    await waitFor(() => {
        expect(screen.getByText("Please select at least one field to search.")).toBeInTheDocument();
        // Fetch should not be called if no indices are selected
        expect(global.fetch).not.toHaveBeenCalledWith(expect.stringContaining('q=test%20query'));
    });
});

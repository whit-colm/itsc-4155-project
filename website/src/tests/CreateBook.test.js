import { render, screen, fireEvent, waitFor } from '@testing-library/react'; // Add waitFor
import CreateBook from '../pages/CreateBook';
import { MemoryRouter, useNavigate } from 'react-router-dom'; // Import MemoryRouter and useNavigate

// Mock useNavigate
const mockedNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
   ...jest.requireActual('react-router-dom'), // use actual for all non-hook parts
  useNavigate: () => mockedNavigate,
}));

// Helper function to render with Router context
const renderWithRouter = (ui, { route = '/' } = {}) => {
  window.history.pushState({}, 'Test page', route);
  return render(ui, { wrapper: MemoryRouter });
};

beforeEach(() => {
  // Mock fetch for potential submission
  global.fetch = jest.fn(() => Promise.resolve({ ok: true, json: () => Promise.resolve({ id: 'new-book-id' }) }));
  // Mock document.cookie for JWT retrieval
  Object.defineProperty(document, 'cookie', {
    writable: true,
    value: 'jwt=mock-jwt-token',
  });
  mockedNavigate.mockClear(); // Clear navigate mock before each test
});

afterEach(() => {
  jest.restoreAllMocks();
});

test('renders Create Book heading', () => {
  renderWithRouter(<CreateBook />);
  const headingElement = screen.getByRole('heading', { name: /Create Book/i });
  expect(headingElement).toBeInTheDocument();
});

test('renders title input', () => {
  renderWithRouter(<CreateBook />);
  const inputElement = screen.getByPlaceholderText(/Title/i);
  expect(inputElement).toBeInTheDocument();
});

test('renders author input with correct placeholder', () => {
  renderWithRouter(<CreateBook />);
  // Match the updated placeholder text from CreateBook.js
  const inputElement = screen.getByPlaceholderText(/Author Name \(Not sent - requires Author ID\)/i);
  expect(inputElement).toBeInTheDocument();
});

test('renders published date input', () => {
  renderWithRouter(<CreateBook />);
  // Input type="date" doesn't always show placeholder reliably, check by label or other means if needed
  // For now, just check its presence
  const inputElement = screen.getByLabelText(/ISBNs:/i).closest('form').querySelector('input[type="date"]');
  expect(inputElement).toBeInTheDocument();
});

test('renders add ISBN button initially', () => {
  renderWithRouter(<CreateBook />);
  const buttonElement = screen.getByRole('button', { name: /Add ISBN-13/i });
  expect(buttonElement).toBeInTheDocument();
  expect(screen.queryByRole('button', { name: /Add ISBN-10/i })).not.toBeInTheDocument();
});

test('adds and removes ISBN input', () => {
  renderWithRouter(<CreateBook />);
  expect(screen.getAllByPlaceholderText(/ISBN/i).length).toBe(1);
  expect(screen.getByPlaceholderText(/Enter ISBN-10/i)).toBeInTheDocument();

  const addButton13 = screen.getByRole('button', { name: /Add ISBN-13/i });
  fireEvent.click(addButton13);

  const isbnInputs = screen.getAllByPlaceholderText(/ISBN/i);
  expect(isbnInputs.length).toBe(2);
  expect(screen.getByPlaceholderText(/Enter ISBN-10/i)).toBeInTheDocument();
  expect(screen.getByPlaceholderText(/Enter ISBN-13/i)).toBeInTheDocument();
  expect(screen.queryByRole('button', { name: /Add ISBN-10/i })).not.toBeInTheDocument();
  expect(screen.queryByRole('button', { name: /Add ISBN-13/i })).not.toBeInTheDocument();

  const removeButtons = screen.getAllByRole('button', { name: /X/i });
  expect(removeButtons.length).toBe(2);

  fireEvent.click(removeButtons[1]);

  expect(screen.getAllByPlaceholderText(/ISBN/i).length).toBe(1);
  expect(screen.getByPlaceholderText(/Enter ISBN-10/i)).toBeInTheDocument();
  expect(screen.queryByPlaceholderText(/Enter ISBN-13/i)).not.toBeInTheDocument();
  expect(screen.getByRole('button', { name: /Add ISBN-13/i })).toBeInTheDocument();
});

test('submits correct data on form submission', async () => {
  renderWithRouter(<CreateBook />);

  // Fill form fields
  fireEvent.change(screen.getByPlaceholderText(/Title/i), { target: { value: 'Test Book Title' } });
  // Author name is filled in UI but not sent in payload
  fireEvent.change(screen.getByPlaceholderText(/Author Name/i), { target: { value: 'Test Author Name' } });
  // Find date input more reliably
  const dateInput = screen.getByLabelText(/ISBNs:/i).closest('form').querySelector('input[type="date"]');
  fireEvent.change(dateInput, { target: { value: '2024-01-15' } });
  fireEvent.change(screen.getByPlaceholderText(/Enter ISBN-10/i), { target: { value: '123456789X' } });

  // Add and fill ISBN-13
  fireEvent.click(screen.getByRole('button', { name: /Add ISBN-13/i }));
  fireEvent.change(screen.getByPlaceholderText(/Enter ISBN-13/i), { target: { value: '978-1234567897' } });

  // Submit form
  const submitButton = screen.getByRole('button', { name: /Create Book/i });
  fireEvent.click(submitButton);

  // Wait for fetch to be called
  await waitFor(() => {
    expect(global.fetch).toHaveBeenCalledWith(
      '/api/books/new',
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer mock-jwt-token',
          'Content-Type': 'application/json',
        },
        // Verify the payload matches the expected structure from CreateBook.js handleSubmit
        body: JSON.stringify({
          title: 'Test Book Title',
          published: '2024-01-15', // Should be in YYYY-MM-DD format
          isbns: [
            { type: 'isbn10', value: '123456789X' }, // Value should be cleaned
            { type: 'isbn13', value: '9781234567897' } // Value should be cleaned
          ],
          // Author and image fields should be omitted as per component logic
        }),
      })
    );
  });

  // Check if navigation happens on success
  await waitFor(() => {
    // Check navigate was called with the ID from the mock fetch response
    expect(mockedNavigate).toHaveBeenCalledWith('/books/new-book-id');
  });
});

test('validates ISBN format on submit', async () => {
  renderWithRouter(<CreateBook />);

  // Fill required fields
  fireEvent.change(screen.getByPlaceholderText(/Title/i), { target: { value: 'Test Book Title' } });
  const dateInput = screen.getByLabelText(/ISBNs:/i).closest('form').querySelector('input[type="date"]');
  fireEvent.change(dateInput, { target: { value: '2024-01-15' } });

  // Enter invalid ISBN
  const isbn10Input = screen.getByPlaceholderText(/Enter ISBN-10/i);
  fireEvent.change(isbn10Input, { target: { value: 'invalid-isbn' } });

  const submitButton = screen.getByRole('button', { name: /Create Book/i });
  fireEvent.click(submitButton);

  // Check for validation error message displayed by the component
  await waitFor(() => {
    expect(screen.getByText('Invalid ISBN-10 format')).toBeInTheDocument();
  });

  // Check that fetch was NOT called due to validation failure
  expect(global.fetch).not.toHaveBeenCalled();
  // Check that navigation did NOT happen
  expect(mockedNavigate).not.toHaveBeenCalled();
});

test('displays error message on fetch failure', async () => {
  // Override fetch mock for this test to simulate failure
  global.fetch.mockImplementationOnce(() => Promise.resolve({
    ok: false,
    status: 500,
    statusText: 'Internal Server Error',
    json: () => Promise.resolve({ summary: 'Simulated creation failure' }),
  }));

  renderWithRouter(<CreateBook />);

  // Fill form fields with valid data
  fireEvent.change(screen.getByPlaceholderText(/Title/i), { target: { value: 'Test Book Title' } });
  const dateInput = screen.getByLabelText(/ISBNs:/i).closest('form').querySelector('input[type="date"]');
  fireEvent.change(dateInput, { target: { value: '2024-01-15' } });
  fireEvent.change(screen.getByPlaceholderText(/Enter ISBN-10/i), { target: { value: '123456789X' } });

  // Submit form
  const submitButton = screen.getByRole('button', { name: /Create Book/i });
  fireEvent.click(submitButton);

  // Wait for fetch to be called and error message to display
  await waitFor(() => {
    expect(global.fetch).toHaveBeenCalled();
    // Check for the specific error summary from the mock response
    expect(screen.getByText('Simulated creation failure')).toBeInTheDocument();
  });

  // Check that navigation did NOT happen
  expect(mockedNavigate).not.toHaveBeenCalled();
});

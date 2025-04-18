import { render, screen, fireEvent } from '@testing-library/react';
import CreateBook from '../pages/CreateBook';
import { MemoryRouter } from 'react-router-dom'; // Import MemoryRouter

// Helper function to render with Router context
const renderWithRouter = (ui, { route = '/' } = {}) => {
  window.history.pushState({}, 'Test page', route);
  return render(ui, { wrapper: MemoryRouter });
};

beforeEach(() => {
  // Mock fetch for potential submission (optional, depends on test scope)
  global.fetch = jest.fn(() => Promise.resolve({ ok: true, json: () => Promise.resolve({ id: 'new-book-id' }) }));
  // Mock document.cookie for JWT retrieval
  Object.defineProperty(document, 'cookie', {
    writable: true,
    value: 'jwt=mock-jwt-token',
  });
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

test('renders author input', () => {
  renderWithRouter(<CreateBook />);
  const inputElement = screen.getByPlaceholderText(/Author/i);
  expect(inputElement).toBeInTheDocument();
});

test('renders published date input', () => {
  renderWithRouter(<CreateBook />);
  const inputElement = screen.getByPlaceholderText(/Published Date/i);
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

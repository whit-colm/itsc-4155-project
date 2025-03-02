import { render, screen, fireEvent } from '@testing-library/react';
import CreateBook from '../pages/CreateBook';

test('renders Create Book heading', () => {
  render(<CreateBook />);
  const headingElement = screen.getByRole('heading', { name: /Create Book/i });
  expect(headingElement).toBeInTheDocument();
});

test('renders title input', () => {
  render(<CreateBook />);
  const inputElement = screen.getByPlaceholderText(/Title/i);
  expect(inputElement).toBeInTheDocument();
});

test('renders author input', () => {
  render(<CreateBook />);
  const inputElement = screen.getByPlaceholderText(/Author/i);
  expect(inputElement).toBeInTheDocument();
});

test('renders published date input', () => {
  render(<CreateBook />);
  const inputElement = screen.getByPlaceholderText(/Published Date/i);
  expect(inputElement).toBeInTheDocument();
});

test('renders add ISBN button', () => {
  render(<CreateBook />);
  const buttonElement = screen.getByRole('button', { name: /Add ISBN-13/i });
  expect(buttonElement).toBeInTheDocument();
});

test('adds and removes ISBN input', () => {
  render(<CreateBook />);
  const addButton = screen.getByRole('button', { name: /Add ISBN-13/i });
  fireEvent.click(addButton);
  const isbnInputs = screen.getAllByPlaceholderText(/ISBN/i);
  expect(isbnInputs.length).toBe(2); // Expecting two ISBN inputs after adding one
  const removeButton = screen.getAllByRole('button', { name: /X/i })[1];
  fireEvent.click(removeButton);
  expect(screen.getAllByPlaceholderText(/ISBN/i).length).toBe(1); // Expecting one ISBN input after removing one
});

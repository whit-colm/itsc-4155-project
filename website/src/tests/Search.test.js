import { render, screen, fireEvent } from '@testing-library/react';
import Search from '../pages/Search';

test('renders Book Search heading', () => {
  render(<Search />);
  const headingElement = screen.getByText(/Book Search/i);
  expect(headingElement).toBeInTheDocument();
});

test('renders search input', () => {
  render(<Search />);
  const inputElement = screen.getByPlaceholderText(/Enter book title/i);
  expect(inputElement).toBeInTheDocument();
});

test('renders search button', () => {
  render(<Search />);
  const buttonElements = screen.getAllByText(/Search/i);
  expect(buttonElements.length).toBeGreaterThan(0);
});

test('updates input value on change', () => {
  render(<Search />);
  const inputElement = screen.getByPlaceholderText(/Enter book title/i);
  fireEvent.change(inputElement, { target: { value: '1984' } });
  expect(inputElement.value).toBe('1984');
});

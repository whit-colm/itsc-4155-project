import { render, screen, fireEvent } from '@testing-library/react';
import Login from '../pages/Login';

test('renders login heading', () => {
  render(<Login />);
  const headingElement = screen.getByRole('heading', { name: /Sign in with GitHub/i });
  expect(headingElement).toBeInTheDocument();
});

test('renders GitHub login button', () => {
  render(<Login />);
  const buttonElement = screen.getByRole('button', { name: /Sign in with GitHub/i });
  expect(buttonElement).toBeInTheDocument();
});

test('redirects to GitHub login on button click', () => {
  const originalLocation = window.location;
  delete window.location;
  window.location = { href: '' }; // Mock window.location

  render(<Login />);
  const buttonElement = screen.getByRole('button', { name: /Sign in with GitHub/i });
  fireEvent.click(buttonElement);
  expect(window.location.href).toBe('/api/auth/github/login');

  window.location = originalLocation; // Restore original location
});

import { render, screen, fireEvent } from '@testing-library/react';
import Login from '../pages/Login';

test('renders Login heading', () => {
  render(<Login />);
  const headingElement = screen.getByRole('heading', { name: /Login/i });
  expect(headingElement).toBeInTheDocument();
});

test('logs in on form submit', () => {
  render(<Login />);
  const usernameInput = screen.getByPlaceholderText(/Username/i);
  const passwordInput = screen.getByPlaceholderText(/Password/i);
  const submitButton = screen.getByRole('button', { name: /Login/i });

  fireEvent.change(usernameInput, { target: { value: 'user123' } });
  fireEvent.change(passwordInput, { target: { value: 'password' } });
  fireEvent.click(submitButton);

  expect(screen.getByDisplayValue(/user123/i)).toBeInTheDocument();
  expect(screen.getByDisplayValue(/password/i)).toBeInTheDocument();
});

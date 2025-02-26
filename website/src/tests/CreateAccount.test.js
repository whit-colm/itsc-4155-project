import { render, screen, fireEvent } from '@testing-library/react';
import CreateAccount from '../pages/CreateAccount';

test('renders Create Account heading', () => {
  render(<CreateAccount />);
  const headingElement = screen.getByRole('heading', { name: /Create Account/i });
  expect(headingElement).toBeInTheDocument();
});

test('creates account on form submit', () => {
  render(<CreateAccount />);
  const nameInput = screen.getByPlaceholderText(/Name/i);
  const emailInput = screen.getByPlaceholderText(/Email/i);
  const passwordInput = screen.getByPlaceholderText(/Password/i);
  const submitButton = screen.getByRole('button', { name: /Create Account/i });

  fireEvent.change(nameInput, { target: { value: 'Jane Doe' } });
  fireEvent.change(emailInput, { target: { value: 'jane@example.com' } });
  fireEvent.change(passwordInput, { target: { value: 'password' } });
  fireEvent.click(submitButton);

  expect(screen.getByDisplayValue(/Jane Doe/i)).toBeInTheDocument();
  expect(screen.getByDisplayValue(/jane@example.com/i)).toBeInTheDocument();
  expect(screen.getByDisplayValue(/password/i)).toBeInTheDocument();
});

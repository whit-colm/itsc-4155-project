import { render, screen, fireEvent } from '@testing-library/react';
import Profile from '../pages/Profile';

test('renders Profile Settings heading', () => {
  render(<Profile />);
  const headingElement = screen.getByText(/Profile Settings/i);
  expect(headingElement).toBeInTheDocument();
});

test('updates profile on form submit', () => {
  render(<Profile />);
  const editButton = screen.getByText(/Edit/i);
  fireEvent.click(editButton);

  const nameInput = screen.getByPlaceholderText(/Name/i);
  const emailInput = screen.getByPlaceholderText(/Email/i);
  const submitButton = screen.getByText(/Save Changes/i);

  fireEvent.change(nameInput, { target: { value: 'John Doe' } });
  fireEvent.change(emailInput, { target: { value: 'john.doe@example.com' } });
  fireEvent.click(submitButton);

  expect(screen.getByText(/John Doe/i)).toBeInTheDocument();
  expect(screen.getByText(/john.doe@example.com/i)).toBeInTheDocument();
});

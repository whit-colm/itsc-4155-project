import { render, screen, fireEvent } from '@testing-library/react';
import Profile from '../pages/Profile';

test('renders Profile Settings heading', () => {
  render(<Profile />);
  const headingElement = screen.getByText(/Profile Settings/i);
  expect(headingElement).toBeInTheDocument();
});

test('updates profile on form submit', () => {
  render(<Profile />);
  const nameInput = screen.getByPlaceholderText(/Name/i);
  const genresInput = screen.getByPlaceholderText(/Favorite Genres/i);
  const preferencesTextarea = screen.getByPlaceholderText(/Reading Preferences/i);
  const submitButton = screen.getByText(/Save Changes/i);

  fireEvent.change(nameInput, { target: { value: 'John Doe' } });
  fireEvent.change(genresInput, { target: { value: 'Fiction' } });
  fireEvent.change(preferencesTextarea, { target: { value: 'Likes thrillers' } });
  fireEvent.click(submitButton);

  expect(screen.getByDisplayValue(/John Doe/i)).toBeInTheDocument();
  expect(screen.getByDisplayValue(/Fiction/i)).toBeInTheDocument();
  expect(screen.getByDisplayValue(/Likes thrillers/i)).toBeInTheDocument();
});

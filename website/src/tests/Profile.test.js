import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import Profile from '../pages/Profile';

const mockJwt = 'mock-jwt-token';

beforeEach(() => {
  global.fetch = jest.fn((url) => {
    if (url.includes('/api/user/me')) {
      return Promise.resolve({
        json: () =>
          Promise.resolve({
            name: 'John Doe',
            username: 'johndoe',
            email: 'johndoe@example.com',
            pronouns: 'he/him',
            avatar: 'mock-avatar-id',
            bref_avatar: 'mock-bref-avatar-id',
          }),
      });
    } else if (url.includes('/api/blob/')) {
      return Promise.resolve({
        blob: () => Promise.resolve(new Blob(['mock-image'], { type: 'image/png' })),
      });
    }
    return Promise.reject(new Error('Unknown URL'));
  });
});

afterEach(() => {
  jest.restoreAllMocks();
});

test('renders Profile Settings heading', () => {
  render(<Profile jwt={mockJwt} />);
  const headingElement = screen.getByText(/Profile Settings/i);
  expect(headingElement).toBeInTheDocument();
});

test('renders user information', async () => {
  render(<Profile jwt={mockJwt} />);

  await waitFor(() => {
    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('johndoe')).toBeInTheDocument();
    expect(screen.getByText('johndoe@example.com')).toBeInTheDocument();
    expect(screen.getByText('he/him')).toBeInTheDocument();
  });

  const avatarElement = screen.getByAltText("John Doe's avatar");
  expect(avatarElement).toHaveAttribute('src', '/api/avatars/mock-avatar-id');
});

test('allows editing user information', async () => {
  render(<Profile jwt={mockJwt} />);

  const editButton = screen.getByRole('button', { name: /Edit/i });
  fireEvent.click(editButton);

  const nameInput = screen.getByPlaceholderText(/Name/i);
  fireEvent.change(nameInput, { target: { value: 'Jane Doe' } });

  const saveButton = screen.getByRole('button', { name: /Save Changes/i });
  fireEvent.click(saveButton);

  await waitFor(() => {
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
  });
});

test('handles account deletion', async () => {
  render(<Profile jwt={mockJwt} />);

  // Simulate clicking the Delete Account button
  const deleteButton = screen.getByText(/Delete Account/i);
  fireEvent.click(deleteButton);

  // Wait for the TOTP input field to appear
  const totpInput = await screen.findByPlaceholderText(/Enter TOTP code/i);

  // Simulate entering the TOTP code
  fireEvent.change(totpInput, { target: { value: 'mock-totp-code' } });

  // Simulate clicking the confirm delete button
  const confirmButton = screen.getByText(/Confirm Delete/i);
  fireEvent.click(confirmButton);

  // Wait for the fetch call and verify it was made with the correct arguments
  await waitFor(() => {
    expect(global.fetch).toHaveBeenCalledWith('/api/user/me?code=mock-totp-code', {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${mockJwt}`,
      },
    });
  });
});

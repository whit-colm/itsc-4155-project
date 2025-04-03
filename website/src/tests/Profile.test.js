import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import Profile from '../pages/Profile';

const mockJwt = 'mock-jwt-token';

beforeEach(() => {
  global.fetch = jest.fn(() =>
    Promise.resolve({
      json: () =>
        Promise.resolve({
          id: 'mock-user-id',
          name: 'John Doe',
          username: 'johndoe',
          email: 'johndoe@example.com',
          favoriteGenres: ['Fiction', 'Mystery'],
        }),
    })
  );
});

afterEach(() => {
  jest.restoreAllMocks();
});

test('renders profile settings heading', async () => {
  render(<Profile jwt={mockJwt} />);
  const headingElement = await screen.findByRole('heading', { name: /Profile Settings/i });
  expect(headingElement).toBeInTheDocument();
});

test('renders user information', async () => {
  render(<Profile jwt={mockJwt} />);
  await waitFor(() => {
    expect(screen.getByText(/Name:/i)).toBeInTheDocument();
    expect(screen.getByText(/Username:/i)).toBeInTheDocument();
    expect(screen.getByText(/Email:/i)).toBeInTheDocument();
  });
});

test('renders edit button and toggles edit mode', async () => {
  render(<Profile jwt={mockJwt} />);
  const editButton = await screen.findByRole('button', { name: /Edit/i });
  fireEvent.click(editButton);
  expect(screen.getByRole('button', { name: /Save Changes/i })).toBeInTheDocument();
});

test('updates profile information on form submission', async () => {
  render(<Profile jwt={mockJwt} />);
  fireEvent.click(await screen.findByRole('button', { name: /Edit/i }));
  const nameInput = screen.getByPlaceholderText(/Name/i);
  fireEvent.change(nameInput, { target: { value: 'New Name' } });
  fireEvent.click(screen.getByRole('button', { name: /Save Changes/i }));
  await waitFor(() => {
    expect(screen.queryByRole('button', { name: /Save Changes/i })).not.toBeInTheDocument();
  });
});

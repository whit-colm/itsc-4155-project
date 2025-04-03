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
  jest.spyOn(console, 'log').mockImplementation(() => {}); // Suppress console.log
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
    expect(screen.getByText('Name:')).toBeInTheDocument(); // Match exact text
    expect(screen.getByText('John Doe')).toBeInTheDocument(); // Match specific name
    expect(screen.getByText('Username:')).toBeInTheDocument(); // Match exact text
    expect(screen.getByText('johndoe')).toBeInTheDocument(); // Match specific username
    expect(screen.getByText('Email:')).toBeInTheDocument(); // Match exact text
    expect(screen.getByText('johndoe@example.com')).toBeInTheDocument(); // Match specific email
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

  // Use more specific queries to avoid ambiguity
  const nameInput = screen.getByPlaceholderText('Name');
  fireEvent.change(nameInput, { target: { value: 'New Name' } });

  const usernameInput = screen.getByPlaceholderText('Username');
  fireEvent.change(usernameInput, { target: { value: 'newusername' } });

  const emailInput = screen.getByPlaceholderText('Email');
  fireEvent.change(emailInput, { target: { value: 'newemail@example.com' } });

  fireEvent.click(screen.getByRole('button', { name: /Save Changes/i }));

  await waitFor(() => {
    expect(screen.queryByRole('button', { name: /Save Changes/i })).not.toBeInTheDocument();
  });
});

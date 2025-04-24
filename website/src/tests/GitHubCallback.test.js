import { render, screen, waitFor } from '@testing-library/react';
import GitHubCallback from '../pages/GitHubCallback';
import { useNavigate } from 'react-router-dom'; // Import useNavigate

// Mock useNavigate
const mockedNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'), // use actual for all non-hook parts
  useNavigate: () => mockedNavigate,
}));

const mockSetJwt = jest.fn();

beforeEach(() => {
  // Reset mocks
  mockSetJwt.mockClear();
  mockedNavigate.mockClear();
  global.fetch = jest.fn(); // Define fetch mock inside tests or specific beforeEach blocks

  // Mock window.location.search
  delete window.location;
  window.location = { search: '?code=mock-code&state=mock-state' }; // Only need search

  // Mock document.cookie
  Object.defineProperty(document, 'cookie', {
    writable: true,
    value: '', // Start with empty cookie
  });
});

afterEach(() => {
  jest.restoreAllMocks();
});

test('processes GitHub login, sets JWT, sets cookie, and navigates on success', async () => {
  // Mock successful fetch
  global.fetch.mockImplementationOnce(() =>
    Promise.resolve({
      ok: true, // Add ok status
      json: () => Promise.resolve({ token: 'mock-jwt-token' }),
    })
  );

  render(<GitHubCallback setJwt={mockSetJwt} />);

  // Check for loading text
  expect(screen.getByText(/Processing GitHub login.../i)).toBeInTheDocument();

  await waitFor(() => {
    // Verify fetch was called correctly
    expect(global.fetch).toHaveBeenCalledWith('/api/auth/github/callback?code=mock-code&state=mock-state');
    // Verify setJwt was called
    expect(mockSetJwt).toHaveBeenCalledWith('mock-jwt-token');
    // Verify cookie was set
    expect(document.cookie).toContain('jwt=mock-jwt-token');
    expect(document.cookie).toContain('path=/');
    expect(document.cookie).toContain('secure');
    expect(document.cookie).toContain('SameSite=Strict');
    // Verify navigation occurred
    expect(mockedNavigate).toHaveBeenCalledWith('/');
  });
});

test('navigates to login on fetch failure', async () => {
  // Mock failed fetch
  global.fetch.mockImplementationOnce(() =>
    Promise.resolve({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
    })
  );

  render(<GitHubCallback setJwt={mockSetJwt} />);

  // Check for loading text
  expect(screen.getByText(/Processing GitHub login.../i)).toBeInTheDocument();

  await waitFor(() => {
    // Verify fetch was called
    expect(global.fetch).toHaveBeenCalledWith('/api/auth/github/callback?code=mock-code&state=mock-state');
    // Verify setJwt was NOT called
    expect(mockSetJwt).not.toHaveBeenCalled();
    // Verify cookie was NOT set
    expect(document.cookie).not.toContain('jwt=');
    // Verify navigation to login occurred
    expect(mockedNavigate).toHaveBeenCalledWith('/login');
  });
});

test('navigates to login if token is missing in response', async () => {
  // Mock successful fetch but missing token
  global.fetch.mockImplementationOnce(() =>
    Promise.resolve({
      ok: true,
      json: () => Promise.resolve({}), // Empty object, no token
    })
  );

  render(<GitHubCallback setJwt={mockSetJwt} />);

  // Check for loading text
  expect(screen.getByText(/Processing GitHub login.../i)).toBeInTheDocument();

  await waitFor(() => {
    // Verify fetch was called
    expect(global.fetch).toHaveBeenCalledWith('/api/auth/github/callback?code=mock-code&state=mock-state');
    // Verify setJwt was NOT called
    expect(mockSetJwt).not.toHaveBeenCalled();
     // Verify cookie was NOT set
    expect(document.cookie).not.toContain('jwt=');
    // Verify navigation to login occurred
    expect(mockedNavigate).toHaveBeenCalledWith('/login');
  });
});

test('navigates to login if code or state is missing', async () => {
  // Mock window.location without code/state
  window.location = { search: '?code=mock-code' }; // Missing state

  render(<GitHubCallback setJwt={mockSetJwt} />);

  // Check for loading text
  expect(screen.getByText(/Processing GitHub login.../i)).toBeInTheDocument();

  // fetch shouldn't even be called in this case based on component logic
  await waitFor(() => {
      expect(global.fetch).not.toHaveBeenCalled();
      expect(mockedNavigate).toHaveBeenCalledWith('/login');
  });

  // Reset and test missing code
  mockedNavigate.mockClear();
  window.location = { search: '?state=mock-state' }; // Missing code
  render(<GitHubCallback setJwt={mockSetJwt} />);
   await waitFor(() => {
      expect(global.fetch).not.toHaveBeenCalled();
      expect(mockedNavigate).toHaveBeenCalledWith('/login');
  });
});
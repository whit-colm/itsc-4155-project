import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useParams } from 'react-router-dom';
import UserProfile from '../pages/UserProfile';

const mockJwt = 'mock-jwt-token';
const mockUserIdParam = 'user-from-url-id';
const mockAvatarId = 'mock-avatar-id-from-api';

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useParams: jest.fn(),
}));

beforeEach(() => {
  useParams.mockReturnValue({ userId: mockUserIdParam });

  Object.defineProperty(document, 'cookie', {
    writable: true,
    value: `jwt=${mockJwt}`,
  });

  global.fetch = jest.fn((url, options) => {
    if (url === `/api/user/${mockUserIdParam}`) {
      return Promise.resolve({
        ok: true,
        json: () =>
          Promise.resolve({
            id: mockUserIdParam,
            name: 'Mock User Name',
            username: 'mockuser#1234',
            pronouns: 'they/them',
            bref_avatar: mockAvatarId,
          }),
      });
    } else if (url === `/api/blob/${mockAvatarId}`) {
      global.URL.createObjectURL = jest.fn(() => 'mock-blob-url');
      return Promise.resolve({
        ok: true,
        blob: () => Promise.resolve(new Blob(['mock-image-content'], { type: 'image/png' })),
      });
    }
    return Promise.reject(new Error(`Unknown URL: ${url}`));
  });

  global.URL.revokeObjectURL = jest.fn();
});

afterEach(() => {
  jest.restoreAllMocks();
  if (global.URL.createObjectURL) delete global.URL.createObjectURL;
  if (global.URL.revokeObjectURL) delete global.URL.revokeObjectURL;
});

const renderWithRouter = (ui, route = `/user/${mockUserIdParam}`) => {
  return render(
    <MemoryRouter initialEntries={[route]}>
      <Routes>
        <Route path="/user/:userId" element={ui} />
      </Routes>
    </MemoryRouter>
  );
};

test('renders loading state initially', () => {
  renderWithRouter(<UserProfile jwt={mockJwt} />);
  expect(screen.getByText(/Loading user profile.../i)).toBeInTheDocument();
});

test('renders user profile details after loading', async () => {
  renderWithRouter(<UserProfile jwt={mockJwt} />);

  expect(await screen.findByRole('heading', { name: /Mock User Name/i })).toBeInTheDocument();
  expect(screen.getByText('mockuser#1234')).toBeInTheDocument();
  expect(screen.getByText('they/them')).toBeInTheDocument();

  expect(fetch).toHaveBeenCalledWith(`/api/user/${mockUserIdParam}`, expect.any(Object));
});

test('renders user avatar after loading', async () => {
  renderWithRouter(<UserProfile jwt={mockJwt} />);

  const avatarElement = await screen.findByAltText(/Mock User Name's avatar/i);
  expect(avatarElement).toBeInTheDocument();
  expect(avatarElement).toHaveAttribute('src', 'mock-blob-url');

  expect(fetch).toHaveBeenCalledWith(`/api/user/${mockUserIdParam}`, expect.any(Object));
  expect(fetch).toHaveBeenCalledWith(`/api/blob/${mockAvatarId}`, expect.any(Object));
  expect(global.URL.createObjectURL).toHaveBeenCalled();
});

test('renders fallback avatar if bref_avatar is missing', async () => {
  global.fetch.mockImplementation((url) => {
    if (url === `/api/user/${mockUserIdParam}`) {
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({
          id: mockUserIdParam,
          name: 'Mock User No Avatar',
          username: 'mockuser#1234',
          pronouns: 'they/them',
          bref_avatar: null,
        }),
      });
    }
    return Promise.reject(new Error(`Unexpected fetch in no-avatar test: ${url}`));
  });

  renderWithRouter(<UserProfile jwt={mockJwt} />);

  expect(await screen.findByRole('heading', { name: /Mock User No Avatar/i })).toBeInTheDocument();

  const avatarElement = screen.getByAltText(/Mock User No Avatar's avatar/i);
  expect(avatarElement).toBeInTheDocument();
  expect(avatarElement).toHaveAttribute('src', '/logo192.png');

  expect(fetch).toHaveBeenCalledTimes(1);
  expect(fetch).toHaveBeenCalledWith(`/api/user/${mockUserIdParam}`, expect.any(Object));
  expect(global.URL.createObjectURL).not.toHaveBeenCalled();
});

test('displays error message on fetch failure', async () => {
  global.fetch.mockImplementationOnce(() =>
    Promise.resolve({
      ok: false,
      status: 404,
      json: () => Promise.resolve({ summary: 'User not found in mock' }),
    })
  );

  renderWithRouter(<UserProfile jwt={mockJwt} />);

  await waitFor(() => {
    expect(screen.getByText(/Error: User not found in mock/i)).toBeInTheDocument();
  });

  expect(screen.queryByRole('heading', { name: /Mock User Name/i })).not.toBeInTheDocument();
});
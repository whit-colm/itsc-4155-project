import { render, screen, waitFor } from '@testing-library/react';
import UserProfile from '../pages/UserProfile';

const mockJwt = 'mock-jwt-token';
const mockUserId = 'mock-user-id';

beforeEach(() => {
  global.fetch = jest.fn((url) => {
    if (url.includes(`/api/user/${mockUserId}`)) {
      return Promise.resolve({
        json: () =>
          Promise.resolve({
            displayName: 'Mock User',
            username: 'mockuser',
            pronouns: 'they/them',
            bref_avatar: 'mock-avatar-id',
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

test('renders user profile details', async () => {
  render(<UserProfile jwt={mockJwt} />);

  await waitFor(() => {
    expect(screen.getByText('Mock User')).toBeInTheDocument();
    expect(screen.getByText('mockuser')).toBeInTheDocument();
    expect(screen.getByText('they/them')).toBeInTheDocument();
  });
});

test('renders user avatar', async () => {
  render(<UserProfile jwt={mockJwt} />);

  await waitFor(() => {
    const avatarElement = screen.getByAltText(/Mock User's avatar/i);
    expect(avatarElement).toBeInTheDocument();
  });
});
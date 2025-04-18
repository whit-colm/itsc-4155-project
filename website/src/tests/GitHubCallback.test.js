import { render, screen, waitFor } from '@testing-library/react';
import GitHubCallback from '../pages/GitHubCallback';

const mockSetJwt = jest.fn();

beforeEach(() => {
  global.fetch = jest.fn(() =>
    Promise.resolve({
      json: () => Promise.resolve({ token: 'mock-jwt-token' }),
    })
  );
  delete window.location;
  window.location = { search: '?code=mock-code&state=mock-state', href: '' };
});

afterEach(() => {
  jest.restoreAllMocks();
});

test('processes GitHub login and sets JWT', async () => {
  render(<GitHubCallback setJwt={mockSetJwt} />);

  await waitFor(() => {
    expect(mockSetJwt).toHaveBeenCalledWith('mock-jwt-token');
    expect(window.location.href).toBe('/');
  });
});
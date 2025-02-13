import { render, screen } from '@testing-library/react';
import Recommendations from '../pages/Recommendations';

test('renders Book Recommendations heading', () => {
  render(<Recommendations />);
  const headingElement = screen.getByText(/Book Recommendations/i);
  expect(headingElement).toBeInTheDocument();
});

test('renders list of books', () => {
  render(<Recommendations />);
  const bookElements = screen.getAllByRole('listitem');
  expect(bookElements.length).toBeGreaterThan(0);
});

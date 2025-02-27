import { render, screen } from '@testing-library/react';
import Home from '../pages/Home';

test('renders Welcome to Jaws heading', () => {
  render(<Home />);
  const headingElement = screen.getByText(/Welcome to Jaws/i);
  expect(headingElement).toBeInTheDocument();
});

test('renders personalized book recommendation text', () => {
  render(<Home />);
  const textElement = screen.getByText(/Your personalized book recommendation system./i);
  expect(textElement).toBeInTheDocument();
});

test('renders images', () => {
  render(<Home />);
  const imageElements = screen.getAllByRole('img');
  expect(imageElements.length).toBe(2);
});

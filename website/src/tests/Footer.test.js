import { render, screen } from '@testing-library/react';
import Footer from '../components/Footer';

test('renders Footer component', () => {
  render(<Footer />);
  const footerElement = screen.getByText(/Follow us on social media:/i);
  expect(footerElement).toBeInTheDocument();
});

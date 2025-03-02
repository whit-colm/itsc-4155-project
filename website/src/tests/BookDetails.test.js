import { render, screen } from '@testing-library/react';
import BookDetails from '../pages/BookDetails';

test('renders Book Details heading', async () => {
  render(<BookDetails uuid="01953a93-21e7-73da-8a27-fc22aa66a95e" />);

  const headingElement = await screen.findByText(/Loading.../i);
  expect(headingElement).toBeInTheDocument();
});

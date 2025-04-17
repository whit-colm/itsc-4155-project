import React, { useState, useEffect } from 'react';
import '../styles/BookDetails.css';
import Comments from './Comments';
import Reviews from './Reviews';

function BookDetails({ uuid, jwt }) {
	const [book, setBook] = useState(null);
	const [coverUrl, setCoverUrl] = useState(null);
	const [error, setError] = useState(null);

	const fetchBook = async (retries = 3) => {
		try {
			const response = await fetch(`/api/books/${uuid}`, {
				method: 'GET',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${jwt}`, // Use jwt for authorization
				},
			});

			if (!response.ok) {
				throw new Error(`Failed to fetch book details: ${response.statusText}`);
			}

			const data = await response.json();
			setBook(data);

			if (data.bref_cover) {
				const coverResponse = await fetch(`/api/blob/${data.bref_cover}`);
				if (!coverResponse.ok) {
					throw new Error(`Failed to fetch cover image: ${coverResponse.statusText}`);
				}
				const coverBlob = await coverResponse.blob();
				setCoverUrl(URL.createObjectURL(coverBlob));
			}
		} catch (err) {
			if (retries > 0) {
				console.warn(`Retrying fetchBook... (${3 - retries + 1})`);
				fetchBook(retries - 1);
			} else {
				setError('Failed to load book details. Please try again later.');
			}
		}
	};

	useEffect(() => {
		fetchBook();
	}, [uuid, jwt]);

	if (error) {
		return <div className="error-message">Error: {error}</div>;
	}

	if (!book) {
		return <div>Loading...</div>;
	}

	return (
		<div className="book-details-container">
			<h1>{book.title}</h1>
			{coverUrl && <img src={coverUrl} alt={`${book.title} cover`} className="book-cover" />}
			<p><strong>Author:</strong> {book.author}</p>
			<p><strong>Published:</strong> {book.published}</p>
			<p><strong>ISBN-10:</strong> {book.isbns.find(isbn => isbn.type === 'isbn10')?.value || 'N/A'}</p>
			<p><strong>ISBN-13:</strong> {book.isbns.find(isbn => isbn.type === 'isbn13')?.value || 'N/A'}</p>
			<Comments bookId={uuid} jwt={jwt} />
			<Reviews bookId={uuid} jwt={jwt} />
		</div>
	);
}

export default BookDetails;
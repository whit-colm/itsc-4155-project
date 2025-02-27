import React, { useState } from 'react';
import '../styles/CreateBook.css';

function CreateBook() {
  const [title, setTitle] = useState('');
  const [author, setAuthor] = useState('');
  const [published, setPublished] = useState('');
  const [isbn10, setIsbn10] = useState('');
  const [isbn13, setIsbn13] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    const newBook = {
      title,
      author,
      published,
      isbns: [
        { type: 'isbn10', value: isbn10 },
        { type: 'isbn13', value: isbn13 }
      ]
    };
    const response = await fetch('/api/books/new', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(newBook)
    });
    const data = await response.json();
    console.log('Book created:', data);
  };

  return (
    <div className="create-book-container">
      <h1>Create Book</h1>
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Title"
          required
        />
        <input
          type="text"
          value={author}
          onChange={(e) => setAuthor(e.target.value)}
          placeholder="Author"
          required
        />
        <input
          type="date"
          value={published}
          onChange={(e) => setPublished(e.target.value)}
          placeholder="Published Date"
          required
        />
        <input
          type="text"
          value={isbn10}
          onChange={(e) => setIsbn10(e.target.value)}
          placeholder="ISBN-10"
          required
        />
        <input
          type="text"
          value={isbn13}
          onChange={(e) => setIsbn13(e.target.value)}
          placeholder="ISBN-13"
          required
        />
        <button type="submit">Create Book</button>
      </form>
    </div>
  );
}

export default CreateBook;

import React, { useState } from 'react';
import '../styles/CreateBook.css';

function CreateBook() {
  const [title, setTitle] = useState('');
  const [author, setAuthor] = useState('');
  const [published, setPublished] = useState('');
  const [isbns, setIsbns] = useState([{ type: 'isbn10', value: '' }]);
  const [errors, setErrors] = useState({});

  const isbn10Regex = /^(?:\d[\ |-]?){9}[\d|X]$/;
  const isbn13Regex = /^(?:\d[\ |-]?){13}$/;

  const handleAddIsbn = (type) => {
    if (isbns.some(isbn => isbn.type === type)) {
      return;
    }
    setIsbns([...isbns, { type, value: '' }]);
  };

  const handleRemoveIsbn = (index) => {
    const newIsbns = isbns.filter((_, i) => i !== index);
    setIsbns(newIsbns);
  };

  const handleIsbnChange = (index, field, value) => {
    const newIsbns = isbns.map((isbn, i) => 
      i === index ? { ...isbn, [field]: value } : isbn
    );
    setIsbns(newIsbns);
  };

  const validateIsbns = () => {
    const newErrors = {};
    isbns.forEach((isbn, index) => {
      if (isbn.type === 'isbn10' && !isbn10Regex.test(isbn.value)) {
        newErrors[index] = 'Invalid ISBN-10 format';
      } else if (isbn.type === 'isbn13' && !isbn13Regex.test(isbn.value)) {
        newErrors[index] = 'Invalid ISBN-13 format';
      }
    });
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validateIsbns()) {
      return;
    }
    const newBook = {
      title,
      author,
      published,
      isbns
    };
    const jwt = document.cookie
      .split('; ')
      .find((row) => row.startsWith('jwt='))
      ?.split('=')[1]; // Retrieve JWT from cookie
    const response = await fetch('/api/books/new', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${jwt}`, // Include JWT for authorization
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
        {isbns.map((isbn, index) => (
          <div key={index} className="isbn-input">
            <select
              value={isbn.type}
              onChange={(e) => handleIsbnChange(index, 'type', e.target.value)}
              disabled
            >
              <option value="isbn10">ISBN-10</option>
              <option value="isbn13">ISBN-13</option>
            </select>
            <input
              type="text"
              value={isbn.value}
              onChange={(e) => handleIsbnChange(index, 'value', e.target.value)}
              placeholder="ISBN"
              required
            />
            <button type="button" onClick={() => handleRemoveIsbn(index)}>X</button>
            {errors[index] && <span className="error">{errors[index]}</span>}
          </div>
        ))}
        {!isbns.some(isbn => isbn.type === 'isbn10') && (
          <button type="button" onClick={() => handleAddIsbn('isbn10')}>Add ISBN-10</button>
        )}
        {!isbns.some(isbn => isbn.type === 'isbn13') && (
          <button type="button" onClick={() => handleAddIsbn('isbn13')}>Add ISBN-13</button>
        )}
        <button type="submit">Create Book</button>
      </form>
    </div>
  );
}

export default CreateBook;

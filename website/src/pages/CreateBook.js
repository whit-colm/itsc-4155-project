import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom'; // Import useNavigate
import '../styles/CreateBook.css';

function CreateBook() {
  const [title, setTitle] = useState('');
  const [author, setAuthor] = useState(''); // Keep for UI, but won't be sent directly
  const [published, setPublished] = useState(''); // Expects YYYY-MM-DD
  const [isbns, setIsbns] = useState([{ type: 'isbn10', value: '' }]);
  const [errors, setErrors] = useState({});
  const [image, setImage] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');
  const navigate = useNavigate();

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
    // Clear validation error for this specific ISBN when its value changes
    if (errors[index]) {
        const newErrors = { ...errors };
        delete newErrors[index];
        setErrors(newErrors);
    }
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

  const handleImageChange = (e) => {
    setImage(e.target.files[0]); // Set the selected image file
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitError('');
    if (!validateIsbns()) {
      return;
    }

    setIsSubmitting(true);

    // Prepare book data matching Go model.Book structure
    const newBookPayload = {
      title,
      published: published, // Already in YYYY-MM-DD format from input type="date"
      isbns: isbns
        .filter(isbn => isbn.value.trim() !== '') // Filter out empty ISBNs
        .map(isbn => ({
            type: isbn.type,
            // Clean ISBN value (remove hyphens and spaces) before sending
            value: isbn.value.replace(/[\s-]/g, '')
        })),
    };

    // --- JWT Retrieval ---
    const jwt = document.cookie
      .split('; ')
      .find((row) => row.startsWith('jwt='))
      ?.split('=')[1];

    if (!jwt) {
        setSubmitError('Authentication error. Please log in.');
        setIsSubmitting(false);
        return;
    }

    // --- Request Configuration ---
    let requestOptions = {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${jwt}`,
        },
    };

    if (image) {
        console.warn("Attempting image upload, but backend /api/books/new expects JSON.");
        setSubmitError("Image upload is not supported by this endpoint. Book will be created without cover.");
        requestOptions.headers['Content-Type'] = 'application/json';
        requestOptions.body = JSON.stringify(newBookPayload);
    } else {
        requestOptions.headers['Content-Type'] = 'application/json';
        requestOptions.body = JSON.stringify(newBookPayload);
    }

    // --- API Call ---
    try {
      const response = await fetch('/api/books/new', requestOptions);

      if (response.ok) {
        const data = await response.json(); // Expects model.Book response
        console.log('Book created:', data);
        navigate(`/books/${data.id}`);
      } else {
        const errorData = await response.json();
        console.error('Failed to create book:', errorData.summary || response.statusText, errorData.details);
        setSubmitError(errorData.summary || `Error: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Unexpected error:', error);
      setSubmitError('An unexpected error occurred. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="create-book-container">
      <h1>Create Book</h1>
      {submitError && <p className="error submit-error">{submitError}</p>}
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Title"
          required
          disabled={isSubmitting}
        />

        <input
          type="text"
          value={author}
          onChange={(e) => setAuthor(e.target.value)}
          placeholder="Author Name" // Clarify placeholder
          disabled={isSubmitting}
        />

        <input
          type="date"
          value={published}
          onChange={(e) => setPublished(e.target.value)}
          placeholder="Published Date"
          required
          disabled={isSubmitting}
        />

        <label className="isbn-label">ISBNs:</label>
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
              placeholder={`Enter ${isbn.type.toUpperCase()}`}
              required={isbn.value.trim() !== '' || isbns.length === 1}
              disabled={isSubmitting}
            />
            {isbns.length > 1 && (
                 <button
                    type="button"
                    className="remove-isbn-button"
                    onClick={() => handleRemoveIsbn(index)}
                    disabled={isSubmitting}
                 >
                    X
                 </button>
            )}
            {errors[index] && <span className="error isbn-error">{errors[index]}</span>}
          </div>
        ))}
        <div className="isbn-buttons">
            {!isbns.some(isbn => isbn.type === 'isbn10') && (
              <button
                type="button"
                className="add-isbn-button"
                onClick={() => handleAddIsbn('isbn10')}
                disabled={isSubmitting}
              >
                Add ISBN-10
              </button>
            )}
            {!isbns.some(isbn => isbn.type === 'isbn13') && (
              <button
                type="button"
                className="add-isbn-button"
                onClick={() => handleAddIsbn('isbn13')}
                disabled={isSubmitting}
              >
                Add ISBN-13
              </button>
            )}
        </div>

        <label htmlFor="image-upload-input" className="image-upload-label">Book Cover (Optional - Upload Not Supported):</label>
        <input
          id="image-upload-input"
          type="file"
          accept="image/*"
          onChange={handleImageChange}
          className="image-upload"
          disabled={isSubmitting}
        />
        {image && <span className="image-filename">Selected: {image.name}</span>}

        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Creating...' : 'Create Book'}
        </button>
      </form>
    </div>
  );
}

export default CreateBook;

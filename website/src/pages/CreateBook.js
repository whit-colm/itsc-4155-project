import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom'; // Import useNavigate
import '../styles/CreateBook.css';

function CreateBook() {
  const [title, setTitle] = useState('');
  const [author, setAuthor] = useState('');
  const [published, setPublished] = useState('');
  const [isbns, setIsbns] = useState([{ type: 'isbn10', value: '' }]);
  const [errors, setErrors] = useState({});
  const [image, setImage] = useState(null); // State for the book image
  const [isSubmitting, setIsSubmitting] = useState(false); // Add submitting state
  const [submitError, setSubmitError] = useState(''); // State for submission errors
  const navigate = useNavigate(); // Initialize useNavigate

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
    setSubmitError(''); // Clear previous submission errors
    if (!validateIsbns()) {
      return;
    }

    setIsSubmitting(true); // Set submitting state to true

    // Ensure published date is in YYYY-MM-DD format or adjust if backend needs ISO
    const formattedPublishedDate = published; // Assuming backend accepts YYYY-MM-DD

    // Prepare book data matching Go model structure
    const newBook = {
      title,
      published: formattedPublishedDate,
      isbns: isbns.map(isbn => ({ type: isbn.type, value: isbn.value.replace(/[\s-]/g, '') })) // Clean ISBN values
    };

    const jwt = document.cookie
      .split('; ')
      .find((row) => row.startsWith('jwt='))
      ?.split('=')[1]; // Retrieve JWT from cookie

    if (!jwt) {
        setSubmitError('Authentication error. Please log in.');
        setIsSubmitting(false);
        return;
    }

    let requestOptions = {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${jwt}`,
        },
    };

    if (image) {
        // Use FormData when image is present
        const formData = new FormData();
        formData.append('book', JSON.stringify(newBook));
        formData.append('image', image);
        requestOptions.body = formData;
        console.warn("Attempting image upload with FormData. Backend might not support this.");
        setSubmitError("Image upload is not currently supported by the backend."); // Inform user proactively
    } else {
        // Use application/json when no image is present
        requestOptions.headers['Content-Type'] = 'application/json';
        requestOptions.body = JSON.stringify(newBook);
    }

    try {
      // Use the configured requestOptions
      const response = await fetch('/api/books/new', requestOptions);

      if (response.ok) {
        const data = await response.json();
        console.log('Book created:', data);
        navigate(`/books/${data.id}`); // Use 'id' from the response
      } else {
        const errorData = await response.json();
        console.error('Failed to create book:', errorData.summary || response.statusText, errorData.details);
        setSubmitError(errorData.summary || `Error: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Unexpected error:', error);
      setSubmitError('An unexpected error occurred. Please try again.');
    } finally {
      setIsSubmitting(false); // Reset submitting state regardless of outcome
    }
  };

  return (
    <div className="create-book-container">
      <h1>Create Book</h1>
      {/* Display submission error */}
      {submitError && <p className="error submit-error">{submitError}</p>}
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Title"
          required
          disabled={isSubmitting} // Disable when submitting
        />
        <input
          type="text"
          value={author}
          onChange={(e) => setAuthor(e.target.value)}
          placeholder="Author"
          required
          disabled={isSubmitting} // Disable when submitting
        />
        <input
          type="date"
          value={published}
          onChange={(e) => setPublished(e.target.value)}
          placeholder="Published Date"
          required
          disabled={isSubmitting} // Disable when submitting
        />

        <label className="isbn-label">ISBNs:</label> {/* Add label for clarity */}
        {isbns.map((isbn, index) => (
          <div key={index} className="isbn-input">
            <select
              value={isbn.type}
              onChange={(e) => handleIsbnChange(index, 'type', e.target.value)}
              disabled // Keep type selection disabled after adding
            >
              <option value="isbn10">ISBN-10</option>
              <option value="isbn13">ISBN-13</option>
            </select>
            <input
              type="text"
              value={isbn.value}
              onChange={(e) => handleIsbnChange(index, 'value', e.target.value)}
              placeholder={`Enter ${isbn.type.toUpperCase()}`}
              required
              disabled={isSubmitting} // Disable when submitting
            />
            {isbns.length > 1 && (
                 <button
                    type="button"
                    className="remove-isbn-button"
                    onClick={() => handleRemoveIsbn(index)}
                    disabled={isSubmitting} // Disable when submitting
                 >
                    X
                 </button>
            )}
            {errors[index] && <span className="error isbn-error">{errors[index]}</span>}
          </div>
        ))}
        <div className="isbn-buttons"> {/* Group add buttons */}
            {!isbns.some(isbn => isbn.type === 'isbn10') && (
              <button
                type="button"
                className="add-isbn-button"
                onClick={() => handleAddIsbn('isbn10')}
                disabled={isSubmitting} // Disable when submitting
              >
                Add ISBN-10
              </button>
            )}
            {!isbns.some(isbn => isbn.type === 'isbn13') && (
              <button
                type="button"
                className="add-isbn-button"
                onClick={() => handleAddIsbn('isbn13')}
                disabled={isSubmitting} // Disable when submitting
              >
                Add ISBN-13
              </button>
            )}
        </div>

        <label htmlFor="image-upload-input" className="image-upload-label">Book Cover (Optional):</label>
        <input
          id="image-upload-input"
          type="file"
          accept="image/*"
          onChange={handleImageChange}
          className="image-upload"
          disabled={isSubmitting} // Disable when submitting
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

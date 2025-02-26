import React, { useState } from 'react';
import '../styles/CreateAccount.css';

function CreateAccount() {
  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [email, setEmail] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    // Implement account creation logic here
    console.log('Account created:', { name, email, password });
  };

  return (
    <div className="create-account-container">
      <h1>Create Account</h1>
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Name"
          required
        />
        <input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="Email"
          required
        />
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Password"
          required
        />
        <button type="submit">Create Account</button>
      </form>
    </div>
  );
}

export default CreateAccount;

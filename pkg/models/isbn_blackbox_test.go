package models

import (
	"encoding/json"
	"testing"
)

func TestIsbnVersionMethods(t *testing.T) {
	// Test String() method
	if ISBN10.String() != "isbn10" {
		t.Errorf("Expected ISBN10.String() to return 'isbn10', got: %s", ISBN10.String())
	}
	if ISBN13.String() != "isbn13" {
		t.Errorf("Expected ISBN13.String() to return 'isbn13', got: %s", ISBN13.String())
	}

	// Test Len() method
	if ISBN10.Len() != 10 {
		t.Errorf("Expected ISBN10.Len() to return 10, got: %d", ISBN10.Len())
	}
	if ISBN13.Len() != 13 {
		t.Errorf("Expected ISBN13.Len() to return 13, got: %d", ISBN13.Len())
	}

	// Test regex() method returns non-nil values
	if ISBN10.regex() == nil {
		t.Error("Expected non-nil regex for ISBN10")
	}
	if ISBN13.regex() == nil {
		t.Error("Expected non-nil regex for ISBN13")
	}
}

func TestNewISBN_ValidISBN10(t *testing.T) {
	// A known valid ISBN10 (hyphens are allowed)
	valid := "0-306-40615-2"
	isbn, err := NewISBN(valid, ISBN10)
	if err != nil {
		t.Fatalf("Expected valid ISBN10, got error: %v", err)
	}
	if !isbn.Check() {
		t.Errorf("Check() returned false for valid ISBN10: %s", valid)
	}
}

func TestNewISBN_InvalidISBN10(t *testing.T) {
	// An ISBN10 with an incorrect check digit
	invalid := "0306406153"
	_, err := NewISBN(invalid, ISBN10)
	if err == nil {
		t.Errorf("Expected error for invalid ISBN10: %s", invalid)
	}
}

func TestNewISBN_ValidISBN10WithX(t *testing.T) {
	// A valid ISBN10 where the last digit is "X" (representing 10)
	valid := "007462542X"
	isbn, err := NewISBN(valid, ISBN10)
	if err != nil {
		t.Fatalf("Expected valid ISBN10 with X, got error: %v", err)
	}
	if !isbn.Check() {
		t.Errorf("Check() returned false for valid ISBN10 with X: %s", valid)
	}
}

func TestNewISBN_ValidISBN13(t *testing.T) {
	// A known valid ISBN13
	valid := "9780306406157"
	isbn, err := NewISBN(valid, ISBN13)
	if err != nil {
		t.Fatalf("Expected valid ISBN13, got error: %v", err)
	}
	if !isbn.Check() {
		t.Errorf("Check() returned false for valid ISBN13: %s", valid)
	}
}

func TestNewISBN_InvalidISBN13(t *testing.T) {
	// An ISBN13 with an incorrect check digit
	invalid := "9780306406158"
	_, err := NewISBN(invalid, ISBN13)
	if err == nil {
		t.Errorf("Expected error for invalid ISBN13: %s", invalid)
	}
}

func TestMarshalJSON_ValidISBN(t *testing.T) {
	// Create a valid ISBN10 and marshal it to JSON.
	// The JSON output should include the type "isbn10" and a cleaned-up value (digits only).
	valid := "0-306-40615-2"
	isbn, err := NewISBN(valid, ISBN10)
	if err != nil {
		t.Fatalf("Error creating ISBN: %v", err)
	}
	data, err := json.Marshal(isbn)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	// Unmarshal the JSON into a temporary structure to verify its fields.
	var out struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}
	if out.Type != "isbn10" {
		t.Errorf("Expected type 'isbn10', got: %s", out.Type)
	}
	// The cleaned value should contain only digits (and "X" replaced as needed).
	if out.Value != "0306406152" {
		t.Errorf("Expected value '0306406152', got: %s", out.Value)
	}
}

func TestMarshalJSON_InvalidISBN(t *testing.T) {
	// Manually construct an ISBN that fails the Check() method.
	isbn := ISBN{"invalid-isbn", ISBN10}
	_, err := json.Marshal(isbn)
	if err == nil {
		t.Error("Expected error when marshalling an invalid ISBN, but got none")
	}
}

func TestUnmarshalJSON_ValidISBN10(t *testing.T) {
	// Test unmarshalling valid JSON for an ISBN10.
	jsonData := []byte(`{"type": "isbn10", "value": "0-306-40615-2"}`)
	var isbn ISBN
	if err := json.Unmarshal(jsonData, &isbn); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if !isbn.Check() {
		t.Errorf("Unmarshalled ISBN did not pass Check(): %v", isbn)
	}
}

func TestUnmarshalJSON_InvalidType(t *testing.T) {
	// Test unmarshalling JSON with an unrecognized ISBN type.
	jsonData := []byte(`{"type": "isbn15", "value": "0-306-40615-2"}`)
	var isbn ISBN
	err := json.Unmarshal(jsonData, &isbn)
	if err == nil {
		t.Error("Expected error for invalid ISBN type in JSON, but got none")
	}
}

func TestUnmarshalJSON_InvalidValue(t *testing.T) {
	// Test unmarshalling JSON where the ISBN value fails the check.
	jsonData := []byte(`{"type": "isbn10", "value": "0306406153"}`)
	var isbn ISBN
	err := json.Unmarshal(jsonData, &isbn)
	if err == nil {
		t.Error("Expected error for invalid ISBN value in JSON, but got none")
	}
}

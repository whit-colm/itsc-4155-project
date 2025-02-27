package model

import (
	"encoding/json"
	"testing"
)

func TestIsbnVersionMethods(t *testing.T) {
	// Verify string representations
	if ISBN10.String() != "isbn10" {
		t.Errorf("Expected ISBN10.String() to return 'isbn10', got: %s", ISBN10.String())
	}
	if ISBN13.String() != "isbn13" {
		t.Errorf("Expected ISBN13.String() to return 'isbn13', got: %s", ISBN13.String())
	}

	// Verify lengths
	if ISBN10.Len() != 10 {
		t.Errorf("Expected ISBN10.Len() to return 10, got: %d", ISBN10.Len())
	}
	if ISBN13.Len() != 13 {
		t.Errorf("Expected ISBN13.Len() to return 13, got: %d", ISBN13.Len())
	}

	// Verify regex functions
	if ISBN10.regex() == nil {
		t.Error("Expected non-nil regex for ISBN10")
	}
	if ISBN13.regex() == nil {
		t.Error("Expected non-nil regex for ISBN13")
	}
}

func TestNewISBN_ValidISBN10(t *testing.T) {
	valid := "0-306-40615-2"
	isbn, err := NewISBN(valid, ISBN10)
	if err != nil {
		t.Fatalf("Expected valid ISBN10, got error: %v", err)
	}
	if !isbn.Check() {
		t.Errorf("Check() returned false for valid ISBN10: %s", valid)
	}
	if isbn.Version() != ISBN10 {
		t.Errorf("Expected ISBN version to be ISBN10, got: %s", isbn.Version().String())
	}
}

func TestNewISBN_ValidISBN13(t *testing.T) {
	valid := "9780306406157"
	isbn, err := NewISBN(valid, ISBN13)
	if err != nil {
		t.Fatalf("Expected valid ISBN13, got error: %v", err)
	}
	if !isbn.Check() {
		t.Errorf("Check() returned false for valid ISBN13: %s", valid)
	}
	if isbn.Version() != ISBN13 {
		t.Errorf("Expected ISBN version to be ISBN13, got: %s", isbn.Version().String())
	}
}

func TestNewISBN_InferVersion(t *testing.T) {
	// Without providing an explicit version, the function should infer based on the cleaned input length.
	validISBN10 := "0-306-40615-2"
	isbn10, err := NewISBN(validISBN10)
	if err != nil {
		t.Fatalf("Expected to infer valid ISBN10, got error: %v", err)
	}
	if isbn10.Version() != ISBN10 {
		t.Errorf("Expected inferred ISBN version to be ISBN10, got: %s", isbn10.Version().String())
	}

	validISBN13 := "9780306406157"
	isbn13, err := NewISBN(validISBN13)
	if err != nil {
		t.Fatalf("Expected to infer valid ISBN13, got error: %v", err)
	}
	if isbn13.Version() != ISBN13 {
		t.Errorf("Expected inferred ISBN version to be ISBN13, got: %s", isbn13.Version().String())
	}
}

func TestNewISBN_Invalid(t *testing.T) {
	// Test with an input that does not have 10 or 13 digits after cleaning.
	invalid := "123456789" // Only 9 digits
	_, err := NewISBN(invalid)
	if err == nil {
		t.Errorf("Expected error for invalid ISBN input: %s", invalid)
	}
}

func TestMustNewISBN_Valid(t *testing.T) {
	valid := "0-306-40615-2"
	isbn := MustNewISBN(valid, ISBN10)
	if !isbn.Check() {
		t.Errorf("MustNewISBN produced an ISBN that did not pass Check() for valid input: %s", valid)
	}
}

func TestMustNewISBN_Invalid(t *testing.T) {
	invalid := "123456789"
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid ISBN in MustNewISBN, but no panic occurred")
		}
	}()
	// This call should panic due to invalid input.
	MustNewISBN(invalid)
}

func TestMarshalJSON_ValidISBN(t *testing.T) {
	valid := "0-306-40615-2"
	isbn, err := NewISBN(valid, ISBN10)
	if err != nil {
		t.Fatalf("Error creating ISBN: %v", err)
	}
	data, err := json.Marshal(isbn)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}
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
	// The cleaned-up value should contain only digits (and X replaced if needed).
	if out.Value != "0306406152" {
		t.Errorf("Expected value '0306406152', got: %s", out.Value)
	}
}

func TestMarshalJSON_InvalidISBN(t *testing.T) {
	// Construct an ISBN that fails Check().
	isbn := ISBN{"invalid-isbn", ISBN10}
	_, err := json.Marshal(isbn)
	if err == nil {
		t.Error("Expected error when marshalling an invalid ISBN, but got none")
	}
}

func TestUnmarshalJSON_ValidISBN10(t *testing.T) {
	jsonData := []byte(`{"type": "isbn10", "value": "0-306-40615-2"}`)
	var isbn ISBN
	if err := json.Unmarshal(jsonData, &isbn); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if !isbn.Check() {
		t.Errorf("Unmarshalled ISBN did not pass Check(): %v", isbn)
	}
	if isbn.Version() != ISBN10 {
		t.Errorf("Expected version isbn10, got: %s", isbn.Version().String())
	}
}

func TestUnmarshalJSON_InvalidType(t *testing.T) {
	jsonData := []byte(`{"type": "isbn15", "value": "0-306-40615-2"}`)
	var isbn ISBN
	err := json.Unmarshal(jsonData, &isbn)
	if err == nil {
		t.Error("Expected error for invalid ISBN type in JSON, but got none")
	}
}

func TestUnmarshalJSON_InvalidValue(t *testing.T) {
	jsonData := []byte(`{"type": "isbn10", "value": "0306406153"}`)
	var isbn ISBN
	err := json.Unmarshal(jsonData, &isbn)
	if err == nil {
		t.Error("Expected error for invalid ISBN value in JSON, but got none")
	}
}

// --- Fuzzing Tests ---

// FuzzNewISBN attempts to create ISBNs from arbitrary strings. The purpose is to
// catch any panics or unexpected behavior when inferring the version.
func FuzzNewISBN(f *testing.F) {
	// Seed with some valid and invalid ISBN strings.
	seeds := []string{
		"0-306-40615-2",
		"9780306406157",
		"1234567890",
		"invalid",
		"9780306406158", // invalid ISBN13
		"007462542X",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data string) {
		// We ignore the result; we're just ensuring no panics occur.
		_, _ = NewISBN(data)
	})
}

// FuzzCheck fuzzes the Check() method for various ISBN inputs.
func FuzzCheck(f *testing.F) {
	seeds := []string{
		"0-306-40615-2",
		"9780306406157",
		"007462542X",
		"123456789X",
		"invalid",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data string) {
		// Attempt to create an ISBN; if successful, run Check().
		isbn, _ := NewISBN(data)
		_ = isbn.Check()
	})
}

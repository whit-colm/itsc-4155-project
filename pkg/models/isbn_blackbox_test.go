package models_test

import (
	"testing"

	m "github.com/whit-colm/itsc-4155-project/pkg/models"
)

func TestNewISBN10(t *testing.T) {
	i10 := "0-19-852663-6"
	isbnZero := m.ISBN10{}
	i, err := m.NewISBN10(i10)
	if err != nil {
		t.Errorf("error for `%v`: %s", i10, err)
	} else if i == isbnZero {
		t.Errorf("ISBN result was empty")
	}

	i10 = "0-590-48348-X"
	i, err = m.NewISBN10(i10)
	if err != nil {
		t.Errorf("error for `%v`: %s", i10, err)
	} else if i == isbnZero {
		t.Errorf("ISBN result was empty")
	}

	i10 = "0590483489"
	i, err = m.NewISBN10(i10)
	if err == nil {
		t.Errorf("undue validation for `%v`", i10)
	} else if i != isbnZero {
		t.Errorf("ISBN was not empty and should have been.")
	}
}

func TestISBN10JsonUnmarshal(t *testing.T) {
	have := `{"type":"isbn10","value":"0198526636"}`
	want, _ := m.NewISBN10("0198526636")
	test := m.ISBN10{}

	err := test.UnmarshalJSON([]byte(have))
	if err != nil {
		t.Errorf("unable to unmarshal: %s", err)
	}
	if test.String() != want.String() {
		t.Errorf("not equal: want %v; have %v", want, test)
	}

	have = `{"type":"isbn10","value":"0-590-48348-X"}`
	want, _ = m.NewISBN10("059048348X")
	test = m.ISBN10{}

	err = test.UnmarshalJSON([]byte(have))
	if err != nil {
		t.Errorf("unable to unmarshal: %s", err)
	}
	if test.String() != want.String() {
		t.Errorf("not equal: want %v; have %v", want, test)
	}
}

/// Test ISBN13 ///

func TestNewISBN13(t *testing.T) {
	i13 := "9780061122415"
	isbnZero := m.ISBN13{}
	i, err := m.NewISBN13(i13)
	if err != nil {
		t.Errorf("error for `%v`: %s", i13, err)
	} else if i == isbnZero {
		t.Errorf("ISBN result was empty")
	}

	i13 = "978-0141439600"
	i, err = m.NewISBN13(i13)
	if err != nil {
		t.Errorf("error for `%v`: %s", i13, err)
	} else if i == isbnZero {
		t.Errorf("ISBN result was empty")
	}
}

func TestISBN13JsonUnmarshal(t *testing.T) {
	have := `{"type":"isbn13","value":"978-0141439600"}`
	want, _ := m.NewISBN13("9780141439600")
	test := m.ISBN13{}

	err := test.UnmarshalJSON([]byte(have))
	if err != nil {
		t.Errorf("unable to unmarshal: %s", err)
	}
	if test.String() != want.String() {
		t.Errorf("not equal: want %v; have %v", want, test)
	}
}

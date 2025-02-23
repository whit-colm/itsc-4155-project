package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	isbn10re *regexp.Regexp
	isbn13re *regexp.Regexp
)

func init() {
	isbn10re = regexp.MustCompile(`\d|[Xx]$`)
	isbn13re = regexp.MustCompile(`\d`)
}

type ISBNFormatter interface {
	Check() bool
	fmt.Stringer
	json.Marshaler
	json.Unmarshaler
}

/// ISBN10 ///

// ISBN10
//
// This is a book app, you should know what an ISBN10 is.
//
// Avoid directly instantiating this, and instead use `NewISBN10`, as
// it sterilizes the input into len 10
type ISBN10 struct {
	str string
}

// Create a new ISBN10
//
// This attempts to create a new ISBN10 from the given string, and
// returns an error if the newly created ISBN does not pass preliminary
// checks.
//
// You should almost always use this instead of creating an ISBN10
// directly.
func NewISBN10(from string) (ISBN10, error) {
	ns := isbn10re.FindAllString(from, -1)
	str := strings.Join(ns, "")

	i := ISBN10{str}
	if !i.Check() {
		return ISBN10{}, fmt.Errorf("could not parse into ISBN10: %v", from)
	}
	return i, nil
}

// Validate an ISBN10
//
// Note that this only checks a that the ISBN is made up of exactly 10
// digits (potentially separated by spaces or dashes) and that the end
// check digit is valid.
//
// It does not validate that the ISBN itself has ever been issued, much
// less if it corresponds to a given book.
//
// See https://isbn-information.com/the-10-digit-isbn.html
func (i *ISBN10) Check() bool {
	ns := isbn10re.FindAllString(i.str, -1)

	// For *some* reason the check digit can be 10. Which will be
	// represented... as a f****** X. We just make it a 10 for the rest
	// of this function
	if strings.ToUpper(ns[len(ns)-1]) == "X" {
		ns[len(ns)-1] = "10"
	}

	if len(ns) != 10 {
		return false
	}
	total := 0
	for i, v := range ns {
		weight := 10 - i
		num, err := strconv.Atoi(v)
		if err != nil {
			return false
		}
		total += num * weight
	}
	return total%11 == 0
}

func (i *ISBN10) String() string {
	return fmt.Sprintf("%v", i.str)
}

// Convert an ISBN10 into JSON
//
// This will also return an error if the check is not valid or if it
// can't be formatted into a 10 character string.
func (i ISBN10) MarshalJSON() ([]byte, error) {
	if !i.Check() {
		return []byte{}, fmt.Errorf("failed to marshall: ISBN did not pass preliminary check")
	}

	ns := isbn10re.FindAllString(i.str, -1)
	str := strings.Join(ns, "")
	if len(str) != 10 {
		return []byte{}, fmt.Errorf("failed to marshall: cleaned ISBN10 had len %d, expected 10", len(str))
	}

	return json.Marshal(struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Type:  "isbn10",
		Value: str,
	})
}

// Unmarshal a ISBN10 JSON object.
//
// Note that this also attempts to clean up the ISBN (e.g. removing
// dashes); failing if it doesn't pass checks or cannot fit into len 10
// like it should.
//
// NOTE: Because we expect this website to run in parallel, do we need
// to be locking these things?
func (i *ISBN10) UnmarshalJSON(b []byte) error {
	var jsonIsbn struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(b, &jsonIsbn); err != nil {
		return err
	}
	if jsonIsbn.Type != "isbn10" {
		return fmt.Errorf("invalid type for ISBN10: %s", jsonIsbn.Type)
	}

	ns := isbn10re.FindAllString(jsonIsbn.Value, -1)
	str := strings.Join(ns, "")
	if j, err := NewISBN10(str); err != nil {
		return fmt.Errorf("could not assign into %v, %s", j, err)
	} else {
		i.str = j.str
	}
	return nil
}

/// ISBN13 ///

// ISBN13
//
// This is a book app, you should know what an ISBN13 is.
//
// Avoid directly instantiating this, and instead use `NewISBN13`, as
// it sterilizes the input into len 13
type ISBN13 struct {
	str string
}

// Create a new ISBN13
//
// This attempts to create a new ISBN13 from the given string, and
// returns an error if the newly created ISBN does not pass preliminary
// checks.
//
// You should almost always use this instead of creating an ISBN13
// directly.
func NewISBN13(from string) (ISBN13, error) {
	ns := isbn13re.FindAllString(from, -1)
	str := strings.Join(ns, "")

	i := ISBN13{str}
	if !i.Check() {
		return ISBN13{}, fmt.Errorf("could not parse into ISBN13: %v", from)
	}
	return i, nil
}

func (i *ISBN13) String() string {
	return fmt.Sprintf("%v", i.str)
}

// Validate an ISBN13
//
// Note that this only checks a that the ISBN is made up of exactly 13
// digits (potentially separated by spaces or dashes) and that the end
// check digit is valid.
//
// It does not validate that the ISBN itself has ever been issued, much
// less if it corresponds to a given book.
//
// See https://isbn-information.com/the-13-digit-isbn.html
func (i *ISBN13) Check() bool {
	ns := isbn13re.FindAllString(i.str, -1)
	if len(ns) != 13 {
		return false
	}
	total := 0
	for i, v := range ns {
		weight := 1
		if i%2 != 0 {
			weight = 3
		}
		num, err := strconv.Atoi(v)
		if err != nil {
			return false
		}
		total += num * weight
	}
	return total%10 == 0
}

// Convert an ISBN13 into JSON
//
// This will also return an error if the check is not valid or if it
// can't be formatted into a 13 character string.
func (i ISBN13) MarshalJSON() ([]byte, error) {
	if !i.Check() {
		return []byte{}, fmt.Errorf("failed to marshall: ISBN did not pass preliminary check")
	}

	ns := isbn13re.FindAllString(i.str, -1)
	str := strings.Join(ns, "")
	if len(str) != 13 {
		return []byte{}, fmt.Errorf("failed to marshall: cleaned ISBN13 had len %d, expected 13", len(str))
	}

	return json.Marshal(struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Type:  "isbn13",
		Value: str,
	})
}

// Unmarshal a ISBN13 JSON object.
//
// Note that this also attempts to clean up the ISBN (e.g. removing
// dashes); failing if it doesn't pass checks or cannot fit into len 13
// like it should.
//
// NOTE: Because we expect this website to run in parallel, do we need
// to be locking these things?
func (i *ISBN13) UnmarshalJSON(b []byte) error {
	var jsonIsbn struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(b, &jsonIsbn); err != nil {
		return err
	}
	if jsonIsbn.Type != "isbn13" {
		return fmt.Errorf("invalid type for ISBN13: %s", jsonIsbn.Type)
	}

	ns := isbn13re.FindAllString(jsonIsbn.Value, -1)
	str := strings.Join(ns, "")
	if j, err := NewISBN13(str); err != nil {
		return fmt.Errorf("could not assign into %v, %s", j, err)
	} else {
		i.str = j.str
	}
	return nil
}

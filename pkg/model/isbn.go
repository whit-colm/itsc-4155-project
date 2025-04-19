package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const ISBNApiVersion string = "isbn.itsc-4155-group-project.edu.whits.io/v1"

var (
	isbn10re *regexp.Regexp
	isbn13re *regexp.Regexp
	//googidre *regexp.Regexp
	//lccnidre *regexp.Regexp
	//oclcidre *regexp.Regexp
)

func init() {
	isbn10re = regexp.MustCompile(`\d|[Xx]$`)
	isbn13re = regexp.MustCompile(`\d`)
	// NOTE: This is a guess, we don't actually know what the rules for
	// google ids are.
	//gvolidre = regexp.MustCompile(`[A-z0-9\-]{12}`)
	//lccnidre = regexp.MustCompile(``)
	//oclcidre = regexp.MustCompile(``)
}

type IsbnVersion int

func (v IsbnVersion) APIVersion() string {
	return ISBNApiVersion
}

const (
	ISBN10 IsbnVersion = iota
	ISBN13
	//GVOLID
	//LCCNID
	//OCLCID
)

func (v IsbnVersion) String() string {
	return [...]string{"isbn10", "isbn13" /*, "gvolID", "lccn", "oclc"*/}[v]
}

// Get the regex necessary to parse a valid ISBN of a given type from
// a string
//
// We use this in combination with .FindAllString(..., -1) in our
func (v IsbnVersion) regex() *regexp.Regexp {
	return [...]*regexp.Regexp{isbn10re, isbn13re /*, googidre, lccnidre, oclcidre*/}[v]
}

func (v IsbnVersion) Len() int {
	return [...]int{10, 13, 12}[v]
}

/**************************************
 **************** ISBN ****************
 **************************************/

func NewISBN(from string, ver ...IsbnVersion) (ISBN, error) {
	var i ISBN
	switch len(ver) {
	case 0:
		if i10s := ISBN10.regex().FindAllString(from, -1); len(i10s) == 10 {
			i = ISBN{strings.Join(i10s, ""), ISBN10}
		} else if i13s := ISBN13.regex().FindAllString(from, -1); len(i13s) == 13 {
			i = ISBN{strings.Join(i13s, ""), ISBN13}
		} else {
			return i, fmt.Errorf("unable to infer ISBN version: %s", from)
		}
	default:
		is := ver[0].regex().FindAllString(from, -1)
		i = ISBN{strings.Join(is, ""), ver[0]}
	}

	if !i.Check() {
		return ISBN{}, fmt.Errorf("could not parse into ISBN: %s as %s", from, ver)
	}

	return i, nil
}

// Forcibly create a new ISBN, panicking if one cannot be created.
func MustNewISBN(from string, ver ...IsbnVersion) ISBN {
	if i, err := NewISBN(from, ver...); err != nil {
		panic(`error creating ISBN: '` + err.Error() + `'`)
	} else {
		return i
	}
}

type ISBN struct {
	isbn    string
	version IsbnVersion
}

func (i ISBN) String() string {
	return i.isbn
}

func (i ISBN) Version() IsbnVersion {
	return i.version
}

func (i ISBN) Check() bool {
	ns := i.version.regex().FindAllString(i.isbn, -1)

	// Perform a preliminary check to make sure the length of the
	// filtered ISBN slice is the expected length given its version
	if len(ns) != i.version.Len() {
		return false
	}

	// This is fine because it only ever applies to ISBN10 due to how
	// the RegEx works out. Yes I kid you not the last digit can be X.
	// which means. 10.
	if strings.ToUpper(ns[len(ns)-1]) == "X" {
		ns[len(ns)-1] = "10"
	}

	switch i.version {
	case ISBN10:
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
	case ISBN13:
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
	default:
		return false
	}
}

func (i ISBN) MarshalJSON() ([]byte, error) {
	if !i.Check() {
		return []byte{}, fmt.Errorf("failed to marshall: ISBN did not pass preliminary check")
	}

	str := strings.Join(i.version.regex().FindAllString(i.isbn, -1), "")
	if len(str) != i.version.Len() {
		return []byte{}, fmt.Errorf("failed to marshall: cleaned `%s` had len %d, expected %d",
			i.version,
			len(str),
			i.version.Len())
	}

	return json.Marshal(struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Type:  i.version.String(),
		Value: str,
	})
}

func (i *ISBN) UnmarshalJSON(b []byte) error {
	var aux struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	var iv IsbnVersion
	switch aux.Type {
	case ISBN10.String():
		iv = ISBN10
	case ISBN13.String():
		iv = ISBN13
	default:
		return fmt.Errorf("did not recognize ISBN version: %s", aux.Type)
	}

	str := strings.Join(iv.regex().FindAllString(aux.Value, -1), "")
	if j, err := NewISBN(str, iv); err != nil {
		return fmt.Errorf("could not assign into ISBN: %v -> %v, %s", aux, j, err)
	} else {
		*i = j
	}
	return nil
}

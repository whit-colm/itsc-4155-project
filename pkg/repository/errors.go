package repository

import (
	"errors"
	"fmt"
)

/* This defines a series of agnostic (or abstract) datastore errors
 *
 * This is not meant to be comprehensive, but allows for cleaner code
 * by defining common errors that can be tested against using go 1.13+
 * error structure.
 */

var (
	ErrorNotFound = errors.New("not found")
)

type Error struct {
	Code error
	Err  error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Err)
	}
	return e.Code.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Is(target error) bool {
	return errors.Is(e.Code, target)
}

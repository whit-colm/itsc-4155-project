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
	ErrNotFound        = errors.New("not found")
	ErrBadConnection   = errors.New("bad connection")
	ErrBadTypecast     = errors.New("failed to typecast")
	ErrMultipleResults = errors.New("multiple results found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrUndefined       = errors.New("undefined error")
)

type Err struct {
	Code error
	Err  error
}

func (e Err) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Err)
	}
	return e.Code.Error()
}

func (e Err) Unwrap() error {
	return e.Err
}

func (e Err) Is(target error) bool {
	return errors.Is(e.Code, target)
}

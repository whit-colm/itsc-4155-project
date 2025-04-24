package repository

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorNotFound(t *testing.T) {
	assert.EqualError(t, ErrNotFound, "not found")
}

func TestErrorBadConnection(t *testing.T) {
	assert.EqualError(t, ErrBadConnection, "bad connection")
}

func TestErrorBadTypecast(t *testing.T) {
	assert.EqualError(t, ErrBadTypecast, "failed to typecast")
}

func TestErr_Error_WithInnerError(t *testing.T) {
	inner := errors.New("db timeout")
	e := Err{Code: ErrBadConnection, Err: inner}
	assert.Equal(t, "bad connection: db timeout", e.Error())
}

func TestErr_Error_WithoutInnerError(t *testing.T) {
	e := Err{Code: ErrNotFound}
	assert.Equal(t, "not found", e.Error())
}

func TestErr_Unwrap(t *testing.T) {
	inner := errors.New("db timeout")
	e := &Err{Code: ErrBadConnection, Err: inner}
	assert.Equal(t, inner, errors.Unwrap(e))
}

func TestErr_Is(t *testing.T) {
	e := &Err{Code: ErrNotFound}
	assert.True(t, errors.Is(e, ErrNotFound))
	assert.False(t, errors.Is(e, ErrBadConnection))
}

func TestErr_Is_WithWrappedError(t *testing.T) {
	inner := errors.New("some detail")
	e := &Err{Code: ErrBadTypecast, Err: inner}
	assert.True(t, errors.Is(e, ErrBadTypecast))
	assert.False(t, errors.Is(e, ErrNotFound))
}

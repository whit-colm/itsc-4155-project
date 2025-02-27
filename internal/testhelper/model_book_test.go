package testhelper

import (
	"testing"
)

func TestIsBookEquals(t *testing.T) {
	if !IsBookEquals(ExampleBook, ExampleBook) {
		t.Error("known equal books are unequal")
	}
}

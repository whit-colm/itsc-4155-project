package testhelper

import (
	"testing"
)

func TestIsBookEquals(t *testing.T) {
	if !IsBookEquals(ExampleBook, ExampleBook) {
		t.Error("known equal books are unequal")
	}

	if IsBookEquals(ExampleBook, ExampleBooks[0]) {
		t.Error("known unequal books are equal")
	}
}

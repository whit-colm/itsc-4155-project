package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestUsernameFromString_Valid(t *testing.T) {
	tests := []struct {
		input                 string
		expectedHandle        string
		expectedDiscriminator int16
	}{
		{"user#1234", "user", 1234},
		{"system#0000", "system", 0},
		{"user #5678", "user ", 5678}, // Handle with trailing space (code allows)
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			un, err := UsernameFromString(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			handle, discriminator := un.Components()
			if handle != tt.expectedHandle {
				t.Errorf("handle: got %q, want %q", handle, tt.expectedHandle)
			}
			if discriminator != tt.expectedDiscriminator {
				t.Errorf("discriminator: got %d, want %d", discriminator, tt.expectedDiscriminator)
			}
		})
	}
}

func TestUsernameFromString_Invalid(t *testing.T) {
	invalidInputs := []string{
		"user",         // Missing discriminator
		"user#12a4",    // Non-digit discriminator
		"user#123",     // Short discriminator
		"@user#1234",   // Invalid handle character
		" user#1234",   // Leading space (invalid)
		"invalid#0000", // Non-protected handle with 0000
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			_, err := UsernameFromString(input)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestUsernameFromComponents_Valid(t *testing.T) {
	tests := []struct {
		handle        string
		discriminator int
	}{
		{"user", 1234},
		{"system", 0},   // Protected handle with 0
		{"user ", 5678}, // Trailing space (allowed by code)
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s#%d", tt.handle, tt.discriminator), func(t *testing.T) {
			_, err := UsernameFromComponents(tt.handle, tt.discriminator)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestUsernameFromComponents_InvalidHandle(t *testing.T) {
	tests := []struct {
		handle        string
		discriminator int
	}{
		{"a", 1234},                     // Too short
		{strings.Repeat("a", 33), 1234}, // Too long
		{"us@r", 1234},                  // Invalid character
		{" user", 1234},                 // Leading space
	}

	for _, tt := range tests {
		t.Run(tt.handle, func(t *testing.T) {
			_, err := UsernameFromComponents(tt.handle, tt.discriminator)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestUsernameFromComponents_InvalidDiscriminator(t *testing.T) {
	tests := []struct {
		handle        string
		discriminator int
	}{
		{"user", 0},       // Non-protected handle with 0
		{"user", 10000},   // Too high
		{"user", -1},      // Negative
		{"system", 10000}, // Protected handle with invalid discriminator
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s#%d", tt.handle, tt.discriminator), func(t *testing.T) {
			_, err := UsernameFromComponents(tt.handle, tt.discriminator)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestUsernameFromHandle(t *testing.T) {
	tests := []struct {
		handle string
		valid  bool
	}{
		{"validhandle", true},
		{"invalid@handle", false},
		{" ", false}, // Too short
	}

	for _, tt := range tests {
		t.Run(tt.handle, func(t *testing.T) {
			un, err := UsernameFromHandle(tt.handle)
			if tt.valid {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if un.handle != tt.handle || un.discriminator != 0 {
					t.Errorf("got %v, expected handle %q discriminator 0", un, tt.handle)
				}
			} else {
				if err == nil {
					t.Error("expected error, got nil")
				}
			}
		})
	}
}

func TestUsernameValidate(t *testing.T) {
	tests := []struct {
		name      string
		handle    string
		disc      int16
		canBeZero bool
		valid     bool
	}{
		{"valid", "user", 1234, false, true},
		{"invalid handle", "us@r", 1234, false, false},
		{"discrim 0 allowed", "system", 0, true, true},
		{"discrim 0 disallowed", "user", 0, false, false},
		{"trailing space handle", "user ", 1234, false, true}, // Allowed per code
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			un := Username{handle: tt.handle, discriminator: tt.disc}
			valid := un.Validate(tt.canBeZero)
			if valid != tt.valid {
				t.Errorf("got %v, want %v", valid, tt.valid)
			}
		})
	}
}

func TestUsernameJSONMarshaling(t *testing.T) {
	un := Username{handle: "user", discriminator: 1234}
	expectedJSON := `"user#1234"`

	bytes, err := json.Marshal(un)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	if string(bytes) != expectedJSON {
		t.Errorf("got %s, want %s", string(bytes), expectedJSON)
	}
}

func TestUsernameJSONUnmarshaling_Valid(t *testing.T) {
	tests := []struct {
		jsonStr  string
		expected Username
	}{
		{`"user#1234"`, Username{handle: "user", discriminator: 1234}},
		{`"user #5678"`, Username{handle: "user ", discriminator: 5678}},
		{`"system#0000"`, Username{handle: "system", discriminator: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.jsonStr, func(t *testing.T) {
			var un Username
			err := json.Unmarshal([]byte(tt.jsonStr), &un)
			if err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}
			if un.handle != tt.expected.handle || un.discriminator != tt.expected.discriminator {
				t.Errorf("got %v, want %v", un, tt.expected)
			}
		})
	}
}

func TestUsernameJSONUnmarshaling_Invalid(t *testing.T) {
	invalidJSONs := []string{
		`"user"`,
		`"user#12a4"`,
		`"user#123"`,
		`"@user#1234"`,
	}

	for _, jsonStr := range invalidJSONs {
		t.Run(jsonStr, func(t *testing.T) {
			var un Username
			err := json.Unmarshal([]byte(jsonStr), &un)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

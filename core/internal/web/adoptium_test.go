package web

import (
	"testing"
)

func TestParseVersionNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"17", 17},
		{"17.0.2", 17},
		{"21", 21},
		{"8", 8},
	}

	for _, tt := range tests {
		result := parseVersionNumber(tt.input)
		if result != tt.expected {
			t.Errorf("parseVersionNumber(%s) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}
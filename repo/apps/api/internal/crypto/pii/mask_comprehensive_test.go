package pii

import (
	"testing"
)

func TestPartialMaskPhone_comprehensive(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"1234", "••••"},
		{"13800138000", "••••••8000"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := PartialMaskPhone(tt.input)
			if got != tt.want {
				t.Errorf("PartialMaskPhone(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPartialMaskID_comprehensive(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"12345", "••••••2345"},
		{"110101199001011234", "••••••1234"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := PartialMaskID(tt.input)
			if got != tt.want {
				t.Errorf("PartialMaskID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPartialMaskEmail_comprehensive(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"a@b.com", "a•••@b.com"},
		{"alice@example.com", "a•••@example.com"},
		{"very.long.email.address@company.org", "v•••@company.org"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := PartialMaskEmail(tt.input)
			if got != tt.want {
				t.Errorf("PartialMaskEmail(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

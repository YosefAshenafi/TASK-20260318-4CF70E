package middleware

import "testing"

func TestBearerToken(t *testing.T) {
	if BearerToken("Bearer abc123") != "abc123" {
		t.Fatal("expected token")
	}
	if BearerToken("bearer lowercase") != "lowercase" {
		t.Fatal("expected case-insensitive scheme")
	}
	if BearerToken("") != "" || BearerToken("Basic x") != "" || BearerToken("Bearer") != "" {
		t.Fatal("expected empty for invalid headers")
	}
	if BearerToken("Bearer  spaced  ") != "spaced" {
		t.Fatal("trim spaces")
	}
}

package pii

import "testing"

func TestPartialMaskPhone(t *testing.T) {
	if got := PartialMaskPhone("13800138000"); got == "" || len(got) < 4 {
		t.Fatalf("got %q", got)
	}
	if PartialMaskPhone("") != "" {
		t.Fatal("empty")
	}
}

func TestPartialMaskEmail(t *testing.T) {
	if got := PartialMaskEmail("alice@example.com"); got == "" {
		t.Fatal("empty")
	}
}

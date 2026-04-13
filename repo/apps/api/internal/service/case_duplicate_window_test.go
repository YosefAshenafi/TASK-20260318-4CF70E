package service

import (
	"testing"
)

func Test_duplicateContentHash_differentInputs(t *testing.T) {
	inst := "10000000-0000-4000-8000-000000000001"
	baseHash := duplicateContentHash(inst, "ADR", "Title", "Description", "2026-04-13T12:00:00Z")

	if baseHash == "" {
		t.Fatal("hash should not be empty")
	}

	cases := []struct {
		name string
		inst, caseType, title, desc, reported string
	}{
		{"diff_institution", "20000000-0000-4000-8000-000000000002", "ADR", "Title", "Description", "2026-04-13T12:00:00Z"},
		{"diff_type", inst, "QUALITY", "Title", "Description", "2026-04-13T12:00:00Z"},
		{"diff_title", inst, "ADR", "Different Title", "Description", "2026-04-13T12:00:00Z"},
		{"diff_desc", inst, "ADR", "Title", "Different Description", "2026-04-13T12:00:00Z"},
		{"diff_time", inst, "ADR", "Title", "Description", "2026-04-13T13:00:00Z"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := duplicateContentHash(tc.inst, tc.caseType, tc.title, tc.desc, tc.reported)
			if h == baseHash {
				t.Fatalf("expected different hash for %s", tc.name)
			}
		})
	}
}

func Test_duplicateContentHash_deterministic(t *testing.T) {
	for i := 0; i < 10; i++ {
		h := duplicateContentHash("i", "t", "title", "desc", "ts")
		if h != duplicateContentHash("i", "t", "title", "desc", "ts") {
			t.Fatal("hash is not deterministic")
		}
	}
}

func Test_allowedTransition_comprehensive(t *testing.T) {
	tests := []struct {
		from, to string
		allowed  bool
	}{
		{"submitted", "assigned", true},
		{"submitted", "in_progress", true},
		{"submitted", "closed", true},
		{"submitted", "pending_review", false},
		{"assigned", "in_progress", true},
		{"assigned", "pending_review", true},
		{"assigned", "closed", true},
		{"in_progress", "pending_review", true},
		{"in_progress", "closed", true},
		{"in_progress", "assigned", true},
		{"pending_review", "closed", true},
		{"pending_review", "in_progress", true},
		{"pending_review", "submitted", false},
		{"closed", "submitted", false},
		{"closed", "assigned", false},
		{"closed", "in_progress", false},
		{"closed", "pending_review", false},
	}
	for _, tt := range tests {
		t.Run(tt.from+"_to_"+tt.to, func(t *testing.T) {
			got := allowedTransition(tt.from, tt.to)
			if got != tt.allowed {
				t.Errorf("allowedTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.allowed)
			}
		})
	}

	for _, status := range []string{"submitted", "assigned", "in_progress", "pending_review", "closed"} {
		if allowedTransition(status, status) {
			t.Errorf("self-transition should be disallowed for %q", status)
		}
	}

	if allowedTransition("unknown_status", "assigned") {
		t.Error("unknown source status should not allow transitions")
	}
}

package service

import (
	"fmt"
	"testing"
	"time"
)

// Design.md §15.1 — Format: YYYYMMDD-{institution}-{6-digit serial}
func TestCaseNumberFormat_designSpec(t *testing.T) {
	dayUTC := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	y := dayUTC.Format("20060102")
	code := "ACME"
	serial := uint32(7)
	got := fmt.Sprintf("%s-%s-%06d", y, code, serial)
	want := "20260413-ACME-000007"
	if got != want {
		t.Fatalf("caseNumber: got %q want %q", got, want)
	}
}

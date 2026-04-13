package service

import (
	"context"
	"testing"
)

func TestAuditService_LogCandidatePIIRead_nilRepo_noPanic(t *testing.T) {
	s := &AuditService{repo: nil}
	err := s.LogCandidatePIIRead(context.Background(), "user-1", "cand-1", "req-1", nil, []string{"phone", "email"})
	if err != nil {
		t.Fatal(err)
	}
}

package service

import (
	"strings"
	"testing"

	_ "embed"
)

//go:embed rbac_service.go
var rbacSrc string

//go:embed recruitment_service.go
var recruitmentSrc string

//go:embed recruitment_extended.go
var recruitmentExtendedSrc string

//go:embed compliance_service.go
var complianceSrc string

//go:embed case_service.go
var caseSrc string

//go:embed file_service.go
var fileSrc string

// Ensures cross-module mutation paths invoke append-only audit writer (design §7.6, §17).
func TestMutationServices_referenceLogMutation(t *testing.T) {
	for _, name := range []struct {
		n string
		s string
	}{
		{"rbac_service.go", rbacSrc},
		{"recruitment_service.go", recruitmentSrc},
		{"recruitment_extended.go", recruitmentExtendedSrc},
		{"compliance_service.go", complianceSrc},
		{"case_service.go", caseSrc},
		{"file_service.go", fileSrc},
	} {
		if !strings.Contains(name.s, "LogMutation") {
			t.Fatalf("%s: expected LogMutation calls for audited mutations", name.n)
		}
	}
}

package httpserver

import (
	"strings"
	"testing"

	_ "embed"
)

//go:embed server.go
var serverGoSource string

// TestRecruitmentExtendedRoutesFromAPISpec guards docs/api-spec.md recruitment operations
// (imports, duplicates, merge, merge-history, match, recommendations) against route drift.
func TestRecruitmentExtendedRoutesFromAPISpec(t *testing.T) {
	required := []string{
		`POST("/recruitment/candidates/imports"`,
		`GET("/recruitment/candidates/imports/:importId"`,
		`POST("/recruitment/candidates/imports/:importId/commit"`,
		`GET("/recruitment/candidates/duplicates"`,
		`POST("/recruitment/candidates/merge"`,
		`GET("/recruitment/candidates/merge-history"`,
		`POST("/recruitment/match/candidate-to-position"`,
		`POST("/recruitment/match/position-to-candidate"`,
		`GET("/recruitment/recommendations/similar-candidates/:candidateId"`,
		`GET("/recruitment/recommendations/similar-positions/:positionId"`,
	}
	for _, sub := range required {
		if !strings.Contains(serverGoSource, sub) {
			t.Errorf("server route registration missing substring %q (see docs/api-spec.md recruitment section)", sub)
		}
	}
}

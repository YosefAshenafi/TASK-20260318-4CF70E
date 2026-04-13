package service

import (
	"testing"

	"pharmaops/api/internal/model"
)

func TestComputeMatchScore_fullMatch(t *testing.T) {
	cand := &model.Candidate{
		ExperienceYears: intPtr(10),
		EducationLevel:  strPtr("PhD"),
		Skills: []model.CandidateSkill{
			{SkillName: "GMP"}, {SkillName: "QA"},
		},
	}
	reqs := []model.PositionRequirement{
		{SkillName: "GMP"}, {SkillName: "QA"},
	}
	pos := &model.Position{Title: "Quality Manager"}
	dto, _, reasons := computeMatchScore(cand, []string{"GMP", "QA"}, pos, reqs)

	if dto.Score < 80 {
		t.Errorf("expected high score for full match, got %d", dto.Score)
	}
	if dto.Score > 100 {
		t.Errorf("score should not exceed 100, got %d", dto.Score)
	}
	if dto.Breakdown.Skills != 50 {
		t.Errorf("expected max skills=50 for 2/2 match, got %d", dto.Breakdown.Skills)
	}
	if len(reasons) < 3 {
		t.Error("expected at least 3 reason lines")
	}
}

func TestComputeMatchScore_noMatch(t *testing.T) {
	cand := &model.Candidate{
		Skills: []model.CandidateSkill{{SkillName: "Python"}},
	}
	reqs := []model.PositionRequirement{{SkillName: "Java"}}
	pos := &model.Position{Title: "Java Developer"}
	dto, _, _ := computeMatchScore(cand, []string{"Python"}, pos, reqs)

	if dto.Breakdown.Skills != 0 {
		t.Errorf("expected 0 skills for no overlap, got %d", dto.Breakdown.Skills)
	}
	if dto.Breakdown.Experience != 0 {
		t.Errorf("expected 0 experience for nil years, got %d", dto.Breakdown.Experience)
	}
}

func TestComputeMatchScore_partialSkills(t *testing.T) {
	cand := &model.Candidate{
		Skills: []model.CandidateSkill{{SkillName: "GMP"}, {SkillName: "Python"}},
	}
	reqs := []model.PositionRequirement{
		{SkillName: "GMP"}, {SkillName: "QA"}, {SkillName: "Audit"},
	}
	pos := &model.Position{Title: "QA Manager"}
	dto, _, _ := computeMatchScore(cand, []string{"GMP", "Python"}, pos, reqs)

	// 1 out of 3 = 33.3% of 50 ≈ 17
	if dto.Breakdown.Skills < 10 || dto.Breakdown.Skills > 20 {
		t.Errorf("expected partial skills ~17, got %d", dto.Breakdown.Skills)
	}
}

func TestComputeMatchScore_noRequirements_tokenizeTitle(t *testing.T) {
	cand := &model.Candidate{
		Skills: []model.CandidateSkill{{SkillName: "quality"}, {SkillName: "manager"}},
	}
	pos := &model.Position{Title: "Quality Manager"}
	dto, _, _ := computeMatchScore(cand, []string{"quality", "manager"}, pos, nil)
	if dto.Breakdown.Skills == 0 {
		t.Error("expected non-zero skills from title token matching")
	}
}

func TestComputeMatchScore_experienceCapped(t *testing.T) {
	cand := &model.Candidate{
		ExperienceYears: intPtr(20),
	}
	pos := &model.Position{Title: "Senior"}
	dto, _, _ := computeMatchScore(cand, nil, pos, nil)
	if dto.Breakdown.Experience != 30 {
		t.Errorf("expected experience capped at 30, got %d", dto.Breakdown.Experience)
	}
}

func TestEducationMatchPoints(t *testing.T) {
	tests := []struct {
		level *string
		want  int
	}{
		{strPtr("PhD in Chemistry"), 20},
		{strPtr("Doctor of Pharmacy"), 20},
		{strPtr("Master of Science"), 16},
		{strPtr("Bachelor"), 12},
		{strPtr("本科"), 12},
		{strPtr("硕士"), 16},
		{strPtr("Diploma"), 8},
		{strPtr(""), 0},
		{nil, 0},
	}
	for _, tt := range tests {
		label := "<nil>"
		if tt.level != nil {
			label = *tt.level
		}
		t.Run(label, func(t *testing.T) {
			got := educationMatchPoints(tt.level)
			if got != tt.want {
				t.Errorf("educationMatchPoints(%q) = %d, want %d", label, got, tt.want)
			}
		})
	}
}

func TestTokenizeTitle(t *testing.T) {
	tokens := tokenizeTitle("Quality Assurance / GMP Manager (Senior)")
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
	// Should not include short words like "gmp" if < 3 chars (it's exactly 3, so it should be included)
	found := false
	for _, tk := range tokens {
		if tk == "gmp" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'gmp' token from title")
	}
}

func intPtr(n int) *int       { return &n }

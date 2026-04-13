package service

import (
	"encoding/json"
	"testing"
)

func Test_parseRestrictionRule_branches(t *testing.T) {
	r, err := parseRestrictionRule([]byte(`{"requiresPrescription":true,"frequencyDays":14}`))
	if err != nil {
		t.Fatal(err)
	}
	if !r.RequiresPrescription || r.FrequencyDays != 14 {
		t.Fatalf("got %+v", r)
	}
	r2, err := parseRestrictionRule(nil)
	if err != nil || r2.RequiresPrescription || r2.FrequencyDays != 0 {
		t.Fatalf("empty: %+v err=%v", r2, err)
	}
	_, err = parseRestrictionRule([]byte(`not-json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func Test_parseRestrictionRule_roundTrip(t *testing.T) {
	raw := map[string]any{"requiresPrescription": false, "frequencyDays": 7}
	b, _ := json.Marshal(raw)
	r, err := parseRestrictionRule(b)
	if err != nil || r.FrequencyDays != 7 {
		t.Fatalf("got %+v err=%v", r, err)
	}
}

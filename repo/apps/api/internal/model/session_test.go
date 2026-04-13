package model

import "testing"

func TestSession_TableName(t *testing.T) {
	var s Session
	if s.TableName() != "sessions" {
		t.Fatalf("TableName: %s", s.TableName())
	}
}

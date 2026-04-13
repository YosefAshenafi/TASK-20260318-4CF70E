package model

import "testing"

func TestUser_TableName(t *testing.T) {
	var u User
	if u.TableName() != "users" {
		t.Fatalf("TableName: %s", u.TableName())
	}
}

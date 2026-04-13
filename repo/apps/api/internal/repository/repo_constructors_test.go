package repository

import "testing"

func TestConstructorPointers_nonNil(t *testing.T) {
	if NewUserRepository(nil) == nil {
		t.Fatal("NewUserRepository")
	}
	if NewSessionRepository(nil) == nil {
		t.Fatal("NewSessionRepository")
	}
	if NewAccessRepository(nil) == nil {
		t.Fatal("NewAccessRepository")
	}
}

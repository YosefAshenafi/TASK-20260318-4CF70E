package httpserver

import (
	"testing"

	"pharmaops/api/internal/config"
)

func TestNew_nonNil(t *testing.T) {
	s := New(config.Config{}, nil)
	if s == nil {
		t.Fatal("New")
	}
}

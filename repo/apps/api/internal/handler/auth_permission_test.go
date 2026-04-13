package handler

import (
	"testing"

	"pharmaops/api/internal/access"
)

func Test_permissionCodesSorted(t *testing.T) {
	p := &access.Principal{
		PermissionSet: map[string]struct{}{
			"z.last":  {},
			"a.first": {},
		},
	}
	got := permissionCodesSorted(p)
	if len(got) != 2 || got[0] != "a.first" || got[1] != "z.last" {
		t.Fatalf("unexpected order: %v", got)
	}
	if permissionCodesSorted(nil) != nil {
		t.Fatal("nil principal should yield nil slice")
	}
}

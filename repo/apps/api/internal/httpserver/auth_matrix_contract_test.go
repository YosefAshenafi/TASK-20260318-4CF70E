package httpserver

import (
	"strings"
	"testing"
)

func TestProtectedModulesStayBehindAuthAndPermissionMiddleware(t *testing.T) {
	required := []string{
		`authz.Use(middleware.SessionAuth(authSvc))`,
		`authz.Use(middleware.AccessContext(accessSvc))`,
		`RequirePermission("recruitment.view")`,
		`RequirePermission("compliance.view")`,
		`RequirePermission("cases.view")`,
		`RequirePermission("files.view")`,
		`RequirePermission("audit.view")`,
	}
	for _, sub := range required {
		if !strings.Contains(serverGoSource, sub) {
			t.Fatalf("missing authz guard substring %q in server routes", sub)
		}
	}
}

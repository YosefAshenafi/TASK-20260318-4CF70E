package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"pharmaops/api/internal/access"
)

func TestEndpointPermissionMatrix_forbiddenWithoutModulePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		method         string
		path           string
		required       string
		permissionSet  map[string]struct{}
		expectedStatus int
	}{
		{
			name:           "recruitment view denied",
			method:         http.MethodGet,
			path:           "/api/v1/recruitment/candidates",
			required:       "recruitment.view",
			permissionSet:  map[string]struct{}{"cases.view": {}},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "compliance manage denied",
			method:         http.MethodPost,
			path:           "/api/v1/compliance/restrictions/check-purchase",
			required:       "compliance.manage",
			permissionSet:  map[string]struct{}{"compliance.view": {}},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "cases manage denied",
			method:         http.MethodPost,
			path:           "/api/v1/cases",
			required:       "cases.manage",
			permissionSet:  map[string]struct{}{"cases.view": {}},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "files manage denied",
			method:         http.MethodPost,
			path:           "/api/v1/files/uploads/init",
			required:       "files.manage",
			permissionSet:  map[string]struct{}{"files.view": {}},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "audit view denied",
			method:         http.MethodGet,
			path:           "/api/v1/audit/logs",
			required:       "audit.view",
			permissionSet:  map[string]struct{}{"files.view": {}},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "system rbac denied",
			method:         http.MethodGet,
			path:           "/api/v1/users",
			required:       "system.rbac",
			permissionSet:  map[string]struct{}{"audit.view": {}},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "full access bypasses module permission",
			method:         http.MethodGet,
			path:           "/api/v1/users",
			required:       "system.rbac",
			permissionSet:  map[string]struct{}{access.PermissionFullAccess: {}},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set(ctxKeyPrincipal, &access.Principal{PermissionSet: tt.permissionSet})
				c.Next()
			})
			r.Handle(tt.method, tt.path, RequirePermission(tt.required), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			r.ServeHTTP(w, req)
			if w.Code != tt.expectedStatus {
				t.Fatalf("expected %d, got %d body=%s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/service"
)

func TestAccessContext_missingUserID_unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessContext(service.NewAccessService(nil)))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without userID, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRequirePermission_forbiddenWithoutCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		p := &access.Principal{
			PermissionSet: map[string]struct{}{"other.perm": {}},
		}
		c.Set(ctxKeyPrincipal, p)
		c.Next()
	})
	r.Use(RequirePermission("cases.manage"))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRequirePermission_okWithCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		p := &access.Principal{
			PermissionSet: map[string]struct{}{"cases.manage": {}},
		}
		c.Set(ctxKeyPrincipal, p)
		c.Next()
	})
	r.Use(RequirePermission("cases.manage"))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRequirePermission_fullAccess_ok(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		p := &access.Principal{
			PermissionSet: map[string]struct{}{access.PermissionFullAccess: {}},
		}
		c.Set(ctxKeyPrincipal, p)
		c.Next()
	})
	r.Use(RequirePermission("audit.view"))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with full access, got %d", w.Code)
	}
}

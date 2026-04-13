package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/response"
)

// RequirePermission returns 403 when the principal lacks the permission (after AccessContext).
func RequirePermission(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, ok := c.Get(ctxKeyPrincipal)
		if !ok {
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "access context missing")
			c.Abort()
			return
		}
		principal := p.(*access.Principal)
		if !principal.Has(code) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN_PERMISSION", "missing permission: "+code)
			c.Abort()
			return
		}
		c.Next()
	}
}

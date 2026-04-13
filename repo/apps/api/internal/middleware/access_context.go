package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/response"
	"pharmaops/api/internal/service"
)

const ctxKeyPrincipal = "accessPrincipal"

// GetPrincipal returns the principal set by AccessContext.
func GetPrincipal(c *gin.Context) (*access.Principal, bool) {
	v, ok := c.Get(ctxKeyPrincipal)
	if !ok {
		return nil, false
	}
	p, ok := v.(*access.Principal)
	return p, ok
}

// AccessContext loads RBAC + data scopes after SessionAuth (requires userID).
func AccessContext(access *service.AccessService) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("userID")
		if uid == "" {
			response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing user context")
			c.Abort()
			return
		}
		p, err := access.LoadPrincipal(c.Request.Context(), uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load access context")
			c.Abort()
			return
		}
		c.Set(ctxKeyPrincipal, p)
		c.Next()
	}
}

package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pharmaops/api/internal/oplog"
	"pharmaops/api/internal/response"
	"pharmaops/api/internal/service"
)

func BearerToken(header string) string {
	if header == "" {
		return ""
	}
	const p = "Bearer "
	if len(header) <= len(p) || !strings.EqualFold(header[:len(p)], p) {
		return ""
	}
	return strings.TrimSpace(header[len(p):])
}

func SessionAuth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetString("requestId")
		ip := c.ClientIP()
		t := BearerToken(c.GetHeader("Authorization"))
		if t == "" {
			oplog.SessionInvalid(rid, ip, "missing bearer token")
			response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing bearer token")
			c.Abort()
			return
		}
		uid, err := auth.SessionUserID(c.Request.Context(), t)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				oplog.SessionInvalid(rid, ip, "invalid or expired session")
				response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "invalid or expired session")
				c.Abort()
				return
			}
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "session validation failed")
			c.Abort()
			return
		}
		c.Set("userID", uid)
		c.Next()
	}
}

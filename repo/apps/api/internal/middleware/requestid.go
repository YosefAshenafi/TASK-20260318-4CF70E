package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"pharmaops/api/internal/response"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(response.HeaderRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set("requestId", rid)
		c.Writer.Header().Set(response.HeaderRequestID, rid)
		c.Next()
	}
}

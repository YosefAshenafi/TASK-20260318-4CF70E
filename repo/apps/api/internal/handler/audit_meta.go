package handler

import (
	"github.com/gin-gonic/gin"
	"pharmaops/api/internal/service"
)

func auditRequestMeta(c *gin.Context) service.AuditRequestMeta {
	ip := c.ClientIP()
	meta := service.AuditRequestMeta{
		OperatorUserID: c.GetString("userID"),
		RequestID:      c.GetString("requestId"),
		RequestSource:  &ip,
	}
	return meta
}

package handler

import (
	"github.com/gin-gonic/gin"

	"pharmaops/api/internal/middleware"
	"pharmaops/api/internal/service"
)

func auditRequestMeta(c *gin.Context) service.AuditRequestMeta {
	ip := c.ClientIP()
	meta := service.AuditRequestMeta{
		OperatorUserID: c.GetString("userID"),
		RequestID:      c.GetString("requestId"),
		RequestSource:  &ip,
	}
	if pr, ok := middleware.GetPrincipal(c); ok && pr != nil && len(pr.Scopes) > 0 {
		meta.InstitutionID = &pr.Scopes[0].InstitutionID
		meta.DepartmentID = pr.Scopes[0].DepartmentID
		meta.TeamID = pr.Scopes[0].TeamID
	}
	return meta
}

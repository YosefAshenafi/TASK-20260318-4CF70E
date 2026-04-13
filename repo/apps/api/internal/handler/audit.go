package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pharmaops/api/internal/middleware"
	"pharmaops/api/internal/response"
	"pharmaops/api/internal/service"
)

type AuditHandler struct {
	svc *service.AuditService
}

func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) ListLogs(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	page, pageSize, offset := ParsePagination(c)
	sortBy := c.DefaultQuery("sortBy", "created_at")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	module := c.Query("module")
	targetType := c.Query("targetType")
	var fromPtr, toPtr *time.Time
	if fs := c.Query("from"); fs != "" {
		t, err := time.Parse(time.RFC3339, fs)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "from must be RFC3339")
			return
		}
		fromPtr = &t
	}
	if ts := c.Query("to"); ts != "" {
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "to must be RFC3339")
			return
		}
		toPtr = &t
	}
	items, total, page, pageSize, err := h.svc.ListAuditLogs(c.Request.Context(), page, pageSize, offset, sortBy, sortOrder, service.ListAuditLogsInput{
		Module:     module,
		TargetType: targetType,
		From:       fromPtr,
		To:         toPtr,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list audit logs")
		return
	}
	response.OK(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

type auditExportBody struct {
	Module     string `json:"module"`
	TargetType string `json:"targetType"`
	From       string `json:"from"`
	To         string `json:"to"`
}

func (h *AuditHandler) RequestExport(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	uid := c.GetString("userID")
	if uid == "" {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing user")
		return
	}
	var body auditExportBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	dto, err := h.svc.RequestExport(c.Request.Context(), uid, service.AuditExportFilter{
		Module:     body.Module,
		TargetType: body.TargetType,
		From:       body.From,
		To:         body.To,
	}, auditRequestMeta(c))
	if errors.Is(err, service.ErrAuditExportValidation) {
		response.Error(c, http.StatusBadRequest, "EXPORT_VALIDATION_FAILED", "invalid export filter")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to request export")
		return
	}
	response.OK(c, dto)
}

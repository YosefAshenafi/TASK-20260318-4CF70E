package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"pharmaops/api/internal/middleware"
	"pharmaops/api/internal/repository"
	"pharmaops/api/internal/response"
	"pharmaops/api/internal/service"
)

type ComplianceHandler struct {
	svc *service.ComplianceService
}

func NewComplianceHandler(svc *service.ComplianceService) *ComplianceHandler {
	return &ComplianceHandler{svc: svc}
}

func (h *ComplianceHandler) ListQualifications(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	page, pageSize, offset := ParsePagination(c)
	sortBy := c.DefaultQuery("sortBy", "created_at")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	items, total, page, pageSize, err := h.svc.ListQualifications(c.Request.Context(), pr, page, pageSize, offset, sortBy, sortOrder)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list qualifications")
		return
	}
	response.OK(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *ComplianceHandler) ListExpiringQualifications(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	days := 30
	if d := c.Query("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 {
			days = n
		}
	}
	items, err := h.svc.ListExpiringQualifications(c.Request.Context(), pr, days)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list expiring qualifications")
		return
	}
	response.OK(c, gin.H{"items": items})
}

func (h *ComplianceHandler) GetQualification(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.GetQualification(c.Request.Context(), pr, id)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "QUALIFICATION_NOT_FOUND", "qualification not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load qualification")
		return
	}
	response.OK(c, dto)
}

type createQualificationBody struct {
	InstitutionID string         `json:"institutionId" binding:"required"`
	ClientID      string         `json:"clientId" binding:"required"`
	DisplayName   string         `json:"displayName" binding:"required"`
	ExpiresOn     *string        `json:"expiresOn"`
	Metadata      map[string]any `json:"metadata"`
}

func (h *ComplianceHandler) CreateQualification(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body createQualificationBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	dto, err := h.svc.CreateQualification(c.Request.Context(), pr, service.CreateQualificationInput{
		InstitutionID: body.InstitutionID,
		ClientID:      body.ClientID,
		DisplayName:   body.DisplayName,
		ExpiresOn:     body.ExpiresOn,
		Metadata:      body.Metadata,
	}, auditRequestMeta(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create qualification")
		return
	}
	response.OK(c, dto)
}

type patchQualificationBody struct {
	DisplayName *string        `json:"displayName"`
	ExpiresOn   *string        `json:"expiresOn"`
	Metadata    map[string]any `json:"metadata"`
	Status      *string        `json:"status"`
}

func (h *ComplianceHandler) PatchQualification(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body patchQualificationBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.UpdateQualification(c.Request.Context(), pr, id, service.UpdateQualificationInput{
		DisplayName: body.DisplayName,
		ExpiresOn:   body.ExpiresOn,
		Metadata:    body.Metadata,
		Status:      body.Status,
	}, auditRequestMeta(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "QUALIFICATION_NOT_FOUND", "qualification not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update qualification")
		return
	}
	response.OK(c, dto)
}

func (h *ComplianceHandler) ActivateQualification(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.ActivateQualification(c.Request.Context(), pr, id, auditRequestMeta(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "QUALIFICATION_NOT_FOUND", "qualification not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to activate qualification")
		return
	}
	response.OK(c, dto)
}

func (h *ComplianceHandler) DeactivateQualification(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.DeactivateQualification(c.Request.Context(), pr, id, auditRequestMeta(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "QUALIFICATION_NOT_FOUND", "qualification not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to deactivate qualification")
		return
	}
	response.OK(c, dto)
}

func (h *ComplianceHandler) RunQualificationJob(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	n, err := h.svc.RunQualificationExpirationJob(c.Request.Context(), pr)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "job failed")
		return
	}
	response.OK(c, gin.H{"deactivated": n})
}

func (h *ComplianceHandler) ListRestrictions(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	page, pageSize, offset := ParsePagination(c)
	sortBy := c.DefaultQuery("sortBy", "created_at")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	items, total, page, pageSize, err := h.svc.ListRestrictions(c.Request.Context(), pr, page, pageSize, offset, sortBy, sortOrder)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list restrictions")
		return
	}
	response.OK(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *ComplianceHandler) GetRestriction(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.GetRestriction(c.Request.Context(), pr, id)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "RESTRICTION_NOT_FOUND", "restriction not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load restriction")
		return
	}
	response.OK(c, dto)
}

type createRestrictionBody struct {
	InstitutionID string         `json:"institutionId" binding:"required"`
	ClientID      string         `json:"clientId" binding:"required"`
	MedicationID  *string        `json:"medicationId"`
	Rule          map[string]any `json:"rule" binding:"required"`
	IsActive      *bool          `json:"isActive"`
}

func (h *ComplianceHandler) CreateRestriction(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body createRestrictionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	active := true
	if body.IsActive != nil {
		active = *body.IsActive
	}
	dto, err := h.svc.CreateRestriction(c.Request.Context(), pr, service.CreateRestrictionInput{
		InstitutionID: body.InstitutionID,
		ClientID:      body.ClientID,
		MedicationID:  body.MedicationID,
		Rule:          body.Rule,
		IsActive:      active,
	}, auditRequestMeta(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create restriction")
		return
	}
	response.OK(c, dto)
}

type patchRestrictionBody struct {
	ClientID     *string        `json:"clientId"`
	MedicationID *string        `json:"medicationId"`
	Rule         map[string]any `json:"rule"`
	IsActive     *bool          `json:"isActive"`
}

func (h *ComplianceHandler) PatchRestriction(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body patchRestrictionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.UpdateRestriction(c.Request.Context(), pr, id, service.UpdateRestrictionInput{
		ClientID:     body.ClientID,
		MedicationID: body.MedicationID,
		Rule:         body.Rule,
		IsActive:     body.IsActive,
	}, auditRequestMeta(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "RESTRICTION_NOT_FOUND", "restriction not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update restriction")
		return
	}
	response.OK(c, dto)
}

type checkPurchaseBody struct {
	InstitutionID            string  `json:"institutionId" binding:"required"`
	ClientID                 string  `json:"clientId" binding:"required"`
	MedicationID             string  `json:"medicationId" binding:"required"`
	IsControlled             bool    `json:"isControlled"`
	PrescriptionAttachmentID *string `json:"prescriptionAttachmentId"`
	PurchaseAt               string  `json:"purchaseAt" binding:"required"`
}

func (h *ComplianceHandler) CheckPurchase(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body checkPurchaseBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	t, err := time.Parse(time.RFC3339, body.PurchaseAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "purchaseAt must be RFC3339")
		return
	}
	result, err := h.svc.CheckPurchase(c.Request.Context(), pr, service.CheckPurchaseInput{
		InstitutionID:            body.InstitutionID,
		ClientID:                 body.ClientID,
		MedicationID:             body.MedicationID,
		IsControlled:             body.IsControlled,
		PrescriptionAttachmentID: body.PrescriptionAttachmentID,
		PurchaseAt:               t,
	})
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "check failed")
		return
	}
	response.OK(c, result)
}

func (h *ComplianceHandler) ListViolations(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	page, pageSize, offset := ParsePagination(c)
	sortBy := c.DefaultQuery("sortBy", "created_at")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	items, total, page, pageSize, err := h.svc.ListViolations(c.Request.Context(), pr, page, pageSize, offset, sortBy, sortOrder)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list violations")
		return
	}
	response.OK(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

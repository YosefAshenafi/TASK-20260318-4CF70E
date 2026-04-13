package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pharmaops/api/internal/middleware"
	"pharmaops/api/internal/repository"
	"pharmaops/api/internal/response"
	"pharmaops/api/internal/service"
)

type CaseHandler struct {
	svc *service.CaseService
}

func NewCaseHandler(svc *service.CaseService) *CaseHandler {
	return &CaseHandler{svc: svc}
}

func (h *CaseHandler) SearchCaseLedger(c *gin.Context) {
	h.listCasesImpl(c)
}

func (h *CaseHandler) ListCases(c *gin.Context) {
	h.listCasesImpl(c)
}

func (h *CaseHandler) listCasesImpl(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	page, pageSize, offset := ParsePagination(c)
	sortBy := c.DefaultQuery("sortBy", "created_at")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	q := c.Query("q")
	status := c.Query("status")
	items, total, page, pageSize, err := h.svc.SearchCaseLedger(c.Request.Context(), pr, page, pageSize, offset, sortBy, sortOrder, q, status)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list cases")
		return
	}
	response.OK(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *CaseHandler) GetCase(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.GetCase(c.Request.Context(), pr, id)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "CASE_NOT_FOUND", "case not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load case")
		return
	}
	response.OK(c, dto)
}

type createCaseBody struct {
	InstitutionID string  `json:"institutionId" binding:"required"`
	DepartmentID  *string `json:"departmentId"`
	TeamID        *string `json:"teamId"`
	CaseType      string  `json:"caseType" binding:"required"`
	Title         string  `json:"title" binding:"required"`
	Description   string  `json:"description" binding:"required"`
	ReportedAt    string  `json:"reportedAt" binding:"required"`
}

func (h *CaseHandler) CreateCase(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body createCaseBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	t, err := time.Parse(time.RFC3339, body.ReportedAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "reportedAt must be RFC3339")
		return
	}
	dto, err := h.svc.CreateCase(c.Request.Context(), pr, service.CreateCaseInput{
		InstitutionID: body.InstitutionID,
		DepartmentID:  body.DepartmentID,
		TeamID:        body.TeamID,
		CaseType:      body.CaseType,
		Title:         body.Title,
		Description:   body.Description,
		ReportedAt:    t,
	})
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if errors.Is(err, service.ErrCaseMandatoryFields) {
		response.Error(c, http.StatusBadRequest, "CASE_MANDATORY_FIELDS_MISSING", "missing required fields")
		return
	}
	if errors.Is(err, service.ErrDuplicateCaseSubmission) {
		response.Error(c, http.StatusConflict, "DUPLICATE_SUBMISSION_BLOCKED", "duplicate submission within 5 minutes")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create case")
		return
	}
	response.OK(c, dto)
}

type patchCaseBody struct {
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	DepartmentID *string `json:"departmentId"`
	TeamID       *string `json:"teamId"`
}

func (h *CaseHandler) PatchCase(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body patchCaseBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.UpdateCase(c.Request.Context(), pr, id, service.UpdateCaseInput{
		Title:        body.Title,
		Description:  body.Description,
		DepartmentID: body.DepartmentID,
		TeamID:       body.TeamID,
	})
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if errors.Is(err, service.ErrCaseMandatoryFields) {
		response.Error(c, http.StatusBadRequest, "CASE_MANDATORY_FIELDS_MISSING", "missing required fields")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "CASE_NOT_FOUND", "case not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update case")
		return
	}
	response.OK(c, dto)
}

type assignCaseBody struct {
	AssigneeUserID string `json:"assigneeUserId" binding:"required"`
}

func (h *CaseHandler) AssignCase(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body assignCaseBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.AssignCase(c.Request.Context(), pr, id, body.AssigneeUserID)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "CASE_NOT_FOUND", "case not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to assign case")
		return
	}
	response.OK(c, dto)
}

type processingRecordBody struct {
	StepCode string  `json:"stepCode" binding:"required"`
	Note     *string `json:"note"`
}

func (h *CaseHandler) PostProcessingRecord(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body processingRecordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	id := c.Param("id")
	uid := c.GetString("userID")
	dto, err := h.svc.AddProcessingRecord(c.Request.Context(), pr, id, uid, body.StepCode, body.Note)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if errors.Is(err, service.ErrCaseMandatoryFields) {
		response.Error(c, http.StatusBadRequest, "CASE_MANDATORY_FIELDS_MISSING", "missing step code")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "CASE_NOT_FOUND", "case not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to add processing record")
		return
	}
	response.OK(c, dto)
}

func (h *CaseHandler) ListProcessingRecords(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("id")
	items, err := h.svc.ListProcessingRecords(c.Request.Context(), pr, id)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "CASE_NOT_FOUND", "case not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list processing records")
		return
	}
	response.OK(c, gin.H{"items": items})
}

type statusTransitionBody struct {
	ToStatus string `json:"toStatus" binding:"required"`
}

func (h *CaseHandler) PostStatusTransition(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body statusTransitionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	id := c.Param("id")
	uid := c.GetString("userID")
	dto, err := h.svc.AddStatusTransition(c.Request.Context(), pr, id, uid, body.ToStatus)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if errors.Is(err, service.ErrInvalidStatusTransition) {
		response.Error(c, http.StatusBadRequest, "INVALID_STATUS_TRANSITION", "transition not allowed")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "CASE_NOT_FOUND", "case not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to record transition")
		return
	}
	response.OK(c, dto)
}

func (h *CaseHandler) ListStatusTransitions(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("id")
	items, err := h.svc.ListStatusTransitions(c.Request.Context(), pr, id)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "CASE_NOT_FOUND", "case not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list transitions")
		return
	}
	response.OK(c, gin.H{"items": items})
}

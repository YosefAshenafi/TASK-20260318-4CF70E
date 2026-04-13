package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"pharmaops/api/internal/middleware"
	"pharmaops/api/internal/repository"
	"pharmaops/api/internal/response"
	"pharmaops/api/internal/service"
)

type FeeHandler struct {
	svc *service.FeeService
}

func NewFeeHandler(svc *service.FeeService) *FeeHandler {
	return &FeeHandler{svc: svc}
}

func (h *FeeHandler) ListFees(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	page, pageSize, offset := ParsePagination(c)
	sortBy := c.DefaultQuery("sortBy", "created_at")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	items, total, page, pageSize, err := h.svc.ListFees(c.Request.Context(), pr, page, pageSize, offset, sortBy, sortOrder)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list fees")
		return
	}
	response.OK(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

type createFeeBody struct {
	InstitutionID string  `json:"institutionId" binding:"required"`
	DepartmentID  *string `json:"departmentId"`
	TeamID        *string `json:"teamId"`
	CaseID        *string `json:"caseId"`
	CandidateID   *string `json:"candidateId"`
	FeeType       string  `json:"feeType" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	Currency      string  `json:"currency"`
	Note          *string `json:"note"`
}

func (h *FeeHandler) CreateFee(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body createFeeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	dto, err := h.svc.CreateFee(c.Request.Context(), pr, service.CreateFeeInput{
		InstitutionID: body.InstitutionID,
		DepartmentID:  body.DepartmentID,
		TeamID:        body.TeamID,
		CaseID:        body.CaseID,
		CandidateID:   body.CandidateID,
		FeeType:       body.FeeType,
		Amount:        body.Amount,
		Currency:      body.Currency,
		Note:          body.Note,
	}, auditRequestMeta(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if errors.Is(err, service.ErrFeeValidation) {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid fee fields")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create fee")
		return
	}
	response.OK(c, dto)
}

type patchFeeBody struct {
	FeeType  *string  `json:"feeType"`
	Amount   *float64 `json:"amount"`
	Currency *string  `json:"currency"`
	Note     *string  `json:"note"`
}

func (h *FeeHandler) PatchFee(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body patchFeeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.UpdateFee(c.Request.Context(), pr, id, service.UpdateFeeInput{
		FeeType:  body.FeeType,
		Amount:   body.Amount,
		Currency: body.Currency,
		Note:     body.Note,
	}, auditRequestMeta(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if errors.Is(err, service.ErrFeeValidation) {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid fee fields")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "FEE_NOT_FOUND", "fee not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update fee")
		return
	}
	response.OK(c, dto)
}

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

type RecruitmentHandler struct {
	svc *service.RecruitmentService
}

func NewRecruitmentHandler(svc *service.RecruitmentService) *RecruitmentHandler {
	return &RecruitmentHandler{svc: svc}
}

func (h *RecruitmentHandler) ListCandidates(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	page, pageSize, offset := ParsePagination(c)
	sortBy := c.DefaultQuery("sortBy", "created_at")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	items, total, page, pageSize, err := h.svc.ListCandidates(c.Request.Context(), pr, page, pageSize, offset, sortBy, sortOrder)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list candidates")
		return
	}
	response.OK(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *RecruitmentHandler) GetCandidate(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.GetCandidate(c.Request.Context(), pr, id)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "CANDIDATE_NOT_FOUND", "candidate not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load candidate")
		return
	}
	response.OK(c, dto)
}

type createCandidateBody struct {
	Name            string   `json:"name" binding:"required"`
	InstitutionID   string   `json:"institutionId" binding:"required"`
	DepartmentID    *string  `json:"departmentId"`
	TeamID          *string  `json:"teamId"`
	ExperienceYears *int     `json:"experienceYears"`
	EducationLevel  *string  `json:"educationLevel"`
	Skills          []string `json:"skills"`
	Tags            []string `json:"tags"`
}

func (h *RecruitmentHandler) CreateCandidate(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body createCandidateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	dto, err := h.svc.CreateCandidate(c.Request.Context(), pr, service.CreateCandidateInput{
		Name:            body.Name,
		InstitutionID:   body.InstitutionID,
		DepartmentID:    body.DepartmentID,
		TeamID:          body.TeamID,
		ExperienceYears: body.ExperienceYears,
		EducationLevel:  body.EducationLevel,
		Skills:          body.Skills,
		Tags:              body.Tags,
	})
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create candidate")
		return
	}
	response.OK(c, dto)
}

type patchCandidateBody struct {
	Name            *string `json:"name"`
	DepartmentID    *string `json:"departmentId"`
	TeamID          *string `json:"teamId"`
	ExperienceYears *int    `json:"experienceYears"`
	EducationLevel  *string `json:"educationLevel"`
}

func (h *RecruitmentHandler) PatchCandidate(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body patchCandidateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.UpdateCandidate(c.Request.Context(), pr, id, service.UpdateCandidateInput{
		Name:            body.Name,
		DepartmentID:    body.DepartmentID,
		TeamID:          body.TeamID,
		ExperienceYears: body.ExperienceYears,
		EducationLevel:  body.EducationLevel,
	})
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "CANDIDATE_NOT_FOUND", "candidate not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update candidate")
		return
	}
	response.OK(c, dto)
}

func (h *RecruitmentHandler) DeleteCandidate(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("id")
	err := h.svc.DeleteCandidate(c.Request.Context(), pr, id)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "CANDIDATE_NOT_FOUND", "candidate not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete candidate")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *RecruitmentHandler) ListPositions(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	page, pageSize, offset := ParsePagination(c)
	sortBy := c.DefaultQuery("sortBy", "created_at")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	items, total, page, pageSize, err := h.svc.ListPositions(c.Request.Context(), pr, page, pageSize, offset, sortBy, sortOrder)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list positions")
		return
	}
	response.OK(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *RecruitmentHandler) GetPosition(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.GetPosition(c.Request.Context(), pr, id)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "POSITION_NOT_FOUND", "position not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load position")
		return
	}
	response.OK(c, dto)
}

type createPositionBody struct {
	InstitutionID string  `json:"institutionId" binding:"required"`
	Title         string  `json:"title" binding:"required"`
	Description   *string `json:"description"`
	Status        string  `json:"status"`
	DepartmentID  *string `json:"departmentId"`
	TeamID        *string `json:"teamId"`
}

func (h *RecruitmentHandler) CreatePosition(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body createPositionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	dto, err := h.svc.CreatePosition(c.Request.Context(), pr, service.CreatePositionInput{
		InstitutionID: body.InstitutionID,
		Title:         body.Title,
		Description:   body.Description,
		Status:        body.Status,
		DepartmentID:  body.DepartmentID,
		TeamID:        body.TeamID,
	})
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create position")
		return
	}
	response.OK(c, dto)
}

type patchPositionBody struct {
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	Status       *string `json:"status"`
	DepartmentID *string `json:"departmentId"`
	TeamID       *string `json:"teamId"`
}

func (h *RecruitmentHandler) PatchPosition(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body patchPositionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.UpdatePosition(c.Request.Context(), pr, id, service.UpdatePositionInput{
		Title:        body.Title,
		Description:  body.Description,
		Status:       body.Status,
		DepartmentID: body.DepartmentID,
		TeamID:       body.TeamID,
	})
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "POSITION_NOT_FOUND", "position not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update position")
		return
	}
	response.OK(c, dto)
}

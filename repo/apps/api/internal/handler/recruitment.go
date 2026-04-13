package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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

func recruitmentPIIAuditOpts(c *gin.Context) service.GetCandidateOpts {
	ip := c.ClientIP()
	return service.GetCandidateOpts{
		OperatorUserID: c.GetString("userID"),
		RequestID:      c.GetString("requestId"),
		RequestSource:  &ip,
	}
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

	search := service.CandidateSearchParams{
		Keyword:        c.Query("keyword"),
		EducationLevel: c.Query("educationLevel"),
	}
	if sk := c.Query("skills"); sk != "" {
		for _, s := range strings.Split(sk, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				search.Skills = append(search.Skills, s)
			}
		}
	}
	if v := c.Query("minExperience"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			search.MinExperience = &n
		}
	}
	if v := c.Query("maxExperience"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			search.MaxExperience = &n
		}
	}

	items, total, page, pageSize, err := h.svc.ListCandidates(c.Request.Context(), pr, page, pageSize, offset, sortBy, sortOrder, search)
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
	dto, err := h.svc.GetCandidate(c.Request.Context(), pr, id, recruitmentPIIAuditOpts(c))
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
	Name            string         `json:"name" binding:"required"`
	InstitutionID   string         `json:"institutionId" binding:"required"`
	DepartmentID    *string        `json:"departmentId"`
	TeamID          *string        `json:"teamId"`
	Phone           *string        `json:"phone"`
	IDNumber        *string        `json:"idNumber"`
	Email           *string        `json:"email"`
	ExperienceYears *int           `json:"experienceYears"`
	EducationLevel  *string        `json:"educationLevel"`
	Skills          []string       `json:"skills"`
	Tags            []string       `json:"tags"`
	CustomFields    map[string]any `json:"customFields"`
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
		Phone:           body.Phone,
		IDNumber:        body.IDNumber,
		Email:           body.Email,
		ExperienceYears: body.ExperienceYears,
		EducationLevel:  body.EducationLevel,
		Skills:          body.Skills,
		Tags:            body.Tags,
		CustomFields:    body.CustomFields,
	}, recruitmentPIIAuditOpts(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if errors.Is(err, service.ErrPIINotConfigured) {
		response.Error(c, http.StatusServiceUnavailable, "PII_KEY_NOT_CONFIGURED", "PII encryption key not configured")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create candidate")
		return
	}
	response.OK(c, dto)
}

type patchCandidateBody struct {
	Name            *string        `json:"name"`
	DepartmentID    *string        `json:"departmentId"`
	TeamID          *string        `json:"teamId"`
	Phone           *string        `json:"phone"`
	IDNumber        *string        `json:"idNumber"`
	Email           *string        `json:"email"`
	ExperienceYears *int           `json:"experienceYears"`
	EducationLevel  *string        `json:"educationLevel"`
	CustomFields    map[string]any `json:"customFields"`
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
		Phone:           body.Phone,
		IDNumber:        body.IDNumber,
		Email:           body.Email,
		ExperienceYears: body.ExperienceYears,
		EducationLevel:  body.EducationLevel,
		CustomFields:    body.CustomFields,
	}, recruitmentPIIAuditOpts(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if errors.Is(err, service.ErrPIINotConfigured) {
		response.Error(c, http.StatusServiceUnavailable, "PII_KEY_NOT_CONFIGURED", "PII encryption key not configured")
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
	err := h.svc.DeleteCandidate(c.Request.Context(), pr, id, auditRequestMeta(c))
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
	}, auditRequestMeta(c))
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
	}, auditRequestMeta(c))
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

func (h *RecruitmentHandler) CreateImportBatch(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	uid := c.GetString("userID")
	var body struct {
		InstitutionID string                      `json:"institutionId" binding:"required"`
		Rows          []service.ImportStagingRow `json:"rows" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	dto, err := h.svc.CreateImportBatch(c.Request.Context(), pr, uid, body.InstitutionID, body.Rows)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "institution not in scope")
		return
	}
	if errors.Is(err, service.ErrImportValidationFailed) {
		response.Error(c, http.StatusBadRequest, "IMPORT_VALIDATION_FAILED", "import validation failed")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create import batch")
		return
	}
	response.OK(c, dto)
}

func (h *RecruitmentHandler) GetImportBatch(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("importId")
	dto, err := h.svc.GetImportBatch(c.Request.Context(), pr, id)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "IMPORT_NOT_FOUND", "import batch not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load import batch")
		return
	}
	response.OK(c, dto)
}

func (h *RecruitmentHandler) CommitImportBatch(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("importId")
	dto, err := h.svc.CommitImportBatch(c.Request.Context(), pr, id)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if errors.Is(err, service.ErrImportValidationFailed) {
		response.Error(c, http.StatusBadRequest, "IMPORT_VALIDATION_FAILED", "import validation failed")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "IMPORT_NOT_FOUND", "import batch not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to commit import")
		return
	}
	response.OK(c, dto)
}

func (h *RecruitmentHandler) ListDuplicateCandidates(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	items, err := h.svc.ListDuplicateGroups(c.Request.Context(), pr)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list duplicates")
		return
	}
	response.OK(c, gin.H{"items": items})
}

func (h *RecruitmentHandler) MergeCandidates(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	uid := c.GetString("userID")
	var body struct {
		BaseCandidateID    string   `json:"baseCandidateId" binding:"required"`
		SourceCandidateIDs []string `json:"sourceCandidateIds" binding:"required"`
		Strategy           string   `json:"strategy"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	err := h.svc.MergeCandidates(c.Request.Context(), pr, uid, service.MergeCandidatesInput{
		BaseCandidateID:    body.BaseCandidateID,
		SourceCandidateIDs: body.SourceCandidateIDs,
		Strategy:           body.Strategy,
	}, auditRequestMeta(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if errors.Is(err, service.ErrMergeValidationFailed) {
		response.Error(c, http.StatusBadRequest, "MERGE_VALIDATION_FAILED", "merge validation failed")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "CANDIDATE_NOT_FOUND", "candidate not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to merge candidates")
		return
	}
	response.OK(c, gin.H{"merged": true})
}

func (h *RecruitmentHandler) ListMergeHistory(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	page, pageSize, offset := ParsePagination(c)
	items, total, page, pageSize, err := h.svc.ListMergeHistory(c.Request.Context(), pr, page, pageSize, offset)
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list merge history")
		return
	}
	response.OK(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

type matchPairBody struct {
	CandidateID string `json:"candidateId" binding:"required"`
	PositionID  string `json:"positionId" binding:"required"`
}

func (h *RecruitmentHandler) MatchCandidateToPosition(c *gin.Context) {
	h.matchPair(c, true)
}

func (h *RecruitmentHandler) MatchPositionToCandidate(c *gin.Context) {
	h.matchPair(c, false)
}

func (h *RecruitmentHandler) matchPair(c *gin.Context, candidateFirst bool) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	var body matchPairBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	var dto *service.MatchScoreDTO
	var err error
	if candidateFirst {
		dto, err = h.svc.MatchCandidateToPosition(c.Request.Context(), pr, body.CandidateID, body.PositionID)
	} else {
		dto, err = h.svc.MatchPositionToCandidate(c.Request.Context(), pr, body.PositionID, body.CandidateID)
	}
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", "candidate or position not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to compute match")
		return
	}
	response.OK(c, dto)
}

func parseSimilarLimit(c *gin.Context) int {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > 50 {
		limit = 50
	}
	return limit
}

func (h *RecruitmentHandler) SimilarCandidates(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("candidateId")
	items, err := h.svc.SimilarCandidates(c.Request.Context(), pr, id, parseSimilarLimit(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "CANDIDATE_NOT_FOUND", "candidate not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load recommendations")
		return
	}
	response.OK(c, gin.H{"items": items})
}

func (h *RecruitmentHandler) SimilarPositions(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	id := c.Param("positionId")
	items, err := h.svc.SimilarPositions(c.Request.Context(), pr, id, parseSimilarLimit(c))
	if errors.Is(err, service.ErrForbiddenScope) {
		response.Error(c, http.StatusForbidden, "FORBIDDEN_SCOPE", "no institution scope")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "POSITION_NOT_FOUND", "position not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load recommendations")
		return
	}
	response.OK(c, gin.H{"items": items})
}

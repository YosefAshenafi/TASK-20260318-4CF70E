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

type RbacHandler struct {
	svc *service.RbacService
}

func NewRbacHandler(svc *service.RbacService) *RbacHandler {
	return &RbacHandler{svc: svc}
}

func (h *RbacHandler) ListUsers(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	items, err := h.svc.ListUsers(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list users")
		return
	}
	response.OK(c, gin.H{"items": items})
}

type createUserBody struct {
	Username    string   `json:"username" binding:"required"`
	Password    string   `json:"password" binding:"required"`
	DisplayName string   `json:"displayName" binding:"required"`
	IsActive    *bool    `json:"isActive"`
	RoleIDs     []string `json:"roleIds"`
}

func (h *RbacHandler) CreateUser(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	var body createUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	active := true
	if body.IsActive != nil {
		active = *body.IsActive
	}
	if body.RoleIDs == nil {
		body.RoleIDs = []string{}
	}
	dto, err := h.svc.CreateUser(c.Request.Context(), service.CreateUserInput{
		Username:    body.Username,
		Password:    body.Password,
		DisplayName: body.DisplayName,
		IsActive:    active,
		RoleIDs:     body.RoleIDs,
	})
	if errors.Is(err, service.ErrRbacValidation) {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid user fields or username taken")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create user")
		return
	}
	response.OK(c, dto)
}

func (h *RbacHandler) GetUser(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	id := c.Param("id")
	dto, err := h.svc.GetUser(c.Request.Context(), id)
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load user")
		return
	}
	response.OK(c, dto)
}

type patchUserBody struct {
	DisplayName *string   `json:"displayName"`
	IsActive    *bool     `json:"isActive"`
	Password    *string   `json:"password"`
	RoleIDs     *[]string `json:"roleIds"` // nil = leave roles unchanged; non-nil = replace set (empty clears)
}

func (h *RbacHandler) PatchUser(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	id := c.Param("id")
	var body patchUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	dto, err := h.svc.UpdateUser(c.Request.Context(), id, service.UpdateUserInput{
		DisplayName: body.DisplayName,
		IsActive:    body.IsActive,
		Password:    body.Password,
		RoleIDs:     body.RoleIDs,
	})
	if errors.Is(err, service.ErrRbacValidation) {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid user fields")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update user")
		return
	}
	response.OK(c, dto)
}

type setUserScopesBody struct {
	ScopeIDs []string `json:"scopeIds"`
}

func (h *RbacHandler) SetUserScopes(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	id := c.Param("id")
	var body setUserScopesBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	if body.ScopeIDs == nil {
		body.ScopeIDs = []string{}
	}
	err := h.svc.SetUserScopes(c.Request.Context(), id, body.ScopeIDs)
	if errors.Is(err, service.ErrRbacValidation) {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "unknown scope id")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update scopes")
		return
	}
	response.OK(c, gin.H{"ok": true})
}

type createRoleBody struct {
	Slug        string  `json:"slug" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

func (h *RbacHandler) CreateRole(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	var body createRoleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	dto, err := h.svc.CreateRole(c.Request.Context(), service.CreateRoleInput{
		Slug:        body.Slug,
		Name:        body.Name,
		Description: body.Description,
	})
	if errors.Is(err, service.ErrRbacValidation) {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid role fields or duplicate slug")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create role")
		return
	}
	response.OK(c, dto)
}

func (h *RbacHandler) ListRoles(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	items, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list roles")
		return
	}
	response.OK(c, gin.H{"items": items})
}

func (h *RbacHandler) GetRole(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	id := c.Param("id")
	dto, err := h.svc.GetRole(c.Request.Context(), id)
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "ROLE_NOT_FOUND", "role not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load role")
		return
	}
	response.OK(c, dto)
}

func (h *RbacHandler) ListPermissions(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	items, err := h.svc.ListPermissions(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list permissions")
		return
	}
	response.OK(c, gin.H{"items": items})
}

type patchRoleBody struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (h *RbacHandler) PatchRole(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	var body patchRoleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	id := c.Param("id")
	dto, err := h.svc.UpdateRole(c.Request.Context(), id, service.UpdateRoleInput{
		Name:        body.Name,
		Description: body.Description,
	})
	if errors.Is(err, service.ErrRbacValidation) {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid role fields")
		return
	}
	if repository.IsNotFound(err) {
		response.Error(c, http.StatusNotFound, "ROLE_NOT_FOUND", "role not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update role")
		return
	}
	response.OK(c, dto)
}

type setRolePermissionsBody struct {
	PermissionIDs []string `json:"permissionIds"`
}

func (h *RbacHandler) SetRolePermissions(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	var body setRolePermissionsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	if body.PermissionIDs == nil {
		body.PermissionIDs = []string{}
	}
	id := c.Param("id")
	if err := h.svc.SetRolePermissions(c.Request.Context(), id, body.PermissionIDs); err != nil {
		if repository.IsNotFound(err) {
			response.Error(c, http.StatusNotFound, "ROLE_NOT_FOUND", "role not found")
			return
		}
		if errors.Is(err, service.ErrRbacValidation) {
			response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "unknown permission id")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update permissions")
		return
	}
	response.OK(c, gin.H{"ok": true})
}

type createScopeBody struct {
	ScopeKey      string  `json:"scopeKey" binding:"required"`
	InstitutionID string  `json:"institutionId" binding:"required"`
	DepartmentID  *string `json:"departmentId"`
	TeamID        *string `json:"teamId"`
}

func (h *RbacHandler) CreateScope(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	var body createScopeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	dto, err := h.svc.CreateDataScope(c.Request.Context(), service.CreateDataScopeInput{
		ScopeKey:      body.ScopeKey,
		InstitutionID: body.InstitutionID,
		DepartmentID:  body.DepartmentID,
		TeamID:        body.TeamID,
	})
	if errors.Is(err, service.ErrRbacValidation) {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid scope fields or duplicate scope key")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create scope")
		return
	}
	response.OK(c, dto)
}

func (h *RbacHandler) ListScopes(c *gin.Context) {
	pr, ok := middleware.GetPrincipal(c)
	if !ok || pr == nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "missing principal")
		return
	}
	_ = pr
	items, err := h.svc.ListScopes(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list scopes")
		return
	}
	response.OK(c, gin.H{"items": items})
}

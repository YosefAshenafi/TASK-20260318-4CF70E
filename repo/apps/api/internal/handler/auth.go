package handler

import (
	"errors"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/middleware"
	"pharmaops/api/internal/oplog"
	"pharmaops/api/internal/repository"
	"pharmaops/api/internal/response"
	"pharmaops/api/internal/service"
)

type AuthHandler struct {
	auth   *service.AuthService
	access *service.AccessService
	users  *repository.UserRepository
}

func NewAuthHandler(auth *service.AuthService, access *service.AccessService, users *repository.UserRepository) *AuthHandler {
	return &AuthHandler{auth: auth, access: access, users: users}
}

type loginBody struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type meScopeDTO struct {
	ID             string  `json:"id"`
	ScopeKey       string  `json:"scopeKey"`
	InstitutionID  string  `json:"institutionId"`
	DepartmentID   *string `json:"departmentId,omitempty"`
	TeamID         *string `json:"teamId,omitempty"`
}

type meUser struct {
	ID          string        `json:"id"`
	Username    string        `json:"username"`
	Roles       []string      `json:"roles"`
	Permissions []string      `json:"permissions"`
	Scopes      []meScopeDTO  `json:"scopes"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var body loginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")
	var ipPtr, uaPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	if ua != "" {
		uaPtr = &ua
	}
	rid := c.GetString("requestId")
	ls, err := h.auth.Login(c.Request.Context(), body.Username, body.Password, ipPtr, uaPtr)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			oplog.AuthFailure(rid, ip, body.Username, "invalid_credentials")
			response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "invalid username or password")
		case errors.Is(err, service.ErrPasswordTooShort):
			oplog.AuthFailure(rid, ip, body.Username, "password_too_short")
			response.Error(c, http.StatusBadRequest, "AUTH_PASSWORD_TOO_SHORT", "password must be at least 8 characters")
		case errors.Is(err, service.ErrAccountDisabled):
			oplog.AuthFailure(rid, ip, body.Username, "account_disabled")
			response.Error(c, http.StatusForbidden, "AUTH_ACCOUNT_DISABLED", "account is disabled")
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "login failed")
		}
		return
	}
	oplog.AuthSuccess(rid, ls.UserID, ip)
	pr, err := h.access.LoadPrincipal(c.Request.Context(), ls.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load user")
		return
	}
	response.OK(c, gin.H{
		"token":     ls.Token,
		"expiresAt": ls.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
		"user": gin.H{
			"id":       ls.UserID,
			"username": ls.Username,
			"roles":    pr.RoleSlugs,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	uid := c.GetString("userID")
	if uid == "" {
		response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "not authenticated")
		return
	}
	u, err := h.users.FindByID(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "user not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load user")
		return
	}
	roles := []string{}
	var perms []string
	var scopeDTOs []meScopeDTO
	if pr, ok := middleware.GetPrincipal(c); ok && pr != nil {
		roles = pr.RoleSlugs
		perms = permissionCodesSorted(pr)
		for _, s := range pr.Scopes {
			scopeDTOs = append(scopeDTOs, meScopeDTO{
				ID:            s.ID,
				ScopeKey:      s.ScopeKey,
				InstitutionID: s.InstitutionID,
				DepartmentID:  s.DepartmentID,
				TeamID:        s.TeamID,
			})
		}
	}
	response.OK(c, meUser{
		ID:          u.ID,
		Username:    u.Username,
		Roles:       roles,
		Permissions: perms,
		Scopes:      scopeDTOs,
	})
}

func permissionCodesSorted(pr *access.Principal) []string {
	if pr == nil {
		return nil
	}
	out := make([]string, 0, len(pr.PermissionSet))
	for code := range pr.PermissionSet {
		out = append(out, code)
	}
	sort.Strings(out)
	return out
}

func (h *AuthHandler) Logout(c *gin.Context) {
	raw := middleware.BearerToken(c.GetHeader("Authorization"))
	if raw != "" {
		_ = h.auth.Logout(c.Request.Context(), raw)
	}
	response.OK(c, gin.H{"loggedOut": true})
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pharmaops/api/internal/response"
)

type HealthHandler struct {
	db           *gorm.DB
	checkToken   string
}

func NewHealthHandler(db *gorm.DB, healthCheckToken string) *HealthHandler {
	return &HealthHandler{db: db, checkToken: healthCheckToken}
}

func (h *HealthHandler) Get(c *gin.Context) {
	if h.checkToken == "" {
		response.Error(c, http.StatusServiceUnavailable, "UNAVAILABLE", "health check token is not configured")
		return
	}
	if c.GetHeader("X-Internal-Health-Token") != h.checkToken {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid health token")
		return
	}
	sqlDB, err := h.db.DB()
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, "UNAVAILABLE", "database connection unavailable")
		return
	}
	if err := sqlDB.Ping(); err != nil {
		response.Error(c, http.StatusServiceUnavailable, "UNAVAILABLE", "database ping failed")
		return
	}
	response.OK(c, gin.H{
		"status":   "ok",
		"database": "connected",
	})
}

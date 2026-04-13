package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pharmaops/api/internal/response"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Get(c *gin.Context) {
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

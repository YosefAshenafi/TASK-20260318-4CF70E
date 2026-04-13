package httpserver

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pharmaops/api/internal/config"
	"pharmaops/api/internal/handler"
	"pharmaops/api/internal/middleware"
	"pharmaops/api/internal/repository"
	"pharmaops/api/internal/service"
)

type Server struct {
	cfg config.Config
	db  *gorm.DB
}

func New(cfg config.Config, db *gorm.DB) *Server {
	return &Server{cfg: cfg, db: db}
}

func (s *Server) Run() error {
	if s.cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())

	userRepo := repository.NewUserRepository(s.db)
	sessRepo := repository.NewSessionRepository(s.db)
	accessRepo := repository.NewAccessRepository(s.db)
	accessSvc := service.NewAccessService(accessRepo)
	authSvc := service.NewAuthService(s.cfg, userRepo, sessRepo)
	authH := handler.NewAuthHandler(authSvc, userRepo)
	healthH := handler.NewHealthHandler(s.db)

	r.GET("/healthz", func(c *gin.Context) {
		c.Status(200)
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthH.Get)
		v1.POST("/auth/login", authH.Login)
		v1.POST("/auth/logout", authH.Logout)
		authz := v1.Group("")
		authz.Use(middleware.SessionAuth(authSvc))
		authz.Use(middleware.AccessContext(accessSvc))
		{
			authz.GET("/auth/me", authH.Me)
		}
	}

	addr := s.cfg.HTTPAddr
	if err := r.Run(addr); err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	return nil
}

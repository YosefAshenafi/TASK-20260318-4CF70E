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
	recRepo := repository.NewRecruitmentRepository(s.db)
	recSvc := service.NewRecruitmentService(recRepo)
	recH := handler.NewRecruitmentHandler(recSvc)
	complianceRepo := repository.NewComplianceRepository(s.db)
	complianceSvc := service.NewComplianceService(complianceRepo)
	complianceH := handler.NewComplianceHandler(complianceSvc)

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

			authz.GET("/recruitment/candidates", middleware.RequirePermission("recruitment.view"), recH.ListCandidates)
			authz.GET("/recruitment/candidates/:id", middleware.RequirePermission("recruitment.view"), recH.GetCandidate)
			authz.POST("/recruitment/candidates", middleware.RequirePermission("recruitment.manage"), recH.CreateCandidate)
			authz.PATCH("/recruitment/candidates/:id", middleware.RequirePermission("recruitment.manage"), recH.PatchCandidate)
			authz.DELETE("/recruitment/candidates/:id", middleware.RequirePermission("recruitment.manage"), recH.DeleteCandidate)

			authz.GET("/recruitment/positions", middleware.RequirePermission("recruitment.view"), recH.ListPositions)
			authz.GET("/recruitment/positions/:id", middleware.RequirePermission("recruitment.view"), recH.GetPosition)
			authz.POST("/recruitment/positions", middleware.RequirePermission("recruitment.manage"), recH.CreatePosition)
			authz.PATCH("/recruitment/positions/:id", middleware.RequirePermission("recruitment.manage"), recH.PatchPosition)

			authz.GET("/compliance/qualifications/expiring", middleware.RequirePermission("compliance.view"), complianceH.ListExpiringQualifications)
			authz.GET("/compliance/qualifications", middleware.RequirePermission("compliance.view"), complianceH.ListQualifications)
			authz.GET("/compliance/qualifications/:id", middleware.RequirePermission("compliance.view"), complianceH.GetQualification)
			authz.POST("/compliance/qualifications", middleware.RequirePermission("compliance.manage"), complianceH.CreateQualification)
			authz.PATCH("/compliance/qualifications/:id", middleware.RequirePermission("compliance.manage"), complianceH.PatchQualification)
			authz.POST("/compliance/qualifications/:id/activate", middleware.RequirePermission("compliance.manage"), complianceH.ActivateQualification)
			authz.POST("/compliance/qualifications/:id/deactivate", middleware.RequirePermission("compliance.manage"), complianceH.DeactivateQualification)
			authz.POST("/compliance/jobs/qualifications/run", middleware.RequirePermission("compliance.manage"), complianceH.RunQualificationJob)

			authz.GET("/compliance/restrictions/violations", middleware.RequirePermission("compliance.view"), complianceH.ListViolations)
			authz.POST("/compliance/restrictions/check-purchase", middleware.RequirePermission("compliance.manage"), complianceH.CheckPurchase)
			authz.GET("/compliance/restrictions", middleware.RequirePermission("compliance.view"), complianceH.ListRestrictions)
			authz.GET("/compliance/restrictions/:id", middleware.RequirePermission("compliance.view"), complianceH.GetRestriction)
			authz.POST("/compliance/restrictions", middleware.RequirePermission("compliance.manage"), complianceH.CreateRestriction)
			authz.PATCH("/compliance/restrictions/:id", middleware.RequirePermission("compliance.manage"), complianceH.PatchRestriction)
		}
	}

	addr := s.cfg.HTTPAddr
	if err := r.Run(addr); err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	return nil
}

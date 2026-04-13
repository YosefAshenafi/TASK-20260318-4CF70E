package httpserver

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/config"
	cryptopii "pharmaops/api/internal/crypto/pii"
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
	authH := handler.NewAuthHandler(authSvc, accessSvc, userRepo)
	healthH := handler.NewHealthHandler(s.db, s.cfg.HealthCheckToken)
	auditRepo := repository.NewAuditRepository(s.db)
	auditSvc := service.NewAuditService(auditRepo)
	recRepo := repository.NewRecruitmentRepository(s.db)
	piiCipher, err := cryptopii.NewCipherFromHex(s.cfg.PIIAESKeyHex)
	if err != nil {
		return fmt.Errorf("PII AES key: %w", err)
	}
	recSvc := service.NewRecruitmentService(recRepo, piiCipher, auditSvc)
	recH := handler.NewRecruitmentHandler(recSvc)
	complianceRepo := repository.NewComplianceRepository(s.db)
	fileRepoForCompliance := repository.NewFileRepository(s.db)
	complianceSvc := service.NewComplianceService(complianceRepo, auditSvc, service.WithFileRepository(fileRepoForCompliance))
	complianceH := handler.NewComplianceHandler(complianceSvc)
	caseRepo := repository.NewCaseRepository(s.db)
	caseSvc := service.NewCaseService(caseRepo, auditSvc)
	caseH := handler.NewCaseHandler(caseSvc)
	auditExportDir := filepath.Join(s.cfg.FileStorageRoot, "audit-exports")
	_ = os.MkdirAll(auditExportDir, 0o700)
	auditH := handler.NewAuditHandler(auditSvc, auditExportDir)
	rbacRepo := repository.NewRbacRepository(s.db)
	rbacSvc := service.NewRbacService(userRepo, rbacRepo, auditSvc)
	rbacH := handler.NewRbacHandler(rbacSvc)
	fileRepo := repository.NewFileRepository(s.db)
	fileSvc := service.NewFileService(s.cfg.FileStorageRoot, fileRepo, caseRepo, auditSvc)
	fileH := handler.NewFileHandler(fileSvc)
	if s.cfg.FileStorageRoot != "" {
		_ = os.MkdirAll(s.cfg.FileStorageRoot, 0o700)
	}

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
			authz.POST("/recruitment/candidates", middleware.RequirePermission("recruitment.manage"), recH.CreateCandidate)
			authz.POST("/recruitment/candidates/imports", middleware.RequirePermission("recruitment.manage"), recH.CreateImportBatch)
			authz.GET("/recruitment/candidates/imports/:importId", middleware.RequirePermission("recruitment.view"), recH.GetImportBatch)
			authz.POST("/recruitment/candidates/imports/:importId/commit", middleware.RequirePermission("recruitment.manage"), recH.CommitImportBatch)
			authz.GET("/recruitment/candidates/duplicates", middleware.RequirePermission("recruitment.view"), recH.ListDuplicateCandidates)
			authz.POST("/recruitment/candidates/merge", middleware.RequirePermission("recruitment.manage"), recH.MergeCandidates)
			authz.GET("/recruitment/candidates/merge-history", middleware.RequirePermission("recruitment.view"), recH.ListMergeHistory)
			authz.GET("/recruitment/candidates/:id", middleware.RequirePermission("recruitment.view"), recH.GetCandidate)
			authz.PATCH("/recruitment/candidates/:id", middleware.RequirePermission("recruitment.manage"), recH.PatchCandidate)
			authz.DELETE("/recruitment/candidates/:id", middleware.RequirePermission("recruitment.manage"), recH.DeleteCandidate)

			authz.POST("/recruitment/match/candidate-to-position", middleware.RequirePermission("recruitment.view"), recH.MatchCandidateToPosition)
			authz.POST("/recruitment/match/position-to-candidate", middleware.RequirePermission("recruitment.view"), recH.MatchPositionToCandidate)
			authz.GET("/recruitment/recommendations/similar-candidates/:candidateId", middleware.RequirePermission("recruitment.view"), recH.SimilarCandidates)
			authz.GET("/recruitment/recommendations/similar-positions/:positionId", middleware.RequirePermission("recruitment.view"), recH.SimilarPositions)

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

			authz.GET("/case-ledger/search", middleware.RequirePermission("cases.view"), caseH.SearchCaseLedger)
			authz.GET("/cases", middleware.RequirePermission("cases.view"), caseH.ListCases)
			authz.POST("/cases", middleware.RequirePermission("cases.manage"), caseH.CreateCase)
			authz.GET("/cases/:id", middleware.RequirePermission("cases.view"), caseH.GetCase)
			authz.PATCH("/cases/:id", middleware.RequirePermission("cases.manage"), caseH.PatchCase)
			authz.POST("/cases/:id/assign", middleware.RequirePermission("cases.manage"), caseH.AssignCase)
			authz.GET("/cases/:id/processing-records", middleware.RequirePermission("cases.view"), caseH.ListProcessingRecords)
			authz.POST("/cases/:id/processing-records", middleware.RequirePermission("cases.manage"), caseH.PostProcessingRecord)
			authz.GET("/cases/:id/status-transitions", middleware.RequirePermission("cases.view"), caseH.ListStatusTransitions)
			authz.POST("/cases/:id/status-transitions", middleware.RequirePermission("cases.manage"), caseH.PostStatusTransition)

			authz.GET("/audit/logs", middleware.RequirePermission("audit.view"), auditH.ListLogs)
			authz.POST("/audit/logs/export", middleware.RequirePermission("audit.view"), auditH.RequestExport)
			authz.GET("/audit/logs/export/:exportId", middleware.RequirePermission("audit.view"), auditH.GetExport)
			authz.GET("/audit/logs/export/:exportId/download", middleware.RequirePermission("audit.view"), auditH.DownloadExport)

			authz.GET("/files", middleware.RequirePermission("files.view"), fileH.ListFiles)
			authz.POST("/files/uploads/init", middleware.RequirePermission("files.manage"), fileH.InitUpload)
			authz.PUT("/files/uploads/:uploadId/chunks/:chunkIndex", middleware.RequirePermission("files.manage"), fileH.PutChunk)
			authz.POST("/files/uploads/:uploadId/complete", middleware.RequirePermission("files.manage"), fileH.CompleteUpload)
			authz.GET("/files/uploads/:uploadId", middleware.RequirePermission("files.view"), fileH.GetUpload)
			authz.GET("/files/:fileId", middleware.RequirePermission("files.view"), fileH.GetFile)
			authz.GET("/files/:fileId/download", middleware.RequirePermission("files.view"), fileH.DownloadFile)
			authz.POST("/files/:fileId/link", middleware.RequirePermission("files.manage"), fileH.LinkFile)

			authz.GET("/users", middleware.RequirePermission("system.rbac"), rbacH.ListUsers)
			authz.POST("/users", middleware.RequirePermission("system.rbac"), rbacH.CreateUser)
			authz.GET("/users/:id", middleware.RequirePermission("system.rbac"), rbacH.GetUser)
			authz.PATCH("/users/:id", middleware.RequirePermission("system.rbac"), rbacH.PatchUser)
			authz.POST("/users/:id/scopes", middleware.RequirePermission("system.rbac"), rbacH.SetUserScopes)
			authz.GET("/roles", middleware.RequirePermission("system.rbac"), rbacH.ListRoles)
			authz.POST("/roles", middleware.RequirePermission("system.rbac"), rbacH.CreateRole)
			authz.GET("/permissions", middleware.RequirePermission("system.rbac"), rbacH.ListPermissions)
			authz.GET("/scopes", middleware.RequirePermission("system.rbac"), rbacH.ListScopes)
			authz.POST("/scopes", middleware.RequirePermission("system.rbac"), rbacH.CreateScope)
			authz.GET("/roles/:id", middleware.RequirePermission("system.rbac"), rbacH.GetRole)
			authz.PATCH("/roles/:id", middleware.RequirePermission("system.rbac"), rbacH.PatchRole)
			authz.POST("/roles/:id/permissions", middleware.RequirePermission("system.rbac"), rbacH.SetRolePermissions)
		}
	}

	go runQualificationExpirationScheduler(complianceSvc)

	addr := s.cfg.HTTPAddr
	if err := r.Run(addr); err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	return nil
}

func runQualificationExpirationScheduler(svc *service.ComplianceService) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	runOnce := func() {
		ctx := context.Background()
		systemPrincipal := &access.Principal{
			PermissionSet: map[string]struct{}{access.PermissionFullAccess: {}},
			Scopes:        []access.Scope{{InstitutionID: "*"}},
		}
		n, err := svc.RunQualificationExpirationJob(ctx, systemPrincipal, service.AuditRequestMeta{OperatorUserID: "system"})
		if err != nil {
			log.Printf("[expiration-scheduler] error: %v", err)
		} else if n > 0 {
			log.Printf("[expiration-scheduler] deactivated %d expired qualifications", n)
		}
	}
	runOnce()
	for range ticker.C {
		runOnce()
	}
}

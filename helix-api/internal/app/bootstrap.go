package app

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"helix-api/internal/config"
	"helix-api/internal/handler"
	"helix-api/internal/pkg/token"
	"helix-api/internal/repo"
	"helix-api/internal/repo/memory"
	mysqlrepo "helix-api/internal/repo/mysql"
	"helix-api/internal/service"
)

func Bootstrap(cfg config.Config) *gin.Engine {
	var (
		userRepo         repo.UserRepo         = memory.NewUserRepoMemory(cfg.AdminUser, cfg.AdminPass)
		auditRepo        repo.AuditRepo        = memory.NewAuditRepoMemory()
		systemConfigRepo repo.SystemConfigRepo = memory.NewSystemConfigRepoMemory()
		db               *sqlx.DB
	)
	tokens := token.NewManager(cfg.JwtSecret, cfg.JwtExpires)

	// If DB_NAME is configured, use MySQL repos (fail fast on any error)
	if cfg.DBName != "" {
		mysqlDB, err := NewDB(cfg)
		if err != nil {
			panic(err)
		}
		if err := EnsureTables(mysqlDB); err != nil {
			panic(err)
		}
		db = mysqlDB
		userRepo = mysqlrepo.NewUserRepoMySQL(mysqlDB)
		auditRepo = mysqlrepo.NewAuditRepoMySQL(mysqlDB)
		systemConfigRepo = mysqlrepo.NewSystemConfigRepoMySQL(mysqlDB)
	}

	authService := service.NewAuthService(userRepo, tokens)
	auditService := service.NewAuditService(auditRepo)
	systemConfigService := service.NewSystemConfigService(systemConfigRepo)

	authHandler := handler.NewAuthHandler(authService, auditService)
	healthHandler := handler.NewHealthHandler()
	systemConfigHandler := handler.NewSystemConfigHandler(systemConfigService, auditService)
	auditHandler := handler.NewAuditHandler(auditService)
	permissionHandler := handler.NewPermissionHandler()

	var adminUserHandler *handler.AdminUserHandler
	var adminRoleHandler *handler.AdminRoleHandler
	if db != nil {
		adminUserHandler = handler.NewAdminUserHandler(db, auditService)
		adminRoleHandler = handler.NewAdminRoleHandler(db, auditService)
	}

	return NewRouter(cfg, authHandler, healthHandler, systemConfigHandler, auditHandler, adminUserHandler, adminRoleHandler, permissionHandler, tokens)
}

package app

import (
	"github.com/gin-gonic/gin"

	"helix-api/internal/config"
	"helix-api/internal/domain"
	"helix-api/internal/handler"
	"helix-api/internal/middleware"
	"helix-api/internal/pkg/token"
)

func NewRouter(
	cfg config.Config,
	auth *handler.AuthHandler,
	health *handler.HealthHandler,
	systemConfigs *handler.SystemConfigHandler,
	audit *handler.AuditHandler,
	adminUsers *handler.AdminUserHandler,
	adminRoles *handler.AdminRoleHandler,
	permissions *handler.PermissionHandler,
	tokens *token.Manager,
) *gin.Engine {
	if cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recover())
	engine.Use(middleware.CORS())

	engine.GET("/healthz", health.Healthz)
	engine.GET("/version", health.Version)

	adminAuth := engine.Group("/admin/auth")
	adminAuth.POST("/login", auth.Login)
	adminAuth.GET("/me", middleware.JWTAuth(tokens), auth.Me)
	adminAuth.POST("/change-password", middleware.JWTAuth(tokens), auth.ChangePassword)

	adminSystem := engine.Group("/admin/system", middleware.JWTAuth(tokens))
	adminSystem.GET("/configs", middleware.RequireAny(domain.PermAdminManage, domain.PermSystemConfigRead), systemConfigs.List)
	adminSystem.PUT("/configs/:key", middleware.RequireAny(domain.PermAdminManage, domain.PermSystemConfigWrite), systemConfigs.Upsert)

	adminAudit := engine.Group("/admin/audit", middleware.JWTAuth(tokens))
	adminAudit.GET("/logs", middleware.RequireAny(domain.PermAdminManage, domain.PermAuditRead), audit.ListLogs)

	if permissions != nil {
		adminPerms := engine.Group(
			"/admin/permissions",
			middleware.JWTAuth(tokens),
			middleware.RequireAny(domain.PermAdminManage, domain.PermAdminRoleRead),
		)
		adminPerms.GET("", permissions.List)
	}

	// Admin management: require admin.manage permission
	if adminUsers != nil {
		adminUsersGroup := engine.Group(
			"/admin/admin-users",
			middleware.JWTAuth(tokens),
		)
		adminUsersGroup.GET("", middleware.RequireAny(domain.PermAdminManage, domain.PermAdminUserRead), adminUsers.List)
		adminUsersGroup.POST("", middleware.RequireAny(domain.PermAdminManage, domain.PermAdminUserWrite), adminUsers.Create)
		adminUsersGroup.POST("/:id/enable", middleware.RequireAny(domain.PermAdminManage, domain.PermAdminUserWrite), adminUsers.Enable)
		adminUsersGroup.POST("/:id/reset-password", middleware.RequireAny(domain.PermAdminManage, domain.PermAdminUserWrite), adminUsers.ResetPassword)
		adminUsersGroup.GET("/:id/roles", middleware.RequireAny(domain.PermAdminManage, domain.PermAdminUserRead), adminUsers.GetRoles)
		adminUsersGroup.POST("/:id/roles", middleware.RequireAny(domain.PermAdminManage, domain.PermAdminUserWrite), adminUsers.SetRoles)
	}

	if adminRoles != nil {
		adminRolesGroup := engine.Group(
			"/admin/admin-roles",
			middleware.JWTAuth(tokens),
		)
		adminRolesGroup.GET("", middleware.RequireAny(domain.PermAdminManage, domain.PermAdminRoleRead), adminRoles.List)
	}

	return engine
}

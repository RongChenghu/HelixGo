package domain

// Permission describes a single permission code and its human-readable description.
type Permission struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

const (
	// Super permission: implies all other permissions when present in claims.
	PermAdminManage = "admin.manage"

	PermAdminUserRead  = "admin.user.read"
	PermAdminUserWrite = "admin.user.write"

	PermAdminRoleRead = "admin.role.read"

	PermSystemConfigRead  = "system.config.read"
	PermSystemConfigWrite = "system.config.write"

	PermAuditRead = "audit.read"
)

// AllPermissions returns the full list of permissions that the API currently understands.
func AllPermissions() []Permission {
	return []Permission{
		{Code: PermAdminManage, Description: "Super admin permission (all access)"},

		{Code: PermAdminUserRead, Description: "View admin users"},
		{Code: PermAdminUserWrite, Description: "Manage admin users (create/enable/reset roles)"},

		{Code: PermAdminRoleRead, Description: "View admin roles and their permissions"},

		{Code: PermSystemConfigRead, Description: "Read system configs"},
		{Code: PermSystemConfigWrite, Description: "Update system configs"},

		{Code: PermAuditRead, Description: "Read audit logs"},
	}
}

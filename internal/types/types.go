package types

type UserRole string

const (
	RoleSuperAdmin UserRole = "superadmin"
	RoleOwner      UserRole = "owner"
	RoleOperator   UserRole = "operator"
	RoleViewer     UserRole = "viewer"
)

type LicenseStatus string

const (
	LicenseStatusActive  LicenseStatus = "active"
	LicenseStatusExpired LicenseStatus = "expired"
)

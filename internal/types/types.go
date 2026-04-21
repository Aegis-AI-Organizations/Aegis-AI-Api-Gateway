package types

type UserRole string

const (
	RoleSuperAdmin UserRole = "superadmin"
	RoleAdmin      UserRole = "admin"
	RoleBillingAegis UserRole = "billing_aegis"
	RoleTechnicien UserRole = "technicien"
	RoleSupport    UserRole = "support"
	RoleCommercial UserRole = "commercial"
	RoleOwner      UserRole = "owner"
	RoleBillingClient UserRole = "billing_client"
	RoleOperateur  UserRole = "operateur"
	RoleViewer     UserRole = "viewer"
)

type LicenseStatus string

const (
	LicenseStatusActive  LicenseStatus = "active"
	LicenseStatusExpired LicenseStatus = "expired"
)

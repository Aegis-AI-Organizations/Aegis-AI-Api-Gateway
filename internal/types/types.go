package types

type UserRole string

const (
	RoleSuperAdmin    UserRole = "superadmin"
	RoleAdmin         UserRole = "admin"
	RoleBillingAegis  UserRole = "billing_aegis"
	RoleTechnicien    UserRole = "technicien"
	RoleSupport       UserRole = "support"
	RoleOwner         UserRole = "owner"
	RoleBillingClient UserRole = "billing_client"
	RoleOperateur     UserRole = "operateur"
	RoleViewer        UserRole = "viewer"
)

type LicenseStatus string

const (
	LicenseStatusActive  LicenseStatus = "active"
	LicenseStatusExpired LicenseStatus = "expired"
)

type ContextKey string

const (
	UserIDKey        ContextKey = "user_id"
	CompanyIDKey     ContextKey = "company_id"
	RoleKey          ContextKey = "role"
	TokenKey         ContextKey = "token"
	AgentTenantIDKey ContextKey = "agent_tenant_id"
	AgentTokenKey    ContextKey = "agent_token"
)

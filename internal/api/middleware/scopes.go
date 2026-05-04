package middleware

import (
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
)

// List of all active scopes in the system
const (
	ScopeScanRead           = "scan:read"
	ScopeScanWrite          = "scan:write"
	ScopeScanExecute        = "scan:execute"
	ScopeVulnerabilityRead  = "vulnerability:read"
	ScopeVulnerabilityWrite = "vulnerability:write"
	ScopeReportRead         = "report:read"
	ScopeAuthRead           = "auth:read"
	ScopeUserRead           = "user:read"
	ScopeUserWrite          = "user:write"
	ScopeCompanyRead        = "company:read"
	ScopeCompanyWrite       = "company:write"
	ScopeBillingRead        = "billing:read"
	ScopeBillingWrite       = "billing:write"
	ScopeAdminWrite         = "admin:write"
	ScopeAdminRead          = "admin:read"
	ScopeAll                = "*"
)

// roleScopes maps each standard user role to a set of granular permissions (scopes).
var roleScopes = map[types.UserRole][]string{
	types.RoleViewer: {
		ScopeScanRead,
		ScopeVulnerabilityRead,
		ScopeReportRead,
		ScopeAuthRead,
		ScopeCompanyRead,
	},
	types.RoleOperateur: {
		ScopeScanRead,
		ScopeVulnerabilityRead,
		ScopeReportRead,
		ScopeAuthRead,
		ScopeScanWrite,
		ScopeScanExecute,
		ScopeVulnerabilityWrite,
		ScopeCompanyRead,
	},
	types.RoleOwner: {
		ScopeScanRead,
		ScopeVulnerabilityRead,
		ScopeReportRead,
		ScopeAuthRead,
		ScopeScanWrite,
		ScopeScanExecute,
		ScopeVulnerabilityWrite,
		ScopeUserRead,
		ScopeUserWrite,
		ScopeCompanyRead,
		ScopeCompanyWrite,
		ScopeBillingRead,
	},
	types.RoleAdmin: {
		ScopeAll,
	},
	types.RoleSuperAdmin: {
		ScopeAll,
	},
	types.RoleBillingAegis: {
		ScopeAuthRead,
		ScopeCompanyRead,
		ScopeAdminRead,
		ScopeBillingRead,
		ScopeBillingWrite,
	},
	types.RoleTechnicien: {
		ScopeScanRead,
		ScopeVulnerabilityRead,
		ScopeReportRead,
		ScopeAuthRead,
		ScopeCompanyRead,
	},
	types.RoleSupport: {
		ScopeScanRead,
		ScopeAuthRead,
		ScopeCompanyRead,
	},
	types.RoleBillingClient: {
		ScopeAuthRead,
		ScopeCompanyRead,
		ScopeBillingRead,
	},
}

// HasScope checks if a given role has the required permission.
func HasScope(role types.UserRole, requiredScope string) bool {
	scopes, ok := roleScopes[role]
	if !ok {
		return false
	}

	for _, s := range scopes {
		if s == ScopeAll || s == requiredScope {
			return true
		}
	}

	return false
}

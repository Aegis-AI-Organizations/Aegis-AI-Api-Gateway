package middleware

import (
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
)

// List of all active scopes in the system
const (
	ScopeScanRead          = "scan:read"
	ScopeScanWrite         = "scan:write"
	ScopeScanExecute       = "scan:execute"
	ScopeVulnerabilityRead  = "vulnerability:read"
	ScopeVulnerabilityWrite = "vulnerability:write"
	ScopeReportRead        = "report:read"
	ScopeAuthRead          = "auth:read"
	ScopeUserRead          = "user:read"
	ScopeUserWrite         = "user:write"
	ScopeCompanyRead       = "company:read"
	ScopeCompanyWrite      = "company:write"
	ScopeAll               = "*"
)

// RoleScopes maps each standard user role to a set of granular permissions (scopes).
var RoleScopes = map[types.UserRole][]string{
	types.RoleViewer: {
		ScopeScanRead,
		ScopeVulnerabilityRead,
		ScopeReportRead,
		ScopeAuthRead,
	},
	types.RoleOperator: {
		ScopeScanRead,
		ScopeVulnerabilityRead,
		ScopeReportRead,
		ScopeAuthRead,
		ScopeScanWrite,
		ScopeScanExecute,
		ScopeVulnerabilityWrite,
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
	},
	types.RoleSuperAdmin: {
		ScopeAll,
	},
}

// HasScope checks if a given role has the required permission.
func HasScope(role types.UserRole, requiredScope string) bool {
	scopes, ok := RoleScopes[role]
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

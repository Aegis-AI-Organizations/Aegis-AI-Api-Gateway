package models

import (
	"encoding/json"
	"time"
)

type CreateScanRequest struct {
	TargetImage string `json:"target_image"`
}

type CreateScanResponse struct {
	ScanID string `json:"scan_id"`
	Status string `json:"status"`
}

type Scan struct {
	ID                 string  `json:"id"`
	TemporalWorkflowID string  `json:"temporal_workflow_id"`
	TargetImage        string  `json:"target_image"`
	Status             string  `json:"status"`
	StartedAt          string  `json:"started_at"`
	CompletedAt        *string `json:"completed_at,omitempty"`
}

type Vulnerability struct {
	ID             string    `json:"id"`
	VulnType       string    `json:"vuln_type"`
	Severity       string    `json:"severity"`
	TargetEndpoint string    `json:"target_endpoint"`
	Description    string     `json:"description"`
	DiscoveredAt   *time.Time `json:"discovered_at,omitempty"`
}

type Evidence struct {
	ID              string          `json:"id"`
	VulnerabilityID string          `json:"vulnerability_id"`
	PayloadUsed     string          `json:"payload_used"`
	LootData        json.RawMessage `json:"loot_data"`
	CapturedAt      *time.Time      `json:"captured_at,omitempty"`
}

type LicenseStatus string

const (
	LicenseStatusActive  LicenseStatus = "active"
	LicenseStatusExpired LicenseStatus = "expired"
)

type License struct {
	ID            string        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()" db:"id"`
	Name          string        `json:"name" gorm:"type:varchar(255);unique;not null" db:"name"`
	LicenseStatus LicenseStatus `json:"license_status" gorm:"type:varchar(50);not null;default:'active'" db:"license_status"`
	CreatedAt     time.Time     `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
}

type UserRole string

const (
	RoleSuperAdmin UserRole = "superadmin"
	RoleOwner      UserRole = "owner"
	RoleOperator   UserRole = "operator"
	RoleViewer     UserRole = "viewer"
)

type Company struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()" db:"id"`
	Name      string    `json:"name" gorm:"type:varchar(255);unique;not null" db:"name"`
	LogoURL   string    `json:"logo_url,omitempty" gorm:"type:varchar(255)" db:"logo_url"`
	OwnerID   string    `json:"owner_id,omitempty" gorm:"type:uuid" db:"owner_id"`
	IsActive  bool      `json:"is_active" gorm:"default:true" db:"is_active"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
}

type User struct {
	ID           string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()" db:"id"`
	CompanyID    *string   `json:"company_id,omitempty" gorm:"type:uuid" db:"company_id"`
	Email        string    `json:"email" gorm:"type:varchar(255);unique;not null" db:"email"`
	PasswordHash string    `json:"-" gorm:"type:varchar(255);not null" db:"password_hash"`
	Role         UserRole  `json:"role" gorm:"type:user_role;default:'viewer';not null" db:"role"`
	IsActive     bool      `json:"is_active" gorm:"default:true" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
}

type RefreshToken struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()" db:"id"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null" db:"user_id"`
	TokenHash string    `json:"token_hash" gorm:"type:varchar(255);unique;not null" db:"token_hash"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null" db:"expires_at"`
	Revoked   bool      `json:"revoked" gorm:"default:false" db:"revoked"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
}

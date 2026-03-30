package models

import (
	"encoding/json"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
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

type License struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	LicenseStatus types.LicenseStatus `json:"license_status"`
	CreatedAt     time.Time           `json:"created_at"`
}

type Company struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	LogoURL   string    `json:"logo_url,omitempty"`
	OwnerID   string    `json:"owner_id,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID        string         `json:"id"`
	CompanyID *string        `json:"company_id,omitempty"`
	Email     string         `json:"email"`
	Role      types.UserRole `json:"role"`
	IsActive  bool           `json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

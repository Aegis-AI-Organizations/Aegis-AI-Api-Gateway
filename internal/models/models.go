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

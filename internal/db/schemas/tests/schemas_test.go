package schemas_test

import (
	"testing"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/db/schemas"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestLicense_ToDTO(t *testing.T) {
	now := time.Now()
	s := &schemas.License{
		ID:            "test-id",
		Name:          "test-name",
		LicenseStatus: types.LicenseStatusActive,
		CreatedAt:     now,
	}

	dto, err := s.ToDTO()
	assert.NoError(t, err)
	assert.Equal(t, s.ID, dto.ID)
	assert.Equal(t, s.Name, dto.Name)
	assert.Equal(t, string(s.LicenseStatus), string(dto.LicenseStatus))
	assert.True(t, s.CreatedAt.Equal(dto.CreatedAt))
}

func TestCompany_ToDTO(t *testing.T) {
	now := time.Now()
	s := &schemas.Company{
		ID:        "test-id",
		Name:      "test-name",
		LogoURL:   "test-url",
		OwnerID:   "test-owner",
		IsActive:  true,
		CreatedAt: now,
	}

	dto, err := s.ToDTO()
	assert.NoError(t, err)
	assert.Equal(t, s.ID, dto.ID)
	assert.Equal(t, s.Name, dto.Name)
	assert.Equal(t, s.LogoURL, dto.LogoURL)
	assert.Equal(t, s.OwnerID, dto.OwnerID)
	assert.Equal(t, s.IsActive, dto.IsActive)
	assert.True(t, s.CreatedAt.Equal(dto.CreatedAt))
}

func TestUser_ToDTO(t *testing.T) {
	now := time.Now()
	companyID := "test-company"
	s := &schemas.User{
		ID:           "test-id",
		CompanyID:    &companyID,
		Email:        "test@example.com",
		PasswordHash: "hashed",
		Role:         types.RoleViewer,
		IsActive:     true,
		CreatedAt:    now,
	}

	dto, err := s.ToDTO()
	assert.NoError(t, err)
	assert.Equal(t, s.ID, dto.ID)
	assert.Equal(t, s.CompanyID, dto.CompanyID)
	assert.Equal(t, s.Email, dto.Email)
	assert.Equal(t, string(s.Role), string(dto.Role))
	assert.Equal(t, s.IsActive, dto.IsActive)
	assert.True(t, s.CreatedAt.Equal(dto.CreatedAt))
}

func TestRefreshToken_ToDTO(t *testing.T) {
	now := time.Now()
	expires := now.Add(24 * time.Hour)
	s := &schemas.RefreshToken{
		ID:        "test-id",
		UserID:    "test-user",
		TokenHash: "test-hash",
		ExpiresAt: expires,
		Revoked:   false,
		CreatedAt: now,
	}

	dto, err := s.ToDTO()
	assert.NoError(t, err)
	assert.Equal(t, s.ID, dto.ID)
	assert.Equal(t, s.UserID, dto.UserID)
	assert.Equal(t, s.TokenHash, dto.TokenHash)
	assert.True(t, s.ExpiresAt.Equal(dto.ExpiresAt))
	assert.Equal(t, s.Revoked, dto.Revoked)
	assert.True(t, s.CreatedAt.Equal(dto.CreatedAt))
}

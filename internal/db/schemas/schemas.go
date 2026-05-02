package schemas

import (
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/models"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
	"github.com/jinzhu/copier"
)

type License struct {
	ID            string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" db:"id"`
	Name          string              `gorm:"type:varchar(255);unique;not null" db:"name"`
	LicenseStatus types.LicenseStatus `gorm:"type:varchar(50);not null;default:'active'" db:"license_status"`
	CreatedAt     time.Time           `gorm:"autoCreateTime" db:"created_at"`
}

func (s *License) ToDTO() (*models.License, error) {
	dto := &models.License{}
	if err := copier.Copy(dto, s); err != nil {
		return nil, err
	}
	return dto, nil
}

type Company struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" db:"id"`
	Name      string    `gorm:"type:varchar(255);unique;not null" db:"name"`
	LogoURL   string    `gorm:"type:varchar(255)" db:"logo_url"`
	OwnerID   string    `gorm:"type:uuid" db:"owner_id"`
	IsActive  bool      `gorm:"default:true" db:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" db:"created_at"`
}

func (s *Company) ToDTO() (*models.Company, error) {
	dto := &models.Company{}
	if err := copier.Copy(dto, s); err != nil {
		return nil, err
	}
	return dto, nil
}

type User struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" db:"id"`
	CompanyID    *string        `gorm:"type:uuid" db:"company_id"`
	Email        string         `gorm:"type:varchar(255);unique;not null" db:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" db:"password_hash"`
	Role         types.UserRole `gorm:"type:user_role;default:'viewer';not null" db:"role"`
	IsActive     bool           `gorm:"default:true" db:"is_active"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" db:"created_at"`
}

func (s *User) ToDTO() (*models.User, error) {
	dto := &models.User{}
	if err := copier.Copy(dto, s); err != nil {
		return nil, err
	}
	return dto, nil
}

type RefreshToken struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" db:"id"`
	UserID    string    `gorm:"type:uuid;not null" db:"user_id"`
	TokenHash string    `gorm:"type:varchar(255);unique;not null" db:"token_hash"`
	ExpiresAt time.Time `gorm:"not null" db:"expires_at"`
	Revoked   bool      `gorm:"default:false" db:"revoked"`
	CreatedAt time.Time `gorm:"autoCreateTime" db:"created_at"`
}

func (s *RefreshToken) ToDTO() (*models.RefreshToken, error) {
	dto := &models.RefreshToken{}
	if err := copier.Copy(dto, s); err != nil {
		return nil, err
	}
	return dto, nil
}

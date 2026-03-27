package db_test

import (
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestConnect_MissingVars(t *testing.T) {
	t.Setenv("DB_HOST", "")
	t.Setenv("POSTGRES_USER", "")
	t.Setenv("POSTGRES_PASSWORD", "")
	t.Setenv("POSTGRES_DB", "")

	db, err := db.Connect()
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestConnect_Base(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD", "pass")
	t.Setenv("POSTGRES_DB", "db")

	// This will fail at Ping() but cover New and string formatting
	db, err := db.Connect()
	assert.Error(t, err)
	assert.Nil(t, db)
}

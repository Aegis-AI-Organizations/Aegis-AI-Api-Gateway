package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect_MissingVars(t *testing.T) {
	_ = os.Unsetenv("DB_HOST")
	_ = os.Unsetenv("POSTGRES_USER")
	_ = os.Unsetenv("POSTGRES_PASSWORD")
	_ = os.Unsetenv("POSTGRES_DB")

	db, err := Connect()
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "missing one or more required database environment variables")
}

func TestConnect_Base(t *testing.T) {
	_ = os.Setenv("DB_HOST", "localhost")
	_ = os.Setenv("POSTGRES_USER", "user")
	_ = os.Setenv("POSTGRES_PASSWORD", "pass")
	_ = os.Setenv("POSTGRES_DB", "db")

	// This will fail at Ping() but cover New and string formatting
	db, err := Connect()
	assert.Error(t, err)
	assert.Nil(t, db)
}

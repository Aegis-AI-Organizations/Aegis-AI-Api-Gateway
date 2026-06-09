package handlers_test

import (
	"os"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api"
	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestInitDB_MissingVars(t *testing.T) {
	// Clear env vars
	os.Unsetenv("DB_HOST")
	os.Unsetenv("POSTGRES_USER")
	os.Unsetenv("POSTGRES_PASSWORD")
	os.Unsetenv("POSTGRES_DB")

	err := api.InitDB()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing database environment variables")
}

func TestNewMinioClient_MissingVars(t *testing.T) {
	os.Unsetenv("MINIO_ENDPOINT")
	os.Unsetenv("MINIO_ACCESS_KEY")
	os.Unsetenv("MINIO_SECRET_KEY")
	os.Unsetenv("MINIO_BUCKET")

	client, err := db.NewMinioClient()
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "missing MinIO configuration")
}

func TestNewRedisClient_MissingVars(t *testing.T) {
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")

	client, err := db.NewRedisClient()
	assert.Error(t, err)
	assert.Nil(t, client)
}

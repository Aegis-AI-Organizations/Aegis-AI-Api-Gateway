package db_test

import (
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestRedisClient_Struct(t *testing.T) {
	client := &db.RedisClient{Client: nil}
	assert.Nil(t, client.GetClient())
}

func TestMinioClient_Struct(t *testing.T) {
	client := &db.MinioClient{Bucket: "test"}
	assert.Equal(t, "test", client.Bucket)
}

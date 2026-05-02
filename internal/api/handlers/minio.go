package handlers

import (
	"net/http"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MinioHandler struct {
	MinioClient *db.MinioClient
}

// GetUploadURLHandler generates a presigned PUT URL for MinIO file uploads.
func (h *MinioHandler) GetUploadURLHandler(c *gin.Context) {
	// Identify by Agent token or User ID to avoid anonymous abuse
	// (Already handled by middlewares)

	// Generate a unique object name
	objectName := uuid.New().String()

	// Optional: allow client to suggest a file extension or prefix
	if prefix := c.Query("prefix"); prefix != "" {
		objectName = prefix + "/" + objectName
	}

	url, err := h.MinioClient.GeneratePresignedPutURL(c.Request.Context(), objectName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate presigned upload URL"})
		return
	}

	// Acceptance Criteria: L'URL expire strictement après 15 minutes.
	c.JSON(http.StatusOK, gin.H{
		"url":         url,
		"object_name": objectName,
		"expires_in":  "15m",
	})
}

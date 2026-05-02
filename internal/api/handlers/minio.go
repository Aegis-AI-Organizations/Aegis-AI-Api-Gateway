package handlers

import (
	"net/http"
	"path"
	"strings"

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
	if h.MinioClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "File storage is not enabled"})
		return
	}

	// Generate a unique object name
	objectName := uuid.New().String()

	// Optional: allow client to suggest a file extension or prefix
	// SECURITY: We MUST validate the prefix to avoid cross-tenant access or path traversal.
	prefix := c.Query("prefix")
	if prefix != "" {
		// Sanitize prefix to prevent directory traversal attacks (e.g. ../other_company/)
		prefix = path.Clean("/" + prefix)
		prefix = strings.TrimPrefix(prefix, "/")
	}

	companyID, _ := c.Get("company_id")

	if prefix != "" && prefix != "." {
		// Scoping to company_id if available for tenant isolation
		if companyID != nil {
			objectName = companyID.(string) + "/" + prefix + "/" + objectName
		} else {
			objectName = prefix + "/" + objectName
		}
	} else {
		// Default scoping if no prefix provided or if prefix was entirely malicious (e.g., "../")
		if companyID != nil {
			objectName = companyID.(string) + "/uploads/" + objectName
		}
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

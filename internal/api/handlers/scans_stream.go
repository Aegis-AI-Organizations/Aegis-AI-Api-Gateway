package handlers

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ScanStreamHandler handles GET /scans/stream or /scans/:id/stream
func (a *API) ScanStreamHandler(c *gin.Context) {
	scanID := c.Param("id")
	if scanID == "" {
		log.Printf("📡 Starting global SSE stream")
	} else {
		log.Printf("📡 Starting SSE stream for scanID: %s", scanID)
	}

	stream, err := a.GRPCClient.WatchScanStatus(c.Request.Context(), scanID)
	if err != nil || stream == nil {
		log.Printf("Failed to open gRPC stream: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize stream"})
		return
	}

	c.Stream(func(w io.Writer) bool {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Printf("gRPC stream finished: %v", err)
			} else {
				log.Printf("gRPC stream error: %v", err)
			}
			return false
		}

		c.SSEvent("message", gin.H{
			"scan_id": resp.ScanId,
			"status":  resp.Status,
		})
		return true
	})
}

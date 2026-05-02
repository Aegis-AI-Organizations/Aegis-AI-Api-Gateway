package handlers

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TeamStreamHandler handles GET /admin/teams/stream
func (a *API) TeamStreamHandler(c *gin.Context) {
	log.Printf("📡 Starting teams SSE stream")

	stream, err := a.GRPCClient.WatchCompanyUpdates(c.Request.Context())
	if err != nil || stream == nil {
		log.Printf("Failed to open gRPC stream for teams: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize stream"})
		return
	}

	// Set standard SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Accel-Buffering", "no")

	c.Stream(func(w io.Writer) bool {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Printf("Teams gRPC stream finished")
			} else {
				log.Printf("Teams gRPC stream error: %v", err)
			}
			return false
		}

		c.SSEvent("message", gin.H{
			"event_type":  resp.EventType,
			"entity_id":   resp.EntityId,
			"entity_name": resp.EntityName,
		})
		return true
	})
}

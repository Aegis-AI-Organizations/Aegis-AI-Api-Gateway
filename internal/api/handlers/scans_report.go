package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) GetScanReportHandler(c *gin.Context) {
	scanID := c.Param("id")
	log.Printf("📥 Received report request for scan ID: %s", scanID)
	if scanID == "" {
		log.Printf("⚠️  Scan ID missing in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "scan id parameter is required"})
		return
	}

	pdfBytes, err := a.GRPCClient.GetScanReport(c.Request.Context(), scanID)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			log.Printf("GRPC GetScanReport not found: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Scan or report not found"})
		} else {
			log.Printf("GRPC GetScanReport error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		}
		return
	}

	if len(pdfBytes) == 0 {
		log.Printf("⚠️  Report PDF is empty for scan ID: %s", scanID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Report PDF not yet generated or is empty"})
		return
	}

	log.Printf("📤 Sending report PDF (%d bytes) for scan ID: %s", len(pdfBytes), scanID)

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"scan_report_%s.pdf\"", scanID))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)

	log.Printf("✅ Drafted response for scan ID: %s", scanID)
}

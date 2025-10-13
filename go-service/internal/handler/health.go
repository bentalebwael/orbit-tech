package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wbentaleb/student-report-service/internal/client"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	backendClient *client.BackendClient
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(backendClient *client.BackendClient) *HealthHandler {
	return &HealthHandler{
		backendClient: backendClient,
	}
}

// Handle processes health check requests
func (h *HealthHandler) Handle(c *gin.Context) {
	backendHealthy := h.backendClient.CheckHealth()

	response := gin.H{
		"status":  "healthy",
		"service": "go-report-service",
		"backend": map[string]interface{}{
			"reachable": backendHealthy,
		},
	}

	if !backendHealthy {
		response["status"] = "degraded"
		response["message"] = "Backend service is not reachable"
	}

	c.JSON(http.StatusOK, response)
}

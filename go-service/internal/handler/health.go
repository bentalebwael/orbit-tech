package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wbentaleb/student-report-service/internal/dto"
	"github.com/wbentaleb/student-report-service/internal/external"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	backendClient external.BackendService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(backendClient external.BackendService) *HealthHandler {
	return &HealthHandler{
		backendClient: backendClient,
	}
}

// Handle processes health check requests
func (h *HealthHandler) Handle(c *gin.Context) {
	backendHealthy := h.backendClient.CheckHealth(c.Request.Context())

	status := "healthy"
	if !backendHealthy {
		status = "degraded"
	}

	response := dto.HealthResponse{
		Status:  status,
		Service: "go-report-service",
		Backend: map[string]bool{
			"reachable": backendHealthy,
		},
	}

	c.JSON(http.StatusOK, response)
}

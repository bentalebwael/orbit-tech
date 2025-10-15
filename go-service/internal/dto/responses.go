package dto

type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

type HealthResponse struct {
	Status  string          `json:"status"`
	Service string          `json:"service"`
	Backend map[string]bool `json:"backend"`
}

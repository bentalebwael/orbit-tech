package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wbentaleb/student-report-service/internal/dto"
	"github.com/wbentaleb/student-report-service/internal/errors"
	"go.uber.org/zap"
)

// BackendClient handles communication with the Node.js backend
type BackendClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewBackendClient creates a new backend client instance
func NewBackendClient(baseURL, apiKey string, logger *zap.Logger) *BackendClient {
	return &BackendClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetStudent fetches a student by ID from the backend with retry logic
func (c *BackendClient) GetStudent(ctx context.Context, id string) (*dto.Student, error) {
	var lastErr error
	delays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for attempt := 0; attempt < 3; attempt++ {
		student, err := c.getStudent(ctx, id)
		if err == nil {
			c.logger.Info("Successfully fetched student",
				zap.String("student_id", id),
				zap.Int("attempt", attempt+1))
			return student, nil
		}

		lastErr = err

		// Don't retry on non-retryable errors (404, etc.)
		if errors.IsNotFound(err) {
			c.logger.Warn("Student not found",
				zap.String("student_id", id))
			return nil, err
		}

		// Retry on transient errors
		if attempt < 2 {
			c.logger.Warn("Request failed, retrying",
				zap.String("student_id", id),
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delays[attempt]),
				zap.Error(err))
			time.Sleep(delays[attempt])
		}
	}

	c.logger.Error("Failed to fetch student after all retries",
		zap.String("student_id", id),
		zap.Error(lastErr))
	return nil, fmt.Errorf("failed after 3 attempts: %w", lastErr)
}

// getStudent performs a single request to fetch student data
func (c *BackendClient) getStudent(ctx context.Context, id string) (*dto.Student, error) {
	url := fmt.Sprintf("%s/api/v1/students/%s", c.baseURL, id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key for authentication
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, &errors.NotFoundError{Resource: "Student"}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &errors.ServiceError{
			Service: "backend",
			Err:     fmt.Errorf("returned status %d: %s", resp.StatusCode, string(body)),
		}
	}

	var student dto.Student
	if err := json.Unmarshal(body, &student); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &student, nil
}

// CheckHealth verifies if the backend is reachable
func (c *BackendClient) CheckHealth(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/health", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

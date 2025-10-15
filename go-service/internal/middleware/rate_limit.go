package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wbentaleb/student-report-service/internal/dto"
)

func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("RequestID"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	rate     int           // Maximum requests per window
	window   time.Duration // Time window for rate limiting
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		rate:     requestsPerMinute,
		window:   time.Minute,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rl.mu.Lock()
		defer rl.mu.Unlock()

		ip := c.ClientIP()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		// Clean old requests
		var validRequests []time.Time
		for _, reqTime := range rl.requests[ip] {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) >= rl.rate {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error:     "Rate limit exceeded. Maximum " + fmt.Sprintf("%d", rl.rate) + " requests per minute allowed.",
				RequestID: getRequestID(c),
			})
			return
		}

		rl.requests[ip] = append(validRequests, now)
		c.Next()
	}
}

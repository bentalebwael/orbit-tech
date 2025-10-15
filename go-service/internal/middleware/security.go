package middleware

import "github.com/gin-gonic/gin"

func BasicSecurity() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("X-Content-Type-Options", "nosniff") // Prevent MIME type sniffing
		c.Header("X-Frame-Options", "DENY")           // Prevent clickjacking

		// Don't cache sensitive data
		if c.Request.URL.Path != "/health" {
			c.Header("Cache-Control", "private, no-store")
		}

		c.Next()
	}
}

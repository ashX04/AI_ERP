package middleware

import (
	"fmt"
	"time"

	"github.com/ashX04/new_website/internal/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Get user information from session
		session := sessions.Default(c)
		userID := session.Get("userID")

		// Get request details
		path := c.Request.URL.Path
		method := c.Request.Method
		ip := c.ClientIP()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get response status
		status := c.Writer.Status()

		// Create user identifier
		userIdentifier := "anonymous"
		if userID != nil {
			userIdentifier = fmt.Sprintf("user_%v", userID)
		}

		// Log the request
		logMessage := fmt.Sprintf(
			"[%s] %s %s %s | Status: %d | Latency: %v | IP: %s",
			userIdentifier,
			method,
			path,
			c.Request.URL.RawQuery,
			status,
			latency,
			ip,
		)

		utils.Logger.Println(logMessage)

		// If there was an error, log it separately
		if len(c.Errors) > 0 {
			utils.Logger.Printf("[ERROR] %s | Errors: %v", userIdentifier, c.Errors)
		}
	}
}

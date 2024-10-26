package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors
		if len(c.Errors) > 0 {
			// Log error
			log.Printf("Error: %v", c.Errors.Last().Err)

			// Don't expose internal errors to users
			c.JSON(500, gin.H{
				"error": "An internal error occurred",
			})
			return
		}
	}
}

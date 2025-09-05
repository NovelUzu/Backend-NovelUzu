package utils

import (
	// "NovelUzu/models/postgres"
	// "fmt"
	"net/http"
	"time"

	// "gorm.io/gorm"

	"github.com/gin-gonic/gin"
	// "github.com/zishang520/socket.io/v2/socket"
)

// LoggerMiddleware logs information about each request
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start time
		startTime := time.Now()

		// Process request
		c.Next()

		// End time
		endTime := time.Now()

		// Calculate latency
		latency := endTime.Sub(startTime)

		// Get request details
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()

		// Register information
		c.JSON(http.StatusOK, gin.H{
			"status":  statusCode,
			"latency": latency,
			"method":  method,
			"path":    path,
		})
	}
}

// ErrorHandler handles global errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}




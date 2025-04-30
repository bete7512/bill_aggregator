// internal/adapter/inbound/rest/middleware/logger.go
package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/bete7512/bill-aggregator/pkg/util/logger"
	"github.com/gin-gonic/gin"
)

// LoggerMiddleware is a middleware for logging HTTP requests
type LoggerMiddleware struct {
	logger *logger.Logger
}

// NewLoggerMiddleware creates a new logger middleware
func NewLoggerMiddleware(logger *logger.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		logger: logger,
	}
}

// Logger returns a gin middleware for logging HTTP requests
func (m *LoggerMiddleware) Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Create a copy of the request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Restore the request body for subsequent handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create a response writer that captures the response
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Prepare log fields
		fields := map[string]interface{}{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"ip":         c.ClientIP(),
			"user-agent": c.Request.UserAgent(),
			"latency":    latency.String(),
		}

		// Add user ID if available
		if userID, exists := c.Get("userID"); exists {
			fields["user_id"] = userID
		}

		// Add request body for non-GET requests (limited to prevent large logs)
		if c.Request.Method != "GET" && len(requestBody) > 0 {
			// Limit body size to 1KB
			if len(requestBody) > 1024 {
				fields["request_body"] = string(requestBody[:1024]) + "..."
			} else {
				fields["request_body"] = string(requestBody)
			}
		}

		// Add response body (limited to prevent large logs)
		responseBody := responseWriter.body.String()
		if len(responseBody) > 0 {
			// Limit body size to 1KB
			if len(responseBody) > 1024 {
				fields["response_body"] = responseBody[:1024] + "..."
			} else {
				fields["response_body"] = responseBody
			}
		}

		// Log based on status code
		logMsg := fmt.Sprintf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)

		if c.Writer.Status() >= 500 {
			m.logger.Error(logMsg, fields)
		} else if c.Writer.Status() >= 400 {
			m.logger.Warn(logMsg, fields)
		} else {
			m.logger.Info(logMsg, fields)
		}
	}
}

// responseBodyWriter captures the response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the written response
func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString captures the written response string
func (w *responseBodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

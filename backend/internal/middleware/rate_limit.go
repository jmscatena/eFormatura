package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
)

// RateLimit middleware
func RateLimit(limit float64, duration time.Duration) gin.HandlerFunc {
	bucket := ratelimit.NewBucket(duration, int64(limit))

	return func(c *gin.Context) {
		// Verificar limite
		if bucket.TakeAvailable(1) == 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

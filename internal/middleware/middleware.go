package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware sets CORS headers for the given frontend URL.
func CORSMiddleware(frontendURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", frontendURL)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-API-Key")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// tokenBucket tracks request tokens for a single IP.
type tokenBucket struct {
	tokens    float64
	lastCheck time.Time
	mu        sync.Mutex
}

// RateLimiter enforces an IP-based rate limit using a token bucket algorithm.
func RateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	var buckets sync.Map
	rate := float64(limit) / window.Seconds()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		val, _ := buckets.LoadOrStore(ip, &tokenBucket{
			tokens:    float64(limit),
			lastCheck: time.Now(),
		})
		bucket := val.(*tokenBucket)

		bucket.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(bucket.lastCheck).Seconds()
		bucket.tokens += elapsed * rate
		if bucket.tokens > float64(limit) {
			bucket.tokens = float64(limit)
		}
		bucket.lastCheck = now

		if bucket.tokens < 1 {
			bucket.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		bucket.tokens--
		bucket.mu.Unlock()

		c.Next()
	}
}

// RequestLogger logs each request with method, path, status, and latency.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
			"ip", c.ClientIP(),
		)
	}
}

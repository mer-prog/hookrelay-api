package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		requests   int
		wantOK     int
		wantLimited int
	}{
		{
			name:        "all requests within limit",
			limit:       5,
			requests:    5,
			wantOK:      5,
			wantLimited: 0,
		},
		{
			name:        "exceeds limit",
			limit:       3,
			requests:    6,
			wantOK:      3,
			wantLimited: 3,
		},
		{
			name:        "single request allowed",
			limit:       1,
			requests:    3,
			wantOK:      1,
			wantLimited: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(RateLimiter(tt.limit, time.Minute))
			r.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			okCount := 0
			limitedCount := 0

			for i := 0; i < tt.requests; i++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.RemoteAddr = "192.168.1.1:12345"
				r.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					okCount++
				} else if w.Code == http.StatusTooManyRequests {
					limitedCount++
				}
			}

			assert.Equal(t, tt.wantOK, okCount)
			assert.Equal(t, tt.wantLimited, limitedCount)
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(CORSMiddleware("http://localhost:3000"))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("preflight OPTIONS", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})

	t.Run("normal GET", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

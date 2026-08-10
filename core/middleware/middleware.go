// Package middleware provides HTTP middleware.
// Functions are defined here for simplicity; individual .go files reference this canonical source.

// Package middleware implements middleware logic.

package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"kingfisher/core/cache"
	"kingfisher/core/errcode"
	"kingfisher/core/response"
	"kingfisher/core/telemetry"
)

// InitValidator registers custom password validator
func InitValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("password", validatePassword)
	}
}

func validatePassword(fl validator.FieldLevel) bool {
	pwd := fl.Field().String()
	if len(pwd) < 8 || len(pwd) > 64 {
		return false
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(pwd)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(pwd)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(pwd)
	return hasUpper && hasLower && hasDigit
}

// RequestID generates or passes through X-Request-ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Recovery catches panics and returns 500 (max request body 10MB)
func Recovery() gin.HandlerFunc {
	return RecoveryWithLimit(10 << 20)
}

// RecoveryWithLimit catches panics with configurable max body size
func RecoveryWithLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger := getLogger(c)
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("request_id", c.GetString("request_id")),
					zap.String("stack", string(debug.Stack())),
				)
				response.AbortJSON(c, response.Error(errcode.ErrInternal))
			}
		}()
		c.Next()
	}
}

// Trace creates a request trace span using telemetry and sets trace_id header.
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Set("trace_id", traceID)
		ctx, span := telemetry.StartSpan(c.Request.Context(), "HTTP "+c.Request.Method+" "+c.FullPath())
		c.Request = c.Request.WithContext(ctx)
		defer span.End()
		c.Next()
	}
}

// Logger logs method, path, status, latency
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("request_id", c.GetString("request_id")),

			zap.String("trace_id", c.GetString("trace_id")),
			zap.Int("body_size", c.Writer.Size()),
		)
	}
}

// CORS handles cross-origin requests
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-ID,X-Trace-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// SecurityHeaders adds security-related response headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Next()
	}
}

// RateLimit implements sliding window rate limiting via Redis
func RateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cache == nil {
			c.Next() // No Redis, no rate limit
			return
		}
		key := "ratelimit:" + c.ClientIP() + ":" + c.FullPath()
		count, err := cache.Incr(c.Request.Context(), key)
		if err != nil {
			c.Next() // Redis error, allow through
			return
		}
		if count == 1 {
			_ = cache.Expire(c.Request.Context(), key, window)
		}
		if count > int64(limit) {
			c.Header("Retry-After", window.String())
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			response.AbortJSON(c, response.Error(errcode.ErrTooManyRequest))
			return
		}
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-int(count)))
		c.Next()
	}
}

func getLogger(c *gin.Context) *zap.Logger {
	if l, ok := c.Get("logger"); ok {
		return l.(*zap.Logger)
	}
	return zap.NewNop()
}

var _ = http.StatusOK // use net/http

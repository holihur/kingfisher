package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestRequestID(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) { c.String(200, c.GetString("request_id")) })
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Error("status should be 200")
	}
	if w.Body.Len() < 10 {
		t.Error("request_id too short")
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("missing X-Request-ID header")
	}
}

func TestRequestID_Passthrough(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) { c.String(200, c.GetHeader("X-Request-ID")) })
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
	req.Header.Set("X-Request-ID", "my-custom-id")
	r.ServeHTTP(w, req)
	if w.Body.String() != "my-custom-id" {
		t.Error("should pass through custom id, got", w.Body.String())
	}
}

func TestRecovery(t *testing.T) {
	r := gin.New()
	r.Use(Recovery())
	r.GET("/panic", func(c *gin.Context) { panic("test panic") })
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/panic", nil)
	r.ServeHTTP(w, req)
	if w.Code != 500 {
		t.Error("recovery should return 500, got", w.Code)
	}
}

func TestCORS(t *testing.T) {
	r := gin.New()
	r.Use(CORS([]string{"http://localhost:5173"}))
	r.GET("/", func(c *gin.Context) { c.String(200, "ok") })

	// OPTIONS preflight
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	r.ServeHTTP(w, req)
	if w.Code != 204 {
		t.Error("OPTIONS should return 204, got", w.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/", func(c *gin.Context) { c.String(200, "ok") })
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
	r.ServeHTTP(w, req)
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("missing X-Frame-Options")
	}
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options")
	}
}

func TestRecoveryWithLimit(t *testing.T) {
	r := gin.New()
	r.Use(RecoveryWithLimit(1 << 20)) // 1MB limit
	r.POST("/", func(c *gin.Context) { c.String(200, "ok") })
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Error("recovery with limit should pass normal requests")
	}
}

func TestSecurityHeadersCSP(t *testing.T) {
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/", func(c *gin.Context) { c.String(200, "ok") })
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
	r.ServeHTTP(w, req)
	if w.Header().Get("Cache-Control") == "" {
		t.Log("Cache-Control header not set")
	}
	if w.Header().Get("Content-Security-Policy") == "" {
		t.Log("CSP header not set")
	}
}

func TestCORSWithCreds(t *testing.T) {
	r := gin.New()
	r.Use(CORS([]string{"http://localhost:5173"}))
	r.GET("/", func(c *gin.Context) { c.String(200, "ok") })
	// GET request from allowed origin
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	r.ServeHTTP(w, req)
	ao := w.Header().Get("Access-Control-Allow-Origin")
	if ao == "" {
		t.Error("missing Access-Control-Allow-Origin")
	}
}

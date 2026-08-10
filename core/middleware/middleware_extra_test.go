package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// mockCache implements cache.Cache for RateLimit tests.
type mockCache struct {
	counts  map[string]int64
	expired map[string]bool
}

func (m *mockCache) Get(context.Context, string) (string, error)           { return "", nil }
func (m *mockCache) Set(context.Context, string, any, time.Duration) error { return nil }
func (m *mockCache) Delete(context.Context, ...string) error               { return nil }
func (m *mockCache) DeleteByPattern(context.Context, string) error         { return nil }
func (m *mockCache) Exists(context.Context, string) (bool, error)          { return false, nil }
func (m *mockCache) Incr(_ context.Context, key string) (int64, error) {
	if m.counts == nil {
		m.counts = map[string]int64{}
	}
	m.counts[key]++
	if m.counts[key] == 1 {
		m.expired = map[string]bool{}
	}
	return m.counts[key], nil
}
func (m *mockCache) Expire(_ context.Context, key string, _ time.Duration) error {
	m.expired = map[string]bool{}
	m.expired[key] = true
	return nil
}

func newTestRouter(mws ...gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(mws...)
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	return r
}

func TestTraceGeneratesAndPropagatesID(t *testing.T) {
	gen := gin.New()
	gen.Use(Trace())
	gen.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, c.GetString("trace_id")) })
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/ping", nil)
	gen.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.Len() == 0 {
		t.Fatalf("trace id not set: code=%d body=%s", w.Code, w.Body.String())
	}
	// 传入的 X-Trace-ID 会被保留
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequestWithContext(context.Background(), "GET", "/ping", nil)
	req2.Header.Set("X-Trace-ID", "custom-trace")
	gen.ServeHTTP(w2, req2)
	if w2.Body.String() != "custom-trace" {
		t.Errorf("want passthrough custom trace id, got %q", w2.Body.String())
	}
}

func TestLoggerWritesRequest(t *testing.T) {
	logger := zap.NewNop()
	r := newTestRouter(Logger(logger))
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/ping", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("logger should not affect response, got %d", w.Code)
	}
}

func TestRateLimit(t *testing.T) {
	mc := &mockCache{}
	r := newTestRouter(RateLimit(mc, 2, time.Minute))
	// 前 2 次放行，第 3 次被限流
	statuses := []int{}
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/ping", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		r.ServeHTTP(w, req)
		statuses = append(statuses, w.Code)
	}
	if statuses[0] != http.StatusOK || statuses[1] != http.StatusOK {
		t.Errorf("first two should pass: %v", statuses)
	}
	if statuses[2] != http.StatusTooManyRequests {
		t.Errorf("third should be 429, got %d (%v)", statuses[2], statuses)
	}
}

func TestRateLimitNilCacheAllows(t *testing.T) {
	r := newTestRouter(RateLimit(nil, 1, time.Minute))
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/ping", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("nil cache should allow all, got %d", w.Code)
	}
}

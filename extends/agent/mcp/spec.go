package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// swaggerDoc 只需解析到 paths 层（提取 method/path/summary）。
type swaggerDoc struct {
	Paths map[string]map[string]swaggerOperation `json:"paths"`
}

type swaggerOperation struct {
	Summary string   `json:"summary"`
	Tags    []string `json:"tags"`
}

// specTTL 端点清单缓存时长。
const specTTL = 30 * time.Second

// SpecLoader 从系统自身的 OpenAPI 规范（/swagger/doc.json）提取可调用的端点清单，
// 注入 LLM system prompt，供其选择调用哪个接口。
type SpecLoader struct {
	selfBaseURL string
	httpClient  *http.Client

	mu       sync.Mutex
	cached   string
	cachedAt time.Time
}

// NewSpecLoader 创建 SpecLoader。
func NewSpecLoader(selfBaseURL string) *SpecLoader {
	return &SpecLoader{
		selfBaseURL: strings.TrimRight(selfBaseURL, "/"),
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// EndpointList 返回端点清单文本（含短 TTL 内存缓存），格式：
//
//	GET /api/v1/users - 用户列表（分页）
//	POST /api/v1/messages - 发送站内信
func (s *SpecLoader) EndpointList() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cached != "" && time.Since(s.cachedAt) < specTTL {
		return s.cached, nil
	}

	doc, err := s.fetch()
	if err != nil {
		// 拉取失败时若已有缓存，降级返回旧缓存。
		if s.cached != "" {
			return s.cached, nil
		}
		return "", err
	}

	var sb strings.Builder
	var lines []string
	for path, ops := range doc.Paths {
		// 只保留 GET/POST/PUT/PATCH/DELETE 操作。
		for _, m := range []string{"get", "post", "put", "patch", "delete"} {
			op, ok := ops[m]
			if !ok {
				continue
			}
			summary := strings.TrimSpace(op.Summary)
			line := strings.ToUpper(m) + " /api/v1" + path
			if summary != "" {
				line += " - " + summary
			}
			lines = append(lines, line)
		}
	}
	sort.Strings(lines)
	for _, l := range lines {
		sb.WriteString(l)
		sb.WriteString("\n")
	}

	s.cached = sb.String()
	s.cachedAt = time.Now()
	return s.cached, nil
}

// fetch 拉取并解析 swagger doc.json。
func (s *SpecLoader) fetch() (*swaggerDoc, error) {
	url := s.selfBaseURL + "/swagger/doc.json"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("构造 OpenAPI spec 请求失败: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("拉取 OpenAPI spec 失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("拉取 OpenAPI spec 状态异常: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("读取 OpenAPI spec 失败: %w", err)
	}
	var doc swaggerDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("解析 OpenAPI spec 失败: %w", err)
	}
	return &doc, nil
}

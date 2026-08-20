package transport

import (
	"context"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/config"
	adapter "kingfisher/extends/agent/adapter/mysql"
	"kingfisher/extends/agent/app"
	"kingfisher/extends/agent/llm"
	"kingfisher/extends/agent/mcp"
	"kingfisher/extends/agent/port"
	rbacTransport "kingfisher/extends/rbac/transport"
)

// AgentModule Agent 聊天模块，实现 router.Module。
type AgentModule struct {
	handler *AgentHandler
	svc     *app.AgentService
}

// NewAgentModule 组装 Agent 模块。
//
// cfg 提供 base_url/model/api_key 环境名等；getLLMKey 用于运行时读取
// system_configs.llm_api_key（后台可改）；getSystemPrompt 返回可覆盖的
// 系统提示词（后台 agent_system_prompt，为空回退默认）；getAllowedMethods 返回
// call_api 的 HTTP 方法白名单（逗号分隔，如 "GET,POST,PUT"，空用默认）。
// selfBaseURL 为 call_api 内部请求目标。
func NewAgentModule(db *gorm.DB, cfg *config.Config, selfBaseURL string, getLLMKey, getSystemPrompt, getAllowedMethods func(ctx context.Context) (string, error)) *AgentModule {
	repo := adapter.NewAgentRepo(db)
	llmClient := llm.NewClient(cfg.Agent.BaseURL, "", cfg.Agent.Model, cfg.Agent.MaxTokens)
	mcpClient := mcp.NewClient(selfBaseURL)
	// 应用 method 白名单（配置非空时覆盖默认；启动时读一次，静态生效）。
	if getAllowedMethods != nil {
		if raw, err := getAllowedMethods(context.Background()); err == nil && raw != "" {
			var methods []string
			for _, m := range strings.Split(raw, ",") {
				if m = strings.TrimSpace(strings.ToUpper(m)); m != "" {
					methods = append(methods, m)
				}
			}
			if len(methods) > 0 {
				mcpClient.SetAllowedMethods(methods)
			}
		}
	}

	enabled := func() bool { return cfg.Agent.Enabled }
	apiKey := func(ctx context.Context) (string, error) {
		if getLLMKey != nil {
			if k, err := getLLMKey(ctx); err == nil && k != "" {
				return k, nil
			}
		}
		if env := cfg.Agent.APIKeyEnv; env != "" {
			if k := os.Getenv(env); k != "" {
				return k, nil
			}
		}
		if k := cfg.Agent.APIKey; k != "" {
			return k, nil
		}
		return "", nil
	}
	// 运行时读取系统提示词：配置有值用配置，为空回退默认常量。
	readSystemPrompt := func(ctx context.Context) (string, error) {
		if getSystemPrompt != nil {
			if p, err := getSystemPrompt(ctx); err == nil && p != "" {
				return p, nil
			}
		}
		return app.SystemPromptFallback, nil
	}

	svc := app.NewAgentService(repo, llmClient, mcpClient, readSystemPrompt, apiKey, enabled, selfBaseURL)
	return &AgentModule{
		handler: NewAgentHandler(svc),
		svc:     svc,
	}
}

// systemPrompt Agent 的系统提示词前缀（端点清单由 service 动态拼接）。
func (m *AgentModule) Name() string                       { return "agent" }
func (m *AgentModule) Init(ctx context.Context) error     { return nil }
func (m *AgentModule) Shutdown(ctx context.Context) error { return nil }
func (m *AgentModule) RegisterPublic(r *gin.RouterGroup)  {}

func (m *AgentModule) RegisterProtected(r *gin.RouterGroup) {
	g := r.Group("/agent")
	// 登录用户即可（agent:list 控制菜单可见性），chat 依赖会话归属校验。
	g.GET("/conversations", rbacTransport.RequirePerm("agent:list"), m.handler.ListConversations)
	g.POST("/conversations", rbacTransport.RequirePerm("agent:list"), m.handler.CreateConversation)
	g.GET("/conversations/:id/messages", rbacTransport.RequirePerm("agent:list"), m.handler.ListMessages)
	g.DELETE("/conversations/:id", rbacTransport.RequirePerm("agent:list"), m.handler.DeleteConversation)
	g.POST("/chat/stream", rbacTransport.RequirePerm("agent:list"), m.handler.ChatStream)
}

// ensure port interface satisfied
var _ port.AgentRepository = (*adapter.AgentRepo)(nil)

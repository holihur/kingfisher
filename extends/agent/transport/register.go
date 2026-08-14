package transport

import (
	"context"
	"os"

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
// cfg 提供 base_url/model/api_key 环境名等；configSvc 用于运行时读取
// system_configs.llm_api_key（后台可改）。selfBaseURL 为 call_api 内部请求目标。
func NewAgentModule(db *gorm.DB, cfg *config.Config, selfBaseURL string, getLLMKey func(ctx context.Context) (string, error)) *AgentModule {
	repo := adapter.NewAgentRepo(db)
	llmClient := llm.NewClient(cfg.Agent.BaseURL, "", cfg.Agent.Model, cfg.Agent.MaxTokens)
	mcpClient := mcp.NewClient(selfBaseURL)

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

	svc := app.NewAgentService(repo, llmClient, mcpClient, systemPrompt, apiKey, enabled, selfBaseURL)
	return &AgentModule{
		handler: NewAgentHandler(svc),
		svc:     svc,
	}
}

// systemPrompt Agent 的系统提示词前缀（端点清单由 service 动态拼接）。
const systemPrompt = `你是 Kingfisher 后台管理系统的 AI 助手，可以通过调用系统接口帮助用户查询和操作系统。
请始终用中文回答，简洁、准确。当需要查询系统数据（如用户、配置、审计、字典等）时，
使用 call_api 工具调用系统接口；接口返回后根据真实数据回答，不要编造。
如果接口返回 403 说明当前用户无权限，请如实告知。`

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

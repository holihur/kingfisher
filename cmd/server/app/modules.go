package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	"kingfisher/core/config"
	"kingfisher/core/jwt"
	"kingfisher/core/mailer"
	"kingfisher/core/taskqueue"

	"kingfisher/core/router"
	agentTransport "kingfisher/extends/agent/transport"
	auditTransport "kingfisher/extends/audit/transport"
	configAdapter "kingfisher/extends/config/adapter/mysql"
	configApp "kingfisher/extends/config/app"
	configTransport "kingfisher/extends/config/transport"
	departmentTransport "kingfisher/extends/department/transport"
	dictTransport "kingfisher/extends/dict/transport"
	docTransport "kingfisher/extends/doc/transport"
	emailTransport "kingfisher/extends/email/transport"
	menuTransport "kingfisher/extends/menu/transport"
	messageTransport "kingfisher/extends/message/transport"
	rbacApp "kingfisher/extends/rbac/app"
	rbacTransport "kingfisher/extends/rbac/transport"
	systemApp "kingfisher/extends/system/app"
	systemTransport "kingfisher/extends/system/transport"
	taskTransport "kingfisher/extends/task/transport"
	templateAdapter "kingfisher/extends/template/adapter/mysql"
	templateTransport "kingfisher/extends/template/transport"
	userTransport "kingfisher/extends/user/transport"
	worktaskTransport "kingfisher/extends/worktask/transport"
)

type moduleBundle struct {
	mods     []router.Module
	auditMod *auditTransport.AuditModule
	rbacSvc  *rbacApp.RoleService
}

func buildModules(db *gorm.DB, rdb *redis.Client, redisCache cache.Cache, jwtMgr *jwt.JWTManager, producer taskqueue.Producer, cfg *config.Config, zapLog *zap.Logger, vi VersionInfo) moduleBundle {
	auditMod := auditTransport.NewAuditModule(db)
	rbacSvc := rbacTransport.NewRoleService(db, redisCache)
	userMod := userTransport.NewUserModule(db, redisCache, jwtMgr, rbacSvc.GetUserPermissions, cfg.RateLimit.LoginPerMinute)
	userMod.InjectAuditLogger(userTransport.AuditLogger(auditMod.AuditLogCallback()))
	userMod.InjectAuditService(auditMod.Service())
	userMod.InjectLandingPageProvider(rbacSvc.GetRoleLandingPage)

	mailerInst := mailer.New(cfg.SMTP)
	emailMod := emailTransport.NewEmailModule(mailerInst, zapLog)
	emailProducer := emailTransport.NewEmailProducer(producer)

	configSvc := configApp.NewConfigService(configAdapter.NewConfigRepo(db), redisCache)
	userMod.InjectConfigProvider(func(ctx context.Context, key string) (string, error) {
		sc, err := configSvc.GetPublic(ctx, key)
		if err != nil {
			return "", err
		}
		return sc.Value, nil
	})
	userMod.InjectEmailSender(emailProducer.EnqueueEmail)
	tmplRepo := templateAdapter.NewTemplateRepo(db)
	userMod.InjectTemplateRenderer(func(ctx context.Context, code string, vars map[string]string) (string, string, error) {
		t, err := tmplRepo.GetByCode(ctx, code)
		if err != nil {
			return "", "", fmt.Errorf("template %s not found", code)
		}
		subject, body := t.Title, t.Content
		for k, v := range vars {
			subject = strings.ReplaceAll(subject, "{{"+k+"}}", v)
			body = strings.ReplaceAll(body, "{{"+k+"}}", v)
		}
		return subject, body, nil
	})

	agentSelfBaseURL := cfg.Agent.SelfBaseURL
	if agentSelfBaseURL == "" {
		agentSelfBaseURL = "http://127.0.0.1:" + fmt.Sprint(cfg.Server.Port)
	}
	agentMod := agentTransport.NewAgentModule(db, cfg, agentSelfBaseURL,
		func(ctx context.Context) (string, error) {
			sc, err := configSvc.Get(ctx, "llm_api_key")
			if err != nil {
				return "", err
			}
			return sc.Value, nil
		},
		func(ctx context.Context) (string, error) {
			sc, err := configSvc.Get(ctx, "agent_system_prompt")
			if err != nil {
				return "", err
			}
			return sc.Value, nil
		},
		func(ctx context.Context) (string, error) {
			sc, err := configSvc.Get(ctx, "agent_allowed_methods")
			if err != nil {
				return "", err
			}
			return sc.Value, nil
		},
	)

	mods := []router.Module{
		agentMod,
		userMod,
		departmentTransport.NewDepartmentModule(db, redisCache),
		rbacTransport.NewRBACModule(db, redisCache),
		menuTransport.NewMenuModule(db, redisCache),
		configTransport.NewConfigModule(db, redisCache),
		dictTransport.NewDictModule(db, redisCache),
		docTransport.NewDocModule(db, redisCache),
		messageTransport.NewMessageModule(db, producer),
		templateTransport.NewTemplateModule(db, redisCache),
		taskTransport.NewTaskModule(db, producer),
		worktaskTransport.NewModule(db, worktaskTransport.ScopeResolver(rbacSvc.ResolveDataScope)),
		systemTransport.NewSystemModule(db, rdb, systemApp.VersionInfo{
			Version: vi.Version, Commit: vi.Commit, BuildTime: vi.BuildTime,
		}),
		emailMod,
		auditMod,
	}
	return moduleBundle{mods: mods, auditMod: auditMod, rbacSvc: rbacSvc}
}

func BuildModulesForWire(db *gorm.DB, rdb *redis.Client, redisCache cache.Cache, jwtMgr *jwt.JWTManager, producer taskqueue.Producer, cfg *config.Config, zapLog *zap.Logger, vi VersionInfo) []router.Module {
	return buildModules(db, rdb, redisCache, jwtMgr, producer, cfg, zapLog, vi).mods
}

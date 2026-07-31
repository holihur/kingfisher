package app
import ("context"; "encoding/json"; "kingfisher/extends/config/domain"; "kingfisher/extends/config/adapter/mysql")
type ConfigService struct{ repo *adapter.ConfigRepo }
func NewConfigService(repo *adapter.ConfigRepo) *ConfigService { return &ConfigService{repo: repo} }
func (s *ConfigService) GetAll(ctx context.Context) ([]domain.SystemConfig, error) { return s.repo.GetAll(ctx) }
func (s *ConfigService) Get(ctx context.Context, key string) (*domain.SystemConfig, error) { return s.repo.GetByKey(ctx, key) }
func (s *ConfigService) Set(ctx context.Context, key, value string) error { return s.repo.Set(ctx, key, value) }
func (s *ConfigService) Delete(ctx context.Context, key string) error { return s.repo.Delete(ctx, key) }
var _ = json.Marshal

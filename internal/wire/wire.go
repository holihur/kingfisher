//go:build wireinject
// +build wireinject

package wire

import (
	"kingfisher/cmd/server/app"
	"kingfisher/core/cache"
	"kingfisher/core/config"
	"kingfisher/core/jwt"
	"kingfisher/core/router"

	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var InfraSet = wire.NewSet(
	cache.NewRedisCache,
	jwt.NewJWTManager,
)

var ModuleSet = wire.NewSet(
	app.BuildModulesForWire,
)

var ServerSet = wire.NewSet(
	router.NewEngine,
)

var _ = wire.NewSet(
	InfraSet,
	ModuleSet,
	ServerSet,
)

var _ = app.VersionInfo{}
var _ = config.Config{}
var _ = zap.Logger{}
var _ = gorm.DB{}
var _ = cache.Cache{}

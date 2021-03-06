//go:build wireinject
// +build wireinject

package app

import (
	"github.com/aserto-dev/go-utils/logger"
	"github.com/aserto-dev/idpsync/pkg/app/impl"
	"github.com/aserto-dev/idpsync/pkg/app/server"
	"github.com/aserto-dev/idpsync/pkg/cc"
	"github.com/aserto-dev/idpsync/pkg/cc/config"
	"github.com/google/wire"
)

var (
	appSet = wire.NewSet(
		cc.NewCC,

		GRPCServerRegistrations,
		GatewayServerRegistrations,
		server.NewServer,

		impl.NewIDPSync,

		wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup"),
	)

	appTestSet = wire.NewSet(
		// Test
		cc.NewTestCC,

		// Normal
		GRPCServerRegistrations,
		GatewayServerRegistrations,
		server.NewServer,

		impl.NewIDPSync,

		wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup"),
	)
)

func BuildIdpsync(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*App, func(), error) {
	wire.Build(
		wire.Struct(new(App), "*"),
		appSet,
	)
	return &App{}, func() {}, nil
}

func BuildTestIdpsync(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*App, func(), error) {
	wire.Build(
		wire.Struct(new(App), "*"),
		appTestSet,
	)
	return &App{}, func() {}, nil
}

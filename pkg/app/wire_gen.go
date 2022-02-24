// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package app

import (
	"github.com/aserto-dev/go-utils/logger"
	"github.com/aserto-dev/idpsync/pkg/app/impl"
	"github.com/aserto-dev/idpsync/pkg/app/server"
	"github.com/aserto-dev/idpsync/pkg/cc"
	"github.com/aserto-dev/idpsync/pkg/cc/config"
	"github.com/google/wire"
)

// Injectors from wire.go:

func BuildIdpsync(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*App, func(), error) {
	ccCC, cleanup, err := cc.NewCC(logOutput, errOutput, configPath, overrides)
	if err != nil {
		return nil, nil, err
	}
	context := ccCC.Context
	zerologLogger := ccCC.Log
	configConfig := ccCC.Config
	group := ccCC.ErrGroup
	idpSync := impl.NewIDPSync(zerologLogger, configConfig)
	registrations := GRPCServerRegistrations(idpSync)
	handlerRegistrations := GatewayServerRegistrations()
	serverServer, err := server.NewServer(context, configConfig, zerologLogger, group, registrations, handlerRegistrations)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	app := &App{
		Context:       context,
		Logger:        zerologLogger,
		Configuration: configConfig,
		Server:        serverServer,
	}
	return app, func() {
		cleanup()
	}, nil
}

func BuildTestIdpsync(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*App, func(), error) {
	ccCC, cleanup, err := cc.NewTestCC(logOutput, errOutput, configPath, overrides)
	if err != nil {
		return nil, nil, err
	}
	context := ccCC.Context
	zerologLogger := ccCC.Log
	configConfig := ccCC.Config
	group := ccCC.ErrGroup
	idpSync := impl.NewIDPSync(zerologLogger, configConfig)
	registrations := GRPCServerRegistrations(idpSync)
	handlerRegistrations := GatewayServerRegistrations()
	serverServer, err := server.NewServer(context, configConfig, zerologLogger, group, registrations, handlerRegistrations)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	app := &App{
		Context:       context,
		Logger:        zerologLogger,
		Configuration: configConfig,
		Server:        serverServer,
	}
	return app, func() {
		cleanup()
	}, nil
}

// wire.go:

var (
	appSet = wire.NewSet(cc.NewCC, GRPCServerRegistrations,
		GatewayServerRegistrations, server.NewServer, impl.NewIDPSync, wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup"),
	)

	appTestSet = wire.NewSet(cc.NewTestCC, GRPCServerRegistrations,
		GatewayServerRegistrations, server.NewServer, impl.NewIDPSync, wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup"),
	)
)

package app

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/aserto-dev/idpsync/pkg/app/server"
	"github.com/aserto-dev/idpsync/pkg/cc/config"
)

type App struct {
	Context       context.Context
	Logger        *zerolog.Logger
	Configuration *config.Config
	Server        *server.Server
}

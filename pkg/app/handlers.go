package app

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/aserto-dev/idpsync/api/idpsync/v1"
	"github.com/aserto-dev/idpsync/pkg/app/impl"
	"github.com/aserto-dev/idpsync/pkg/app/server"
	// infoconfig "github.com/aserto-dev/go-grpc-internal/aserto/common/info/v1"
)

// GRPCServerRegistrations is where we register implementations with the GRPC server
func GRPCServerRegistrations(implIDPSync *impl.IDPSync) server.Registrations {
	return func(server *grpc.Server) {
		idpsync.RegisterIDPSyncServer(server, implIDPSync)
	}
}

// GatewayServerRegistrations is where we register implementations with the Gateway server
func GatewayServerRegistrations() server.HandlerRegistrations {
	return func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {

		err := idpsync.RegisterIDPSyncHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		if err != nil {
			return errors.Wrap(err, "failed to register info handler with the gateway")
		}

		return nil
	}
}

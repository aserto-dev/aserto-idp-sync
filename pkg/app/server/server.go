package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/aserto-dev/go-utils/certs"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/aserto-dev/idpsync/pkg/cc/config"
)

const (
	svcName = "idpsync"
)

// Server manages the GRPC and HTTP servers, as well as their health servers.
type Server struct {
	ctx context.Context
	cfg *config.Config

	grpcServer   *grpc.Server
	gtwServer    *http.Server
	healthServer *HealthServer

	gtwMux               *runtime.ServeMux
	errGroup             *errgroup.Group
	logger               *zerolog.Logger
	handlerRegistrations HandlerRegistrations
}

// NewServer sets up a new server
func NewServer(ctx context.Context, cfg *config.Config, logger *zerolog.Logger, errGroup *errgroup.Group, registrations Registrations, handlerRegistrations HandlerRegistrations) (*Server, error) {
	grpcServer, err := newGRPCServer(cfg, logger, registrations)
	if err != nil {
		return nil, err
	}

	gtwMux := gatewayMux()
	gtwServer, err := newGatewayServer(logger, cfg, gtwMux)
	if err != nil {
		return nil, err
	}

	healthServer := newGRPCHealthServer()

	appServer := &Server{
		ctx:                  ctx,
		cfg:                  cfg,
		grpcServer:           grpcServer,
		gtwServer:            gtwServer,
		gtwMux:               gtwMux,
		healthServer:         healthServer,
		errGroup:             errGroup,
		logger:               logger,
		handlerRegistrations: handlerRegistrations,
	}

	return appServer, nil
}

// Start starts the GRPC and HTTP servers, as well as their health servers.
func (s *Server) Start() error {
	s.logger.Info().Msg("server::Start")

	grpc.EnableTracing = true

	// Health Server
	healthListener, err := net.Listen("tcp", s.cfg.API.Health.ListenAddress)
	if err != nil {
		s.logger.Error().Err(err).Str("address", s.cfg.API.Health.ListenAddress).Msg("grpc health socket failed to listen")
		return errors.Wrap(err, "grpc health socket failed to listen")
	}
	s.logger.Info().Str("address", s.cfg.API.Health.ListenAddress).Msg("GRPC Health Server starting")
	s.errGroup.Go(func() error {
		return s.healthServer.GRPCServer.Serve(healthListener)
	})

	// GRPC Server
	s.logger.Info().Str("address", s.cfg.API.GRPC.ListenAddress).Msg("GRPC Server starting")
	grpcListener, err := net.Listen("tcp", s.cfg.API.GRPC.ListenAddress)
	if err != nil {
		return errors.Wrap(err, "grpc socket failed to listen")
	}
	s.errGroup.Go(func() error {
		err := s.grpcServer.Serve(grpcListener)
		if err != nil {
			s.logger.Error().Err(err).Str("address", s.cfg.API.GRPC.ListenAddress).Msg("GRPC Server failed to listen")
		}
		return errors.Wrap(err, "grpc server failed to listen")
	})

	// OpenAPI Gateway HTTP Server
	s.logger.Info().Msg("Registering OpenAPI Gateway handlers")
	err = s.registerGateway()
	if err != nil {
		return errors.Wrap(err, "failed to register grpc gateway handlers")
	}
	s.logger.Info().
		Str("address", "https://"+s.cfg.API.Gateway.ListenAddress).
		Msg("gRPC-Gateway and OpenAPI endpoint starting")
	s.errGroup.Go(func() error {
		return s.gtwServer.ListenAndServeTLS("", "")
	})

	s.healthServer.Server.SetServingStatus(fmt.Sprintf("grpc.health.v1.%s", svcName), healthpb.HealthCheckResponse_SERVING)

	return nil
}

// Stop stops the GRPC and HTTP servers, as well as their health servers.
func (s *Server) Stop() error {
	var result error

	s.logger.Info().Msg("Server stopping.")

	ctx, shutdownCancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer shutdownCancel()

	if s.gtwServer != nil {
		err := s.gtwServer.Shutdown(ctx)
		if err != nil {
			if err == context.Canceled {
				s.logger.Info().Msg("server context was canceled - shutting down")
			} else {
				result = multierror.Append(result, errors.Wrap(err, "failed to stop gateway server"))
			}
		}
	}

	if s.healthServer != nil {
		s.healthServer.Server.SetServingStatus(
			fmt.Sprintf("grpc.health.v1.%s", svcName),
			healthpb.HealthCheckResponse_NOT_SERVING,
		)
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	if s.healthServer.GRPCServer != nil {
		s.healthServer.GRPCServer.GracefulStop()
	}

	err := s.errGroup.Wait()
	if err != nil {
		s.logger.Info().Err(err).Msg("shutdown complete")
	}

	return result
}

func (s *Server) registerGateway() error {
	_, port, err := net.SplitHostPort(s.cfg.API.GRPC.ListenAddress)
	if err != nil {
		return errors.Wrap(err, "failed to determine port from configured GRPC listen address")
	}

	dialAddr := fmt.Sprintf("dns:///127.0.0.1:%s", port)

	tlsCreds, err := certs.GatewayAsClientTLSCreds(s.cfg.API.GRPC.Certs)
	if err != nil {
		return errors.Wrap(err, "failed to calculate tls config for gateway service")
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithBlock(),
		grpc.WithTimeout(2 * time.Second), // nolint:staticcheck // using context.WithTimeout makes us unable to call defer ctx.Cancel
	}

	err = s.handlerRegistrations(s.ctx, s.gtwMux, dialAddr, opts)
	if err != nil {
		return errors.Wrap(err, "failed to register handlers with the gateway")
	}

	return nil
}

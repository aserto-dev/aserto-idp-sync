package impl

import (
	"context"

	aserto "github.com/aserto-dev/aserto-go/client"
	client "github.com/aserto-dev/aserto-go/client/grpc"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/authorizer/directory/v1"
	"github.com/aserto-dev/idpsync/api/idpsync/v1"
	"github.com/aserto-dev/idpsync/pkg/auth0"
	"github.com/aserto-dev/idpsync/pkg/cc/config"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rs/zerolog"
)

type IDPSync struct {
	logger *zerolog.Logger
	cfg    *config.Config
}

func NewIDPSync(logger *zerolog.Logger, cfg *config.Config) *IDPSync {
	serviceLogger := logger.With().Str("component", "impl.sync").Logger()

	return &IDPSync{
		logger: &serviceLogger,
		cfg:    cfg,
	}
}

func (s *IDPSync) SyncUser(ctx context.Context, req *idpsync.SyncUserRequest) (*idpsync.SyncUserResponse, error) {
	if req.EmailAddress == "" {
		return &idpsync.SyncUserResponse{}, status.Error(codes.InvalidArgument, "empty email")
	}

	s.logger.Info().Str("tenantID", s.cfg.Directory.TenantID).Str("email", req.EmailAddress).Msg("SyncUser")

	mgr := auth0.NewManager(&s.cfg.IDP.Auth0)
	if err := mgr.Init(); err != nil {
		return &idpsync.SyncUserResponse{}, errors.Wrapf(err, "auth0 init")
	}

	auth0User, err := mgr.GetUserFromEmail(req.EmailAddress)
	if err != nil {
		return &idpsync.SyncUserResponse{}, errors.Wrapf(err, "auth0 get user from email [%s]", req.EmailAddress)
	}

	user, err := auth0.Transform(auth0User)
	if err != nil {
		return &idpsync.SyncUserResponse{}, errors.Wrapf(err, "transform auth0 to aserto user [%s]", req.EmailAddress)
	}

	newUser, err := s.upsert(req.EmailAddress, user)
	if err != nil {
		return &idpsync.SyncUserResponse{}, errors.Wrapf(err, "upsert user %s", req.EmailAddress)
	}

	s.logger.Info().Str("tenantID", s.cfg.Directory.TenantID).Str("email", req.EmailAddress).Str("id", newUser.Id).Msg("SyncUser")

	return &idpsync.SyncUserResponse{}, nil
}

func (s *IDPSync) upsert(email string, user *api.User) (*api.User, error) {
	ctx := context.Background()

	c, err := client.New(
		ctx,
		aserto.WithAPIKeyAuth(s.cfg.Directory.DirectoryAPIKey),
		aserto.WithTenantID(s.cfg.Directory.TenantID),
	)
	if err != nil {
		return &api.User{}, errors.Wrapf(err, "create gRPC directory connection")
	}

	identity, err := getVerifiedIdentity(user)
	if err != nil {
		return &api.User{}, errors.Wrapf(err, "no verified identity found for user [%s]", email)
	}

	identResp, err := c.Directory.GetIdentity(ctx, &directory.GetIdentityRequest{
		Identity: identity,
	})

	var newUser *api.User
	if err != nil || identResp == nil || identResp.Id == "" {
		if s.cfg.Directory.GenerateUserID {
			user.Id = "" // Providing an empty user.ID will make directory.CreateUser() generate a stable ID (lowercase uuid) for the user
		}

		resp, err := c.Directory.CreateUser(ctx, &directory.CreateUserRequest{
			User: user,
		})
		if err != nil {
			return &api.User{}, errors.Wrapf(err, "create user [%s]", email)
		}
		newUser = resp.Result
	} else {
		user.Id = identResp.Id
		resp, err := c.Directory.UpdateUser(ctx, &directory.UpdateUserRequest{
			Id:   identResp.Id,
			User: user,
		})
		if err != nil {
			return &api.User{}, errors.Wrapf(err, "update user [%s]", email)
		}
		newUser = resp.Result
	}

	return newUser, nil
}

func getVerifiedIdentity(user *api.User) (string, error) {
	for identity, identitySource := range user.Identities {
		if identitySource.Verified {
			return identity, nil
		}
	}
	return "", errors.Errorf("no verified identity found")
}

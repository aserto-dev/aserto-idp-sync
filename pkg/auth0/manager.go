package auth0

import (
	"github.com/pkg/errors"
	"gopkg.in/auth0.v5/management"
)

type Config struct {
	Domain       string `json:"domain"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type Manager struct {
	config *Config
	mgnt   *management.Management
}

// NewManager Auth0 management interface.
func NewManager(cfg *Config) *Manager {
	manager := Manager{
		config: cfg,
	}
	return &manager
}

var (
	ErrNotFound  = errors.New("user not found")
	ErrAmbiguous = errors.New("ambiguous result")
)

// Init initialize management connection.
func (m *Manager) Init() error {
	mgnt, err := management.New(
		m.config.Domain,
		management.WithClientCredentials(
			m.config.ClientID,
			m.config.ClientSecret,
		),
	)
	if err != nil {
		return errors.Wrapf(err, "create management instance")
	}

	m.mgnt = mgnt

	return nil
}

func (m *Manager) GetUserFromEmail(email string) (*management.User, error) {
	users, err := m.mgnt.User.ListByEmail(email)
	switch {
	case err != nil:
		return nil, err
	case len(users) == 0:
		return nil, errors.Wrapf(ErrNotFound, "email [%s]", email)
	case len(users) == 1:
		return users[0], nil
	default:
		return nil, errors.Wrapf(ErrAmbiguous, "[%d] results found for email [%s]", len(users), email)
	}
}

// Package mail implements Connection/Account management and mailcap Service
// for linapro-mail-core. Persistence uses plugin-owned tables; transports are
// resolved through mailcap/spi.
package mail

import (
	"context"

	"lina-core/pkg/plugin/capability/plugincap"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
	"lina-plugin-linapro-mail-core/backend/internal/model/entity"
)

// Status values for connection and account rows.
const (
	// StatusDisabled marks a disabled record.
	StatusDisabled = 0
	// StatusEnabled marks an enabled record.
	StatusEnabled = 1
	// FlagDefault marks the default account.
	FlagDefault = 1
	// FlagNotDefault marks a non-default account.
	FlagNotDefault = 0
)

// TLS and auth mode tokens stored on Connection.
const (
	// TLSModeDisable disables TLS.
	TLSModeDisable = "disable"
	// TLSModeStartTLS uses STARTTLS.
	TLSModeStartTLS = "starttls"
	// TLSModeTLS uses implicit TLS.
	TLSModeTLS = "tls"
	// AuthModePassword uses username/password.
	AuthModePassword = "password"
)

// Service manages connections, accounts, and mailcap send/fetch/probe.
type Service interface {
	mailcap.Service

	// ListConnections returns one page of connections.
	ListConnections(ctx context.Context, in ListConnectionsInput) (*ListConnectionsOutput, error)
	// GetConnection returns one connection by ID.
	GetConnection(ctx context.Context, id int64) (*entity.Connection, error)
	// CreateConnection creates one connection.
	CreateConnection(ctx context.Context, in CreateConnectionInput) (int64, error)
	// UpdateConnection updates one connection.
	UpdateConnection(ctx context.Context, in UpdateConnectionInput) error
	// DeleteConnections soft-deletes connections by IDs.
	DeleteConnections(ctx context.Context, ids []int64) error

	// ListAccounts returns one page of accounts.
	ListAccounts(ctx context.Context, in ListAccountsInput) (*ListAccountsOutput, error)
	// GetAccount returns one account by ID.
	GetAccount(ctx context.Context, id int64) (*entity.Account, error)
	// CreateAccount creates one account.
	CreateAccount(ctx context.Context, in CreateAccountInput) (int64, error)
	// UpdateAccount updates one account.
	UpdateAccount(ctx context.Context, in UpdateAccountInput) error
	// DeleteAccounts soft-deletes accounts by IDs.
	DeleteAccounts(ctx context.Context, ids []int64) error
	// ResolveAccount resolves explicit or default account ID.
	ResolveAccount(ctx context.Context, accountID int64) (*entity.Account, error)

	// GetPlatformSettings returns the masked single platform account projection.
	GetPlatformSettings(ctx context.Context) (*PlatformSettingsProjection, error)
	// SavePlatformSettings upserts the single platform account and connections.
	SavePlatformSettings(ctx context.Context, in PlatformSettingsInput) (*PlatformSettingsProjection, error)
	// TestPlatformSettings probes SMTP and optional inbound without persisting.
	// On success it returns OK=true. Probe/validation failures return a structured
	// bizerr so the host Response middleware can localize messageKey for the UI.
	TestPlatformSettings(ctx context.Context, in PlatformSettingsInput) (*PlatformSettingsTestResult, error)
	// SendPlatformTestMail sends one diagnostic message through the form SMTP endpoint.
	// It uses current form values (with empty password resolved from stored secret when present)
	// and MUST NOT silently switch to a different saved outbound connection when form SMTP differs.
	// On success it returns OK=true. Transport/validation failures return a structured bizerr
	// (not an English-only soft Message) so the host can localize the user-visible reason.
	SendPlatformTestMail(ctx context.Context, in PlatformSettingsInput, to, subject, body string) (*PlatformSettingsTestResult, error)
	// TestPlatformReceive probes inbound receive capability through the form IMAP/POP3 endpoint.
	// It uses current form values (with empty password resolved from stored secret when present)
	// and MUST NOT silently switch to a different saved inbound connection when form inbound differs.
	// inboundKind none/empty is rejected. On success it returns OK=true. Transport/validation
	// failures return a structured bizerr so the host can localize the user-visible reason.
	TestPlatformReceive(ctx context.Context, in PlatformSettingsInput) (*PlatformSettingsTestResult, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	pluginState plugincap.StateService
}

// New creates one mail domain Service.
func New(pluginState plugincap.StateService) Service {
	return &serviceImpl{pluginState: pluginState}
}

// ListConnectionsInput constrains connection listing.
type ListConnectionsInput struct {
	PageNum  int
	PageSize int
	Name     string
	Kind     string
	Status   *int
}

// ListConnectionsOutput is one connection page.
type ListConnectionsOutput struct {
	List  []*entity.Connection
	Total int
}

// CreateConnectionInput creates one connection.
type CreateConnectionInput struct {
	Name      string
	Kind      string
	Host      string
	Port      int
	Username  string
	SecretRef string
	TLSMode   string
	AuthMode  string
	ExtraJSON string
	Status    int
	Remark    string
}

// UpdateConnectionInput updates one connection.
type UpdateConnectionInput struct {
	ID        int64
	Name      string
	Kind      string
	Host      string
	Port      int
	Username  string
	SecretRef string
	TLSMode   string
	AuthMode  string
	ExtraJSON string
	Status    int
	Remark    string
}

// ListAccountsInput constrains account listing.
type ListAccountsInput struct {
	PageNum  int
	PageSize int
	Name     string
	Status   *int
}

// ListAccountsOutput is one account page.
type ListAccountsOutput struct {
	List  []*entity.Account
	Total int
}

// CreateAccountInput creates one account.
type CreateAccountInput struct {
	Name                 string
	FromAddress          string
	OutboundConnectionID int64
	InboundConnectionID  int64
	IsDefault            bool
	Status               int
	Remark               string
}

// UpdateAccountInput updates one account.
type UpdateAccountInput struct {
	ID                   int64
	Name                 string
	FromAddress          string
	OutboundConnectionID int64
	InboundConnectionID  int64
	IsDefault            bool
	Status               int
	Remark               string
}

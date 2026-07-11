// Package extidcap is the plugin-owned external-identity domain capability
// published by linapro-extid-core. Protocol plugins (Google, Discord, future
// WeChat/QQ/Douyin adapters) and host-facing HTTP controllers consume this
// contract. Token minting stays on the host authcap.ExternalLogin seam.
//
// This package only publishes contracts, DTOs, error semantics, and thin
// process-bound facades. Runtime stores (handoff, provider catalog) live in
// backend/internal/service/* and are bound by the owner plugin at startup.
//
// Service is a wide entry that aggregates role-oriented sub surfaces (Ticket,
// Login, Linkage, Providers). Catalog registration and SPA handoff stay on the
// separate CatalogService / HandoffService facades.
package extidcap

import (
	"context"
	"time"

	"lina-core/pkg/plugin/capability/authcap/extlogin/extidspi"
)

// OwnerPluginID is the fixed owner plugin id for this domain capability.
const OwnerPluginID = "linapro-extid-core"

// Service is the wide external-identity domain entry. Consumers should depend on
// this interface (or a sub surface obtained from it), not on internal packages.
// Stable sub capabilities are exposed via Ticket / Login / Linkage / Providers,
// matching host capability aggregation style (for example orgcap Assignment()).
type Service interface {
	// Ticket returns verified-identity ticket operations used by protocol plugins.
	Ticket() TicketService
	// Login returns resolve-or-provision orchestration before host session minting.
	Login() LoginService
	// Linkage returns bind/unbind/list/profile operations for identity links.
	Linkage() LinkageService
	// Providers returns catalog query and provision-policy views with domain errors.
	// Protocol plugins register descriptors through CatalogService, not here.
	Providers() ProviderService
}

// TicketService stores short-lived protocol-verified identities for login or bind.
type TicketService interface {
	// IssueVerifiedTicket stores one protocol-verified identity and returns a
	// short-lived ticket id. Identity.Provider and Identity.Subject are required.
	IssueVerifiedTicket(ctx context.Context, identity VerifiedIdentity) (*TicketIssueResult, error)
	// PeekVerifiedTicket returns a ticket without consuming it. Missing or
	// expired tickets return a bizerr ticket-invalid code.
	PeekVerifiedTicket(ctx context.Context, ticketID string) (*VerifiedIdentity, error)
	// ConsumeVerifiedTicket returns and invalidates a ticket. Second consume fails.
	ConsumeVerifiedTicket(ctx context.Context, ticketID string) (*VerifiedIdentity, error)
	// InvalidateVerifiedTicket drops a ticket if present. Missing tickets are not errors.
	InvalidateVerifiedTicket(ctx context.Context, ticketID string) error
}

// LoginService orchestrates resolve-or-provision before the host mints a session.
type LoginService interface {
	// Resolve maps a verified (provider, subject) pair to a linked local user.
	// found=false without error means not linked.
	Resolve(ctx context.Context, in extidspi.ResolveInput) (userID int, found bool, err error)
	// LoginPrepare resolves or provisions from a ticket or in-process identity.
	// When DryRun is true it never writes or consumes tickets (preview path).
	LoginPrepare(ctx context.Context, in LoginPrepareInput) (*LoginPrepareResult, error)
	// PreviewBind reports whether a ticket can bind to the given user without writing.
	PreviewBind(ctx context.Context, userID int, ticketID string) (*BindPreviewResult, error)
}

// LinkageService manages external-identity links for a local user.
// Authorization (session self-isolation vs operator paths) is enforced by the
// call site; this surface does not expose a separate Admin directory.
type LinkageService interface {
	// BindByTicket consumes a verified ticket and links it to userID.
	BindByTicket(ctx context.Context, userID int, ticketID string) error
	// Unbind removes one linkage for the session user path via host SPI input.
	Unbind(ctx context.Context, in extidspi.UnbindInput) error
	// GetLinkage returns one linkage by authoritative (provider, subject) key.
	// Missing linkage returns (nil, nil).
	GetLinkage(ctx context.Context, provider, subject string) (*LinkageDetail, error)
	// ListByUser returns all linkages for one user as domain details.
	ListByUser(ctx context.Context, userID int) ([]LinkageDetail, error)
	// ListByProvider returns a page of linkages for one provider (operator projection).
	ListByProvider(ctx context.Context, provider string, page, pageSize int) ([]LinkageDetail, int, error)
	// SyncProfile refreshes snapshots from a ticket without changing the subject key.
	SyncProfile(ctx context.Context, ticketID string) error
}

// ProviderService queries the process-local protocol catalog and provision policy.
// Registration remains on CatalogService so init-time protocol plugins can buffer
// descriptors before the owner plugin binds.
type ProviderService interface {
	// ListProviders returns the bound process-local protocol catalog.
	ListProviders(ctx context.Context, filter ProviderFilter) ([]ProviderView, error)
	// GetProvider returns one catalog entry or a not-found bizerr.
	GetProvider(ctx context.Context, providerID string) (*ProviderView, error)
	// GetProvisionPolicy returns the per-provider auto-provision policy view.
	GetProvisionPolicy(ctx context.Context, providerID string) (*ProvisionPolicy, error)
}

// VerifiedIdentity is the unified model produced by protocol plugins after IdP
// verification. It is never trusted from unauthenticated browser input without
// a ticket or host-stamped plugin path.
type VerifiedIdentity struct {
	Provider          string
	Subject           string
	SubjectKind       extidspi.SubjectKind
	SecondarySubjects []extidspi.SecondarySubject
	AppContext        string
	Email             string
	EmailVerified     bool
	Phone             string
	PhoneVerified     bool
	DisplayName       string
	AvatarURL         string
	Locale            string
	PluginID          string
	IssuedAt          time.Time
	ExpiresAt         time.Time
	AssuranceLevel    string
}

// TicketIssueResult is returned when a verified identity is ticketed.
type TicketIssueResult struct {
	TicketID  string
	ExpiresAt time.Time
}

// LoginPrepareInput drives resolve-or-provision orchestration.
type LoginPrepareInput struct {
	// TicketID when set is consumed unless DryRun is true (then Peek only).
	TicketID string
	// Identity is used when TicketID is empty (server-side protocol plugin path).
	Identity *VerifiedIdentity
	// AllowAutoProvision permits automatic account creation.
	AllowAutoProvision bool
	// DryRun when true never writes or consumes tickets (login preview).
	DryRun bool
}

// LoginPrepareOutcome classifies LoginPrepare results.
type LoginPrepareOutcome string

const (
	// LoginOutcomeLinked means the identity already maps to a local user.
	LoginOutcomeLinked LoginPrepareOutcome = "linked"
	// LoginOutcomeProvisioned means a local user was created (or would be in DryRun).
	LoginOutcomeProvisioned LoginPrepareOutcome = "provisioned"
	// LoginOutcomeNeedsBind means no linkage and auto-provision is not allowed.
	LoginOutcomeNeedsBind LoginPrepareOutcome = "needs_bind"
	// LoginOutcomeConflictEmail means same-email local account blocks provision.
	LoginOutcomeConflictEmail LoginPrepareOutcome = "conflict_email"
	// LoginOutcomeConflictSubject means the subject is already bound elsewhere.
	LoginOutcomeConflictSubject LoginPrepareOutcome = "conflict_subject"
	// LoginOutcomeUnavailable means the provider or identity path is disabled.
	LoginOutcomeUnavailable LoginPrepareOutcome = "provider_disabled"
)

// LoginPrepareResult is the domain conclusion before host session minting.
type LoginPrepareResult struct {
	Outcome LoginPrepareOutcome
	UserID  int
	Linkage *LinkageDetail
}

// BindPreviewReason classifies why a bind preview failed.
type BindPreviewReason string

const (
	// BindPreviewInvalidUser means the target user id is invalid.
	BindPreviewInvalidUser BindPreviewReason = "invalid_user"
	// BindPreviewTicketInvalid means the ticket is missing, expired, or consumed.
	BindPreviewTicketInvalid BindPreviewReason = "ticket_invalid"
	// BindPreviewConflict means the subject is linked to another user.
	BindPreviewConflict BindPreviewReason = "conflict"
)

// BindPreviewResult describes whether a ticket can bind to a user.
type BindPreviewResult struct {
	OK      bool
	Reason  BindPreviewReason
	Linkage *LinkageDetail
}

// LinkageDetail is one external-identity link projection.
type LinkageDetail struct {
	Provider    string
	Subject     string
	SubjectKind extidspi.SubjectKind
	AppContext  string
	UserID      int
	PluginID    string
	Email       string
	Phone       string
	DisplayName string
	AvatarURL   string
}

// ProviderCapabilities describes what one protocol provider supports.
type ProviderCapabilities struct {
	Login              bool
	Bind               bool
	Unbind             bool
	AutoProvision      bool
	Email              bool
	Phone              bool
	Avatar             bool
	OneTap             bool
	QRCode             bool
	MiniProgram        bool
	RefreshableProfile bool
}

// SubjectStrategy names how a protocol plugin derives the authoritative subject.
type SubjectStrategy string

const (
	// SubjectStrategyOIDCSub uses the OIDC "sub" claim as the subject.
	SubjectStrategyOIDCSub SubjectStrategy = "oidc_sub"
)

// ConflictPolicy classifies provision-time conflict handling.
type ConflictPolicy string

const (
	// ConflictPolicyRequireBind rejects silent link/provision and requires bind.
	ConflictPolicyRequireBind ConflictPolicy = "require_bind"
	// ConflictPolicyReject hard-rejects the operation.
	ConflictPolicyReject ConflictPolicy = "reject"
)

// UsernameAnchorStrategy names deterministic username derivation for provision.
type UsernameAnchorStrategy string

const (
	// UsernameAnchorSHA256ProviderSubject derives anchors from sha256(provider+subject).
	UsernameAnchorSHA256ProviderSubject UsernameAnchorStrategy = "sha256_provider_subject"
)

// ProviderDescriptor is registered by protocol plugins at declaration/runtime.
type ProviderDescriptor struct {
	ID              string
	DisplayName     string
	Icon            string
	Order           int
	PluginID        string
	Protocols       []string
	Capabilities    ProviderCapabilities
	SubjectStrategy SubjectStrategy
	LoginEntryPath  string
}

// ProviderView is the runtime catalog projection.
//
// Without a persistent admin configuration store, Enabled and Configured are
// both true for every registered descriptor (registration presence).
type ProviderView struct {
	ProviderDescriptor
	Enabled    bool
	Configured bool
}

// ProviderFilter selects catalog rows.
type ProviderFilter struct {
	EnabledOnly    bool
	ConfiguredOnly bool
	SupportsBind   bool
	SupportsLogin  bool
}

// ProvisionPolicy is the per-provider auto-provision policy view.
type ProvisionPolicy struct {
	Provider               string
	AllowAutoProvision     bool
	RequireEmail           bool
	RequirePhone           bool
	OnEmailConflict        ConflictPolicy
	OnSubjectConflict      ConflictPolicy
	UsernameAnchorStrategy UsernameAnchorStrategy
	SubjectStrategy        SubjectStrategy
}

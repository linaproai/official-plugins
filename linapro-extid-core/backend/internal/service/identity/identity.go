// Package identity implements the linapro-extid-core external-identity provider:
// (provider, subject) linkage storage, host-delegated least-privilege account
// provisioning policy, and current-user bind/unbind/list operations.
//
// Data permission boundary (self-isolation exception, see design D3): login
// resolution uses only the authoritative (provider, subject) unique key and
// reports a uniform not-found outcome, never leaking whether an email exists on
// another account. Bind, Unbind, and List act exclusively on the current
// session user's own linkages; cross-user targets are rejected as a whole.
//
// Transaction contract (design D6/D8): host account provisioning and the plugin
// linkage write cannot share one database transaction across module boundaries.
// Correctness converges on the (provider, subject) partial unique index: the
// account is provisioned first, then the linkage is inserted; a unique-index
// conflict from a concurrent provision is absorbed by re-resolving and reusing
// the winning linkage instead of surfacing a 500. Username-anchor derivation is
// deterministic per (provider, subject), so a losing provision attempt reuses
// the same host account rather than minting duplicates.
package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/authcap/extlogin/extidspi"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-plugin-linapro-extid-core/backend/cap/extidcap"
	"lina-plugin-linapro-extid-core/backend/internal/dao"
	"lina-plugin-linapro-extid-core/backend/internal/model/do"
	"lina-plugin-linapro-extid-core/backend/internal/model/entity"
)

// usernameAnchorPrefix namespaces derived username anchors so provisioned
// accounts are recognizable in host user management.
const usernameAnchorPrefix = "oidc-"

// usernameAnchorHashLength is the hex length kept from the identity digest. 16
// hex characters (64 bits) are collision-resistant for account-scale
// cardinality while keeping the derived username within the host's 30-character
// provisioning budget.
const usernameAnchorHashLength = 16

// Service implements the extidspi.Provider contract and the plugin-owned
// extidcap.Service domain surface (Ticket / Login / Linkage / Providers) on top
// of the user_external_identity table and the host user capability.
type Service struct {
	users   usercap.Service
	tickets *ticketStore
}

// Ensure Service implements the host external-identity provider SPI and the
// plugin-owned domain capability (wide entry plus role-oriented sub surfaces).
var (
	_ extidspi.Provider        = (*Service)(nil)
	_ extidcap.Service         = (*Service)(nil)
	_ extidcap.TicketService   = (*Service)(nil)
	_ extidcap.LoginService    = (*Service)(nil)
	_ extidcap.LinkageService  = (*Service)(nil)
	_ extidcap.ProviderService = (*Service)(nil)
)

// Ticket returns verified-identity ticket operations.
func (s *Service) Ticket() extidcap.TicketService {
	if s == nil {
		return nil
	}
	return s
}

// Login returns resolve-or-provision orchestration.
func (s *Service) Login() extidcap.LoginService {
	if s == nil {
		return nil
	}
	return s
}

// Linkage returns bind/unbind/list/profile operations.
func (s *Service) Linkage() extidcap.LinkageService {
	if s == nil {
		return nil
	}
	return s
}

// Providers returns catalog query and provision-policy views.
func (s *Service) Providers() extidcap.ProviderService {
	if s == nil {
		return nil
	}
	return s
}

// New creates the linapro-extid-core external-identity provider from the
// injected host user capability.
func New(users usercap.Service) (*Service, error) {
	if users == nil {
		return nil, gerror.New("linapro-extid-core identity provider requires host user capability")
	}
	return &Service{
		users:   users,
		tickets: newTicketStore(),
	}, nil
}

// Resolve maps a verified (provider, subject) pair to a linked local user ID
// through the authoritative partial unique index. Missing linkage returns
// found=false without error so the host keeps a uniform not-provisioned
// outcome.
func (s *Service) Resolve(ctx context.Context, in extidspi.ResolveInput) (int, bool, error) {
	provider := strings.TrimSpace(in.Provider)
	subject := strings.TrimSpace(in.Subject)
	if provider == "" || subject == "" {
		return 0, false, bizerr.NewCode(CodeIdentityInvalid)
	}
	linkage, err := s.findLinkage(ctx, provider, subject)
	if err != nil {
		return 0, false, err
	}
	if linkage == nil {
		return 0, false, nil
	}
	return linkage.UserId, true, nil
}

// Provision runs the plugin-owned auto-provisioning policy for one unlinked
// verified identity: idempotent (provider, subject) reuse, same-email conflict
// rejection, email-less deterministic anchor derivation, and host-delegated
// least-privilege account creation followed by the linkage write.
func (s *Service) Provision(ctx context.Context, in extidspi.ProvisionInput) (int, error) {
	provider := strings.TrimSpace(in.Provider)
	subject := strings.TrimSpace(in.Subject)
	if provider == "" || subject == "" {
		return 0, bizerr.NewCode(CodeIdentityInvalid)
	}
	if !in.AllowAutoProvision {
		return 0, bizerr.NewCode(CodeProvisionNotAllowed)
	}
	if s == nil || s.users == nil {
		return 0, bizerr.NewCode(CodeUserCapabilityUnavailable)
	}
	// Idempotent fast path: a concurrent or earlier provision already linked
	// this identity.
	if linkage, err := s.findLinkage(ctx, provider, subject); err != nil {
		return 0, err
	} else if linkage != nil {
		return linkage.UserId, nil
	}
	email := strings.TrimSpace(in.Email)
	anchor := strings.TrimSpace(in.UsernameAnchor)
	if email == "" && anchor == "" {
		// Email-less provisioning derives a deterministic, collision-resistant
		// anchor from the identity key so the same identity always reaches the
		// same derived username, and distinct identities cannot alias.
		anchor = deriveUsernameAnchor(provider, subject)
	}
	userID, err := s.provisionAccount(ctx, provider, email, anchor, in.DisplayName)
	if err != nil {
		return 0, err
	}
	if _, err = dao.UserExternalIdentity.Ctx(ctx).Data(do.UserExternalIdentity{
		UserId:              userID,
		Provider:            provider,
		Subject:             subject,
		SubjectKind:         subjectKindOrDefault(in.SubjectKind),
		AppContext:          strings.TrimSpace(in.AppContext),
		PluginId:            strings.TrimSpace(in.PluginID),
		EmailSnapshot:       email,
		PhoneSnapshot:       strings.TrimSpace(in.Phone),
		DisplayNameSnapshot: strings.TrimSpace(in.DisplayName),
		AvatarUrlSnapshot:   strings.TrimSpace(in.AvatarURL),
	}).Insert(); err != nil {
		// A unique-index conflict means a concurrent provision won the race.
		// Reuse the winning linkage idempotently instead of bubbling a 500; a
		// losing provision attempt reused the same deterministic account, so no
		// duplicate account is minted.
		if linkage, findErr := s.findLinkage(ctx, provider, subject); findErr == nil && linkage != nil {
			return linkage.UserId, nil
		}
		return 0, bizerr.WrapCode(err, CodeIdentityWriteFailed)
	}
	return userID, nil
}

// Bind links a verified external identity to the current session user. It only
// acts on the current user: a (provider, subject) already owned by another
// account is rejected with a conflict, and re-binding an identity the current
// user already owns succeeds idempotently.
func (s *Service) Bind(ctx context.Context, in extidspi.BindInput) error {
	provider := strings.TrimSpace(in.Provider)
	subject := strings.TrimSpace(in.Subject)
	if in.UserID <= 0 || provider == "" || subject == "" {
		return bizerr.NewCode(CodeIdentityInvalid)
	}
	linkage, err := s.findLinkage(ctx, provider, subject)
	if err != nil {
		return err
	}
	if linkage != nil {
		if linkage.UserId == in.UserID {
			return nil
		}
		return bizerr.NewCode(CodeBindConflict)
	}
	if _, err = dao.UserExternalIdentity.Ctx(ctx).Data(do.UserExternalIdentity{
		UserId:              in.UserID,
		Provider:            provider,
		Subject:             subject,
		SubjectKind:         subjectKindOrDefault(in.SubjectKind),
		AppContext:          strings.TrimSpace(in.AppContext),
		PluginId:            strings.TrimSpace(in.PluginID),
		EmailSnapshot:       strings.TrimSpace(in.Email),
		PhoneSnapshot:       strings.TrimSpace(in.Phone),
		DisplayNameSnapshot: strings.TrimSpace(in.DisplayName),
		AvatarUrlSnapshot:   strings.TrimSpace(in.AvatarURL),
	}).Insert(); err != nil {
		// Absorb a concurrent bind race through the unique index: same-user
		// winner keeps idempotent success, another owner reports conflict.
		if existing, findErr := s.findLinkage(ctx, provider, subject); findErr == nil && existing != nil {
			if existing.UserId == in.UserID {
				return nil
			}
			return bizerr.NewCode(CodeBindConflict)
		}
		return bizerr.WrapCode(err, CodeIdentityWriteFailed)
	}
	return nil
}

// Unbind removes one of the current session user's external-identity links. A
// linkage that does not belong to the current user reports not-found without
// leaking other accounts' linkage existence. Soft deletion frees the partial
// unique index for a future relink.
func (s *Service) Unbind(ctx context.Context, in extidspi.UnbindInput) error {
	provider := strings.TrimSpace(in.Provider)
	subject := strings.TrimSpace(in.Subject)
	if in.UserID <= 0 || provider == "" || subject == "" {
		return bizerr.NewCode(CodeIdentityInvalid)
	}
	result, err := dao.UserExternalIdentity.Ctx(ctx).Where(do.UserExternalIdentity{
		UserId:   in.UserID,
		Provider: provider,
		Subject:  subject,
	}).Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeIdentityWriteFailed)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return bizerr.WrapCode(err, CodeIdentityWriteFailed)
	}
	if affected == 0 {
		return bizerr.NewCode(CodeIdentityNotFound)
	}
	return nil
}

// List returns the current session user's bound external identities, strictly
// self-isolated by the caller-supplied session user ID.
func (s *Service) List(ctx context.Context, userID int) ([]extidspi.BoundIdentity, error) {
	if userID <= 0 {
		return []extidspi.BoundIdentity{}, nil
	}
	var rows []*entity.UserExternalIdentity
	if err := dao.UserExternalIdentity.Ctx(ctx).
		Where(do.UserExternalIdentity{UserId: userID}).
		OrderAsc(dao.UserExternalIdentity.Columns().Provider).
		OrderAsc(dao.UserExternalIdentity.Columns().Subject).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeIdentityQueryFailed)
	}
	identities := make([]extidspi.BoundIdentity, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		identities = append(identities, extidspi.BoundIdentity{
			Provider:    row.Provider,
			Subject:     row.Subject,
			SubjectKind: extidspi.SubjectKind(row.SubjectKind),
			AppContext:  row.AppContext,
			Email:       row.EmailSnapshot,
			Phone:       row.PhoneSnapshot,
			DisplayName: row.DisplayNameSnapshot,
			AvatarURL:   row.AvatarUrlSnapshot,
		})
	}
	return identities, nil
}

func subjectKindOrDefault(kind extidspi.SubjectKind) string {
	if strings.TrimSpace(string(kind)) == "" {
		return string(extidspi.SubjectKindOIDCSub)
	}
	return string(kind)
}

func linkageDetailFromEntity(row *entity.UserExternalIdentity) *extidcap.LinkageDetail {
	if row == nil {
		return nil
	}
	return &extidcap.LinkageDetail{
		Provider:    row.Provider,
		Subject:     row.Subject,
		SubjectKind: extidspi.SubjectKind(row.SubjectKind),
		AppContext:  row.AppContext,
		UserID:      row.UserId,
		PluginID:    row.PluginId,
		Email:       row.EmailSnapshot,
		Phone:       row.PhoneSnapshot,
		DisplayName: row.DisplayNameSnapshot,
		AvatarURL:   row.AvatarUrlSnapshot,
	}
}

// findLinkage returns the live linkage row for one (provider, subject) key, or
// nil when none exists. Soft-deleted rows are filtered automatically.
func (s *Service) findLinkage(ctx context.Context, provider string, subject string) (*entity.UserExternalIdentity, error) {
	var linkage *entity.UserExternalIdentity
	if err := dao.UserExternalIdentity.Ctx(ctx).Where(do.UserExternalIdentity{
		Provider: provider,
		Subject:  subject,
	}).Scan(&linkage); err != nil {
		return nil, bizerr.WrapCode(err, CodeIdentityQueryFailed)
	}
	return linkage, nil
}

// provisionAccount creates one least-privilege host account through
// usercap.CreateFromExternal and parses the returned domain ID. Same-email
// conflict policy: the login path carries no actor context, so the unfiltered
// email-existence check is enforced by the host create primitive itself; this
// provider maps the capability sentinel into its caller-visible conflict code,
// rejecting silent auto-create so the user signs in to the existing account
// and binds the identity explicitly instead of enabling an email-assertion
// takeover.
func (s *Service) provisionAccount(ctx context.Context, provider string, email string, anchor string, displayName string) (int, error) {
	provisionedID, err := s.users.CreateFromExternal(ctx, usercap.CreateFromExternalInput{
		Email:          email,
		DisplayName:    strings.TrimSpace(displayName),
		Remark:         "auto-provisioned by external login provider " + provider,
		UsernameAnchor: anchor,
	})
	if err != nil {
		if errors.Is(err, usercap.ErrCreateFromExternalEmailConflict) {
			return 0, bizerr.NewCode(CodeProvisionEmailConflict)
		}
		return 0, bizerr.WrapCode(err, CodeProvisionFailed)
	}
	userID, err := strconv.Atoi(string(provisionedID))
	if err != nil || userID <= 0 {
		return 0, bizerr.NewCode(CodeProvisionFailed)
	}
	return userID, nil
}

// deriveUsernameAnchor derives the deterministic, collision-resistant username
// anchor for one email-less external identity. The digest covers provider and
// subject with a separator so distinct identities cannot produce colliding
// anchors through concatenation ambiguity.
func deriveUsernameAnchor(provider string, subject string) string {
	digest := sha256.Sum256([]byte(provider + "\x00" + subject))
	return usernameAnchorPrefix + hex.EncodeToString(digest[:])[:usernameAnchorHashLength]
}

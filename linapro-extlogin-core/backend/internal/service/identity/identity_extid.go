// identity_extid.go implements the plugin-owned extidcap sub surfaces:
// TicketService, LoginService, LinkageService, and ProviderService.

package identity

import (
	"context"
	"errors"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/authcap/extlogin/extidspi"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-plugin-linapro-extlogin-core/backend/cap/extidcap"
	"lina-plugin-linapro-extlogin-core/backend/internal/dao"
	"lina-plugin-linapro-extlogin-core/backend/internal/model/do"
	"lina-plugin-linapro-extlogin-core/backend/internal/model/entity"
)

// IssueVerifiedTicket stores one protocol-verified identity for later login or bind.
func (s *Service) IssueVerifiedTicket(ctx context.Context, identity extidcap.VerifiedIdentity) (*extidcap.TicketIssueResult, error) {
	if s == nil || s.tickets == nil {
		return nil, bizerr.NewCode(CodeTicketUnavailable)
	}
	return s.tickets.Issue(ctx, identity)
}

// PeekVerifiedTicket returns a ticket without consuming it.
func (s *Service) PeekVerifiedTicket(ctx context.Context, ticketID string) (*extidcap.VerifiedIdentity, error) {
	if s == nil || s.tickets == nil {
		return nil, bizerr.NewCode(CodeTicketUnavailable)
	}
	return s.tickets.Peek(ctx, ticketID)
}

// ConsumeVerifiedTicket returns and invalidates a ticket.
func (s *Service) ConsumeVerifiedTicket(ctx context.Context, ticketID string) (*extidcap.VerifiedIdentity, error) {
	if s == nil || s.tickets == nil {
		return nil, bizerr.NewCode(CodeTicketUnavailable)
	}
	return s.tickets.Consume(ctx, ticketID)
}

// InvalidateVerifiedTicket drops a ticket if present.
func (s *Service) InvalidateVerifiedTicket(ctx context.Context, ticketID string) error {
	if s == nil || s.tickets == nil {
		return bizerr.NewCode(CodeTicketUnavailable)
	}
	return s.tickets.Invalidate(ctx, ticketID)
}

// LoginPrepare resolves or provisions from a ticket or in-process identity.
func (s *Service) LoginPrepare(ctx context.Context, in extidcap.LoginPrepareInput) (*extidcap.LoginPrepareResult, error) {
	identity, err := s.resolvePrepareIdentity(ctx, in)
	if err != nil {
		return nil, err
	}
	if identity == nil {
		return &extidcap.LoginPrepareResult{Outcome: extidcap.LoginOutcomeUnavailable}, nil
	}
	userID, found, err := s.Resolve(ctx, extidspi.ResolveInput{
		Provider: identity.Provider,
		Subject:  identity.Subject,
	})
	if err != nil {
		return nil, err
	}
	if found {
		linkage, _ := s.GetLinkage(ctx, identity.Provider, identity.Subject)
		return &extidcap.LoginPrepareResult{
			Outcome: extidcap.LoginOutcomeLinked,
			UserID:  userID,
			Linkage: linkage,
		}, nil
	}
	if !in.AllowAutoProvision {
		return &extidcap.LoginPrepareResult{Outcome: extidcap.LoginOutcomeNeedsBind}, nil
	}
	if in.DryRun {
		return &extidcap.LoginPrepareResult{Outcome: extidcap.LoginOutcomeProvisioned}, nil
	}
	userID, err = s.Provision(ctx, extidspi.ProvisionInput{
		Provider:           identity.Provider,
		Subject:            identity.Subject,
		SubjectKind:        identity.SubjectKind,
		SecondarySubjects:  identity.SecondarySubjects,
		AppContext:         identity.AppContext,
		Email:              identity.Email,
		Phone:              identity.Phone,
		DisplayName:        identity.DisplayName,
		AvatarURL:          identity.AvatarURL,
		PluginID:           identity.PluginID,
		AllowAutoProvision: true,
	})
	if err != nil {
		if errors.Is(err, usercap.ErrCreateFromExternalEmailConflict) || bizerr.Is(err, CodeProvisionEmailConflict) {
			return &extidcap.LoginPrepareResult{Outcome: extidcap.LoginOutcomeConflictEmail}, nil
		}
		if bizerr.Is(err, CodeBindConflict) {
			return &extidcap.LoginPrepareResult{Outcome: extidcap.LoginOutcomeConflictSubject}, nil
		}
		return nil, err
	}
	linkage, _ := s.GetLinkage(ctx, identity.Provider, identity.Subject)
	return &extidcap.LoginPrepareResult{
		Outcome: extidcap.LoginOutcomeProvisioned,
		UserID:  userID,
		Linkage: linkage,
	}, nil
}

// PreviewBind reports whether a ticket can bind to the given user.
func (s *Service) PreviewBind(ctx context.Context, userID int, ticketID string) (*extidcap.BindPreviewResult, error) {
	if userID <= 0 {
		return &extidcap.BindPreviewResult{OK: false, Reason: extidcap.BindPreviewInvalidUser}, nil
	}
	identity, err := s.PeekVerifiedTicket(ctx, ticketID)
	if err != nil {
		return &extidcap.BindPreviewResult{OK: false, Reason: extidcap.BindPreviewTicketInvalid}, nil
	}
	linkage, err := s.findLinkage(ctx, identity.Provider, identity.Subject)
	if err != nil {
		return nil, err
	}
	if linkage != nil && linkage.UserId != userID {
		return &extidcap.BindPreviewResult{OK: false, Reason: extidcap.BindPreviewConflict}, nil
	}
	return &extidcap.BindPreviewResult{
		OK:      true,
		Linkage: linkageDetailFromEntity(linkage),
	}, nil
}

// BindByTicket consumes a verified ticket and links it to the session user.
func (s *Service) BindByTicket(ctx context.Context, userID int, ticketID string) error {
	identity, err := s.ConsumeVerifiedTicket(ctx, ticketID)
	if err != nil {
		return err
	}
	return s.Bind(ctx, extidspi.BindInput{
		UserID:            userID,
		Provider:          identity.Provider,
		Subject:           identity.Subject,
		SubjectKind:       identity.SubjectKind,
		SecondarySubjects: identity.SecondarySubjects,
		AppContext:        identity.AppContext,
		Email:             identity.Email,
		Phone:             identity.Phone,
		DisplayName:       identity.DisplayName,
		AvatarURL:         identity.AvatarURL,
		PluginID:          identity.PluginID,
	})
}

// GetLinkage returns one linkage detail by authoritative key.
func (s *Service) GetLinkage(ctx context.Context, provider, subject string) (*extidcap.LinkageDetail, error) {
	row, err := s.findLinkage(ctx, strings.TrimSpace(provider), strings.TrimSpace(subject))
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	return linkageDetailFromEntity(row), nil
}

// ListByUser returns all linkages for one user as domain details.
func (s *Service) ListByUser(ctx context.Context, userID int) ([]extidcap.LinkageDetail, error) {
	bound, err := s.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]extidcap.LinkageDetail, 0, len(bound))
	for _, item := range bound {
		out = append(out, extidcap.LinkageDetail{
			Provider:    item.Provider,
			Subject:     item.Subject,
			SubjectKind: item.SubjectKind,
			AppContext:  item.AppContext,
			UserID:      userID,
			Email:       item.Email,
			Phone:       item.Phone,
			DisplayName: item.DisplayName,
			AvatarURL:   item.AvatarURL,
		})
	}
	return out, nil
}

// ListByProvider returns a page of linkages for one provider (operator projection).
func (s *Service) ListByProvider(ctx context.Context, provider string, page, pageSize int) ([]extidcap.LinkageDetail, int, error) {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return nil, 0, bizerr.NewCode(CodeIdentityInvalid)
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	model := dao.UserExternalIdentity.Ctx(ctx).Where(do.UserExternalIdentity{Provider: provider})
	total, err := model.Count()
	if err != nil {
		return nil, 0, bizerr.WrapCode(err, CodeIdentityQueryFailed)
	}
	var rows []*entity.UserExternalIdentity
	if err = model.Page(page, pageSize).
		OrderAsc(dao.UserExternalIdentity.Columns().Subject).
		Scan(&rows); err != nil {
		return nil, 0, bizerr.WrapCode(err, CodeIdentityQueryFailed)
	}
	out := make([]extidcap.LinkageDetail, 0, len(rows))
	for _, row := range rows {
		if detail := linkageDetailFromEntity(row); detail != nil {
			out = append(out, *detail)
		}
	}
	return out, total, nil
}

// SyncProfile refreshes snapshots from a ticket without changing the subject key.
func (s *Service) SyncProfile(ctx context.Context, ticketID string) error {
	identity, err := s.ConsumeVerifiedTicket(ctx, ticketID)
	if err != nil {
		return err
	}
	result, err := dao.UserExternalIdentity.Ctx(ctx).Where(do.UserExternalIdentity{
		Provider: identity.Provider,
		Subject:  identity.Subject,
	}).Data(do.UserExternalIdentity{
		EmailSnapshot:       strings.TrimSpace(identity.Email),
		PhoneSnapshot:       strings.TrimSpace(identity.Phone),
		DisplayNameSnapshot: strings.TrimSpace(identity.DisplayName),
		AvatarUrlSnapshot:   strings.TrimSpace(identity.AvatarURL),
		AppContext:          strings.TrimSpace(identity.AppContext),
		SubjectKind:         subjectKindOrDefault(identity.SubjectKind),
	}).Update()
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

// ListProviders returns the process-local protocol catalog.
func (s *Service) ListProviders(ctx context.Context, filter extidcap.ProviderFilter) ([]extidcap.ProviderView, error) {
	return s.listProvidersFromCap(ctx, filter), nil
}

// GetProvider returns one catalog entry.
func (s *Service) GetProvider(ctx context.Context, providerID string) (*extidcap.ProviderView, error) {
	view, ok := s.getProviderFromCap(ctx, providerID)
	if !ok {
		return nil, bizerr.NewCode(CodeProviderNotFound)
	}
	return view, nil
}

// GetProvisionPolicy returns a conservative default policy view.
func (s *Service) GetProvisionPolicy(_ context.Context, providerID string) (*extidcap.ProvisionPolicy, error) {
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		return nil, bizerr.NewCode(CodeIdentityInvalid)
	}
	return &extidcap.ProvisionPolicy{
		Provider:               providerID,
		AllowAutoProvision:     false,
		RequireEmail:           false,
		RequirePhone:           false,
		OnEmailConflict:        extidcap.ConflictPolicyRequireBind,
		OnSubjectConflict:      extidcap.ConflictPolicyReject,
		UsernameAnchorStrategy: extidcap.UsernameAnchorSHA256ProviderSubject,
		SubjectStrategy:        extidcap.SubjectStrategyOIDCSub,
	}, nil
}

func (s *Service) resolvePrepareIdentity(ctx context.Context, in extidcap.LoginPrepareInput) (*extidcap.VerifiedIdentity, error) {
	if strings.TrimSpace(in.TicketID) != "" {
		if in.DryRun {
			return s.PeekVerifiedTicket(ctx, in.TicketID)
		}
		return s.ConsumeVerifiedTicket(ctx, in.TicketID)
	}
	if in.Identity == nil {
		return nil, bizerr.NewCode(CodeIdentityInvalid)
	}
	identity := *in.Identity
	identity.Provider = strings.TrimSpace(identity.Provider)
	identity.Subject = strings.TrimSpace(identity.Subject)
	if identity.Provider == "" || identity.Subject == "" {
		return nil, bizerr.NewCode(CodeIdentityInvalid)
	}
	return &identity, nil
}

// resolver_impl.go implements request tenant resolution for the multi-tenant
// plugin. It evaluates configured header, query, subdomain, domain, path, and
// default strategies in priority order while returning explicit resolution
// failures for callers that must fail closed.

package resolver

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// Register registers or replaces one resolver implementation.
func (s *serviceImpl) Register(resolver Resolver) {
	if resolver == nil {
		return
	}
	s.resolvers[resolver.Name()] = resolver
}

// Resolve runs the configured resolver chain.
func (s *serviceImpl) Resolve(ctx context.Context, r *ghttp.Request, config Config) (*Result, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	identity := Identity{
		UserID:          int64(bizCtx.UserID),
		TenantID:        int64(bizCtx.TenantID),
		ActingUserID:    int64(bizCtx.ActingUserID),
		ActingAsTenant:  bizCtx.ActingAsTenant,
		IsImpersonation: bizCtx.IsImpersonation,
		IsPlatform:      int64(bizCtx.TenantID) == shared.PlatformTenantID,
	}
	chain := config.Chain
	if len(chain) == 0 {
		chain = shared.DefaultResolverChain()
	}
	for _, name := range chain {
		resolverImpl := s.resolvers[strings.TrimSpace(name)]
		if resolverImpl == nil {
			continue
		}
		result, ok, err := resolverImpl.Resolve(ctx, r, identity, config)
		if err != nil || !ok {
			if err != nil {
				return nil, err
			}
			continue
		}
		if err = s.validateMembership(ctx, identity, result); err != nil {
			return nil, err
		}
		return result, nil
	}
	switch config.OnAmbiguous {
	case shared.OnAmbiguousReject:
		return nil, bizerr.NewCode(CodeTenantForbidden)
	default:
		return nil, bizerr.NewCode(CodeTenantRequired)
	}
}

// validateMembership verifies non-platform users belong to the resolved tenant.
func (s *serviceImpl) validateMembership(ctx context.Context, identity Identity, result *Result) error {
	if result == nil || result.TenantID == shared.PlatformTenantID {
		return nil
	}
	if result.ActingAsTenant || identity.ActingAsTenant || identity.IsImpersonation {
		return nil
	}
	if identity.UserID == 0 {
		return nil
	}
	_, err := s.membershipSvc.GetByUserAndTenant(ctx, identity.UserID, result.TenantID)
	return err
}

// Package resolver implements the plugin-owned tenant resolution chain.
package resolver

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	plugincontract "lina-core/pkg/pluginservice/contract"
	"lina-plugin-multi-tenant/backend/internal/service/membership"
)

// Result is the output of one resolver.
type Result struct {
	TenantID       int64
	Source         string
	ActingAsTenant bool
}

// Resolver resolves a tenant from an HTTP request without mutating tenant data.
type Resolver interface {
	// Name returns the configured resolver name.
	Name() string
	// Resolve evaluates the request, authenticated identity, and resolver config.
	// It returns the resolved tenant, whether this resolver matched, and any
	// parsing or validation error.
	Resolve(ctx context.Context, r *ghttp.Request, identity Identity, config Config) (*Result, bool, error)
}

// Identity describes the authenticated user known to the resolver chain.
type Identity struct {
	UserID          int64 // UserID is the authenticated host user identifier.
	TenantID        int64 // TenantID is the tenant already attached by host JWT auth.
	ActingUserID    int64 // ActingUserID is the real platform user during impersonation.
	ActingAsTenant  bool  // ActingAsTenant reports whether the request is operating through a tenant view.
	IsImpersonation bool  // IsImpersonation reports whether the token is an impersonation token.
	IsPlatform      bool  // IsPlatform reports whether the caller is in platform context.
}

// Config defines resolver chain behavior.
type Config struct {
	Chain              []string
	ReservedSubdomains []string
	RootDomain         string
	OnAmbiguous        string
}

// Service defines tenant resolution operations for the plugin provider seam.
type Service interface {
	// Resolve runs the configured resolver chain for the current request, then
	// validates membership when required. It returns business/config errors without
	// changing cache, i18n, or persisted tenant state.
	Resolve(ctx context.Context, r *ghttp.Request, config Config) (*Result, error)
	// Register registers or replaces one resolver implementation by name for the
	// in-process chain used by this service instance.
	Register(resolver Resolver)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc     plugincontract.BizCtxService
	membershipSvc membership.Service
	resolvers     map[string]Resolver
}

// New creates and returns a resolver service with the built-in resolver set.
func New(bizCtxSvc plugincontract.BizCtxService, membershipSvc membership.Service) Service {
	s := &serviceImpl{
		bizCtxSvc:     bizCtxSvc,
		membershipSvc: membershipSvc,
		resolvers:     make(map[string]Resolver),
	}
	s.Register(overrideResolver{})
	s.Register(jwtResolver{})
	s.Register(sessionResolver{})
	s.Register(headerResolver{})
	s.Register(subdomainResolver{})
	s.Register(defaultResolver{membershipSvc: s.membershipSvc})
	return s
}

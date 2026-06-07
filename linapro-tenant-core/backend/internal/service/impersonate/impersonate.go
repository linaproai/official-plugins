// Package impersonate implements the plugin-side impersonation command shape.
package impersonate

import (
	"context"

	"lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/usercap"
	tenantsvc "lina-plugin-linapro-tenant-core/backend/internal/service/tenant"
)

// Service defines platform-to-tenant impersonation operations.
type Service interface {
	// Start validates a platform user's request to enter a target tenant, creates
	// a compatible host token/session, writes audit rows, and returns token metadata.
	// It returns business or persistence errors when authorization, tenant status,
	// config, token signing, session creation, or audit writes fail.
	Start(ctx context.Context, in StartInput) (*StartOutput, error)
	// Stop validates and revokes one current impersonation token for the supplied
	// tenant. It returns token parsing, tenant mismatch, or persistence errors.
	Stop(ctx context.Context, in StopInput) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	authSvc   token.Service
	authzSvc  authz.Service
	bizCtxSvc bizctxcap.Service
	tenantSvc tenantsvc.Service
	users     usercap.Service
}

// New creates and returns an impersonation service.
func New(
	authSvc token.Service,
	authzSvc authz.Service,
	bizCtxSvc bizctxcap.Service,
	tenantSvc tenantsvc.Service,
	users usercap.Service,
) Service {
	return &serviceImpl{
		authSvc:   authSvc,
		authzSvc:  authzSvc,
		bizCtxSvc: bizCtxSvc,
		tenantSvc: tenantSvc,
		users:     users,
	}
}

// StartInput defines impersonation start input.
type StartInput struct {
	TenantID int64
	Reason   string
}

// StartOutput defines impersonation start output.
type StartOutput struct {
	Token          string
	TenantID       int64
	ActingUserID   int64
	IsImpersonated bool
}

// StopInput defines impersonation stop input.
type StopInput struct {
	TenantID int64
	Token    string
}

// loginLogData is a typed insert payload for plugin_linapro_monitor_loginlog.
type loginLogData struct {
	TenantID           int64  `orm:"tenant_id"`
	ActingUserID       int64  `orm:"acting_user_id"`
	OnBehalfOfTenantID int64  `orm:"on_behalf_of_tenant_id"`
	IsImpersonation    bool   `orm:"is_impersonation"`
	UserName           string `orm:"user_name"`
	Status             int    `orm:"status"`
	IP                 string `orm:"ip"`
	Browser            string `orm:"browser"`
	OS                 string `orm:"os"`
	Msg                string `orm:"msg"`
}

// operLogData is a typed insert payload for plugin_linapro_monitor_operlog.
type operLogData struct {
	TenantID           int64  `orm:"tenant_id"`
	ActingUserID       int64  `orm:"acting_user_id"`
	OnBehalfOfTenantID int64  `orm:"on_behalf_of_tenant_id"`
	IsImpersonation    bool   `orm:"is_impersonation"`
	Title              string `orm:"title"`
	OperSummary        string `orm:"oper_summary"`
	RouteOwner         string `orm:"route_owner"`
	RouteMethod        string `orm:"route_method"`
	RoutePath          string `orm:"route_path"`
	RouteDocKey        string `orm:"route_doc_key"`
	OperType           string `orm:"oper_type"`
	Method             string `orm:"method"`
	RequestMethod      string `orm:"request_method"`
	OperName           string `orm:"oper_name"`
	OperURL            string `orm:"oper_url"`
	OperIP             string `orm:"oper_ip"`
	OperParam          string `orm:"oper_param"`
	JsonResult         string `orm:"json_result"`
	Status             int    `orm:"status"`
	ErrorMsg           string `orm:"error_msg"`
	CostTime           int    `orm:"cost_time"`
}

// auditInput defines impersonation audit fields.
type auditInput struct {
	TenantID     int64
	ActingUserID int64
	Username     string
	Reason       string
	Client       clientInfo
}

// clientInfo contains normalized request client metadata.
type clientInfo struct {
	IP      string
	Browser string
	OS      string
	URL     string
}

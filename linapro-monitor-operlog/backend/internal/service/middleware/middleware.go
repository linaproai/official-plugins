// Package middleware implements linapro-monitor-operlog HTTP audit middleware and
// request normalization services for the source plugin.
package middleware

import (
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/routecap"

	"github.com/gogf/gf/v2/net/ghttp"

	operlogsvc "lina-plugin-linapro-monitor-operlog/backend/internal/service/operlog"
)

// Service defines the linapro-monitor-operlog middleware service contract.
type Service interface {
	// Audit captures one completed request and persists the normalized operation log.
	Audit(request *ghttp.Request)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	routeMetaSvc routecap.Service   // dynamic-route metadata reader
	bizCtxSvc    bizctxcap.Service  // authenticated operator identity reader
	operLogSvc   operlogsvc.Service // plugin-owned operation-log persistence service
}

// New creates and returns a new linapro-monitor-operlog middleware service instance.
func New(
	routeMetaSvc routecap.Service,
	bizCtxSvc bizctxcap.Service,
	operLogSvc operlogsvc.Service,
) Service {
	return &serviceImpl{
		routeMetaSvc: routeMetaSvc,
		bizCtxSvc:    bizCtxSvc,
		operLogSvc:   operLogSvc,
	}
}

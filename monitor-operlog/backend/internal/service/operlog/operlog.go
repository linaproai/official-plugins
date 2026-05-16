// Package operlog implements operation-log persistence, query, cleanup,
// and export services for the monitor-operlog source plugin. It owns the
// plugin_monitor_operlog table access instead of depending on host-internal operlog
// services.
package operlog

import (
	"context"

	plugincontract "lina-core/pkg/pluginservice/contract"
	entitymodel "lina-plugin-monitor-operlog/backend/internal/model/entity"
	"lina-plugin-monitor-operlog/backend/internal/model/operlogtype"
)

// Table, column, and dictionary constants used by the plugin-owned operation-log service.
const (
	colID            = "id"
	colTitle         = "title"
	colOperSummary   = "oper_summary"
	colRouteOwner    = "route_owner"
	colRouteMethod   = "route_method"
	colRoutePath     = "route_path"
	colRouteDocKey   = "route_doc_key"
	colOperType      = "oper_type"
	colMethod        = "method"
	colRequestMethod = "request_method"
	colOperName      = "oper_name"
	colOperURL       = "oper_url"
	colOperIP        = "oper_ip"
	colOperParam     = "oper_param"
	colJSONResult    = "json_result"
	colStatus        = "status"
	colErrorMsg      = "error_msg"
	colCostTime      = "cost_time"
	colOperTime      = "oper_time"

	colDictType  = "dict_type"
	colDictValue = "value"
	colDictLabel = "label"
	colDictSort  = "sort"
)

// Operation-log runtime i18n key fragments.
const (
	dictKeyPrefix  = "dict"
	labelKeySuffix = "label"
)

// Operation-log export limit and dictionary constants.
const (
	MaxExportRows      = 10000
	DictTypeOperType   = "sys_oper_type"
	DictTypeOperStatus = "sys_oper_status"
)

// Operation status values stored in plugin_monitor_operlog.
const (
	OperStatusSuccess = 0
	OperStatusFail    = 1
)

// Service defines tenant-scoped operation-log persistence, query, cleanup, and export.
type Service interface {
	// Create inserts one operation-log record using explicit audit tenant
	// overrides when provided, otherwise ctx's tenant metadata.
	Create(ctx context.Context, in CreateInput) error
	// List returns ctx-visible operation logs with pagination, filters, ordering,
	// and runtime i18n localization for route titles and dictionary-backed values.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById returns one tenant-visible operation log by primary key, localized
	// for ctx's locale, or nil when the row is outside scope or absent.
	GetById(ctx context.Context, id int) (*OperLogEntity, error)
	// Clean hard-deletes tenant-visible operation logs within the optional time range.
	// It returns the affected row count and any database error.
	Clean(ctx context.Context, in CleanInput) (int, error)
	// DeleteByIds hard-deletes tenant-visible operation logs by ID list and
	// returns the affected row count; rows outside data scope are ignored.
	DeleteByIds(ctx context.Context, ids []int) (int, error)
	// Export generates an Excel workbook for tenant-visible operation logs using
	// apidoc/runtime i18n and dictionary fallbacks. It caps output at MaxExportRows.
	Export(ctx context.Context, in ExportInput) (data []byte, err error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	apiDocSvc    plugincontract.APIDocService       // host apidoc translation service
	i18nSvc      plugincontract.I18nService         // host runtime translation service
	tenantFilter plugincontract.TenantFilterService // tenantFilter constrains plugin-owned operation-log rows.
}

// New creates and returns a new monitor-operlog service instance.
func New(
	apiDocSvc plugincontract.APIDocService,
	i18nSvc plugincontract.I18nService,
	tenantFilter plugincontract.TenantFilterService,
) Service {
	return &serviceImpl{
		apiDocSvc:    apiDocSvc,
		i18nSvc:      i18nSvc,
		tenantFilter: tenantFilter,
	}
}

// OperLogEntity mirrors the plugin-local generated plugin_monitor_operlog entity.
type OperLogEntity = entitymodel.Operlog

// dictDataRow reuses the plugin-local generated sys_dict_data entity.
type dictDataRow = entitymodel.SysDictData

// CreateInput defines the operation-log create input.
type CreateInput struct {
	TenantID           *int
	ActingUserID       *int
	OnBehalfOfTenantID *int
	IsImpersonation    *bool
	Title              string
	OperSummary        string
	RouteOwner         string
	RouteMethod        string
	RoutePath          string
	RouteDocKey        string
	OperType           operlogtype.OperType
	Method             string
	RequestMethod      string
	OperName           string
	OperUrl            string
	OperIp             string
	OperParam          string
	JsonResult         string
	Status             int
	ErrorMsg           string
	CostTime           int
}

// ListInput defines the operation-log list filter input.
type ListInput struct {
	PageNum        int
	PageSize       int
	Title          string
	OperName       string
	OperType       *operlogtype.OperType
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
}

// ListOutput defines the operation-log list output.
type ListOutput struct {
	List  []*OperLogEntity
	Total int
}

// CleanInput defines the operation-log cleanup input.
type CleanInput struct {
	BeginTime string
	EndTime   string
}

// ExportInput defines the operation-log export input.
type ExportInput struct {
	Title          string
	OperName       string
	OperType       *operlogtype.OperType
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
	Ids            []int
}

// exportHeader describes one localized Excel header cell.
type exportHeader struct {
	Key      string // Key is the runtime i18n key for the header.
	Fallback string // Fallback is used when the runtime bundle has no translation.
}

// auditTenantContext stores tenant metadata persisted with one operation log.
type auditTenantContext struct {
	TenantID           int  // TenantID owns the log row.
	ActingUserID       int  // ActingUserID is the platform actor during impersonation.
	OnBehalfOfTenantID int  // OnBehalfOfTenantID is the operated tenant.
	IsImpersonation    bool // IsImpersonation marks platform impersonation.
}

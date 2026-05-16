// Package loginlog implements login-log persistence, query, cleanup, and
// export services for the monitor-loginlog source plugin. It owns the
// plugin_monitor_loginlog table access instead of depending on host-internal loginlog
// services.
package loginlog

import (
	"context"

	plugincontract "lina-core/pkg/pluginservice/contract"
	entitymodel "lina-plugin-monitor-loginlog/backend/internal/model/entity"
)

// Table, column, and dictionary constants used by the plugin-owned login-log service.
const (
	colID        = "id"
	colUserName  = "user_name"
	colStatus    = "status"
	colIP        = "ip"
	colBrowser   = "browser"
	colOS        = "os"
	colMsg       = "msg"
	colLoginTime = "login_time"

	colDictType  = "dict_type"
	colDictValue = "value"
	colDictLabel = "label"
	colDictSort  = "sort"
)

// Login-log export and dictionary constants.
const (
	MaxExportRows       = 10000
	DictTypeLoginStatus = "sys_login_status"
)

// Runtime i18n key fragments used by dictionary display projection.
const (
	// dictKeyPrefix is the runtime i18n root for dictionary labels.
	dictKeyPrefix = "dict"
	// labelKeySuffix is the final i18n segment for dictionary display labels.
	labelKeySuffix = "label"
	// loginLogMessagePrefix is the plugin-owned runtime i18n root for auth messages.
	loginLogMessagePrefix = "plugin.monitor-loginlog.logMessage"
)

// Login status values stored in plugin_monitor_loginlog.
const (
	LoginStatusSuccess = 0
	LoginStatusFail    = 1
)

// Service defines tenant-scoped login-log persistence, query, cleanup, and export.
type Service interface {
	// Create inserts one login-log record using explicit audit tenant overrides
	// when provided, otherwise the tenant context from ctx. It returns database errors.
	Create(ctx context.Context, in CreateInput) error
	// List returns ctx-visible login logs with pagination, filter, ordering, and
	// runtime i18n localization applied to display fields.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById returns one tenant-visible login log by primary key, localized for
	// ctx's locale, or nil when the row is outside scope or absent.
	GetById(ctx context.Context, id int) (*LoginLogEntity, error)
	// Clean hard-deletes tenant-visible login logs within the optional time range.
	// It returns the affected row count and any database error.
	Clean(ctx context.Context, in CleanInput) (int, error)
	// DeleteByIds hard-deletes tenant-visible login logs by ID list and returns
	// the affected row count; rows outside data scope are ignored by the filter.
	DeleteByIds(ctx context.Context, ids []int) (int, error)
	// Export generates an Excel workbook for tenant-visible login logs using
	// runtime i18n and dictionary fallbacks. It caps output at MaxExportRows.
	Export(ctx context.Context, in ExportInput) (data []byte, err error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	i18nSvc      plugincontract.I18nService         // i18nSvc resolves host runtime translations for plugin data.
	tenantFilter plugincontract.TenantFilterService // tenantFilter constrains plugin-owned login-log rows.
}

// New creates and returns a new monitor-loginlog service instance.
func New(i18nSvc plugincontract.I18nService, tenantFilter plugincontract.TenantFilterService) Service {
	return &serviceImpl{
		i18nSvc:      i18nSvc,
		tenantFilter: tenantFilter,
	}
}

// LoginLogEntity mirrors the plugin-local generated plugin_monitor_loginlog entity.
type LoginLogEntity = entitymodel.Loginlog

// dictDataRow reuses the plugin-local generated sys_dict_data entity.
type dictDataRow = entitymodel.SysDictData

// CreateInput defines the login-log create input.
type CreateInput struct {
	TenantID           *int
	ActingUserID       *int
	OnBehalfOfTenantID *int
	IsImpersonation    *bool
	UserName           string
	Status             int
	Ip                 string
	Browser            string
	Os                 string
	Msg                string
}

// ListInput defines the login-log list filter input.
type ListInput struct {
	PageNum        int
	PageSize       int
	UserName       string
	Ip             string
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
}

// ListOutput defines the login-log list output.
type ListOutput struct {
	List  []*LoginLogEntity
	Total int
}

// CleanInput defines the login-log cleanup input.
type CleanInput struct {
	BeginTime string
	EndTime   string
}

// ExportInput defines the login-log export input.
type ExportInput struct {
	UserName       string
	Ip             string
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
	Ids            []int
}

// auditTenantContext stores tenant metadata persisted with one login log.
type auditTenantContext struct {
	TenantID           int  // TenantID owns the log row.
	ActingUserID       int  // ActingUserID is the platform actor during impersonation.
	OnBehalfOfTenantID int  // OnBehalfOfTenantID is the operated tenant.
	IsImpersonation    bool // IsImpersonation marks platform impersonation.
}

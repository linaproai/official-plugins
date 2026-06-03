// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProviderOperationDao is the data access object for the table plugin_linapro_ai_provider_operation.
type ProviderOperationDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  ProviderOperationColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// ProviderOperationColumns defines and stores column names for the table plugin_linapro_ai_provider_operation.
type ProviderOperationColumns struct {
	Id               string // Provider operation row ID
	OperationRef     string // Opaque provider operation reference
	CapabilityType   string // Capability type
	CapabilityMethod string // Capability method
	Purpose          string // Governed AI purpose
	SourcePluginId   string // Source plugin ID
	ProviderId       string // Provider ID
	ModelId          string // Model ID
	ProviderName     string // Provider display name snapshot
	ModelName        string // Model name snapshot
	Protocol         string // Protocol snapshot
	Status           string // Provider operation status
	NextPollAfterMs  string // Recommended next poll delay in milliseconds
	ExpiresAt        string // Operation reference expiration time
	AssetSummaryJson string // Asset reference summary JSON without file contents
	ErrorCode        string // Stable error code
	ErrorSummary     string // Masked error summary
	CreatedAt        string // Creation time
	UpdatedAt        string // Update time
	DeletedAt        string // Deletion time
}

// providerOperationColumns holds the columns for the table plugin_linapro_ai_provider_operation.
var providerOperationColumns = ProviderOperationColumns{
	Id:               "id",
	OperationRef:     "operation_ref",
	CapabilityType:   "capability_type",
	CapabilityMethod: "capability_method",
	Purpose:          "purpose",
	SourcePluginId:   "source_plugin_id",
	ProviderId:       "provider_id",
	ModelId:          "model_id",
	ProviderName:     "provider_name",
	ModelName:        "model_name",
	Protocol:         "protocol",
	Status:           "status",
	NextPollAfterMs:  "next_poll_after_ms",
	ExpiresAt:        "expires_at",
	AssetSummaryJson: "asset_summary_json",
	ErrorCode:        "error_code",
	ErrorSummary:     "error_summary",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
	DeletedAt:        "deleted_at",
}

// NewProviderOperationDao creates and returns a new DAO object for table data access.
func NewProviderOperationDao(handlers ...gdb.ModelHandler) *ProviderOperationDao {
	return &ProviderOperationDao{
		group:    "default",
		table:    "plugin_linapro_ai_provider_operation",
		columns:  providerOperationColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProviderOperationDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProviderOperationDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProviderOperationDao) Columns() ProviderOperationColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProviderOperationDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProviderOperationDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *ProviderOperationDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

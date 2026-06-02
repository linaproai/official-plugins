// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// InvocationDao is the data access object for the table plugin_linapro_ai_invocation.
type InvocationDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  InvocationColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// InvocationColumns defines and stores column names for the table plugin_linapro_ai_invocation.
type InvocationColumns struct {
	Id             string // Invocation ID
	RequestId      string // Request correlation ID
	CapabilityType string // Capability type
	Purpose        string // Governed AI purpose
	TierCode       string // Tier code
	SourcePluginId string // Source plugin ID
	TenantId       string // Tenant ID
	UserId         string // User ID
	ProviderId     string // Provider ID
	ModelId        string // Model ID
	ProviderName   string // Provider display name snapshot
	ModelName      string // Model name snapshot
	Protocol       string // Protocol snapshot
	ThinkingEffort string // Requested or applied thinking effort
	Status         string // Invocation status: success or failed
	InputTokens    string // Input token count
	OutputTokens   string // Output token count
	LatencyMs      string // Provider call latency in milliseconds
	ErrorCode      string // Stable error code
	ErrorSummary   string // Masked error summary
	CreatedAt      string // Creation time
}

// invocationColumns holds the columns for the table plugin_linapro_ai_invocation.
var invocationColumns = InvocationColumns{
	Id:             "id",
	RequestId:      "request_id",
	CapabilityType: "capability_type",
	Purpose:        "purpose",
	TierCode:       "tier_code",
	SourcePluginId: "source_plugin_id",
	TenantId:       "tenant_id",
	UserId:         "user_id",
	ProviderId:     "provider_id",
	ModelId:        "model_id",
	ProviderName:   "provider_name",
	ModelName:      "model_name",
	Protocol:       "protocol",
	ThinkingEffort: "thinking_effort",
	Status:         "status",
	InputTokens:    "input_tokens",
	OutputTokens:   "output_tokens",
	LatencyMs:      "latency_ms",
	ErrorCode:      "error_code",
	ErrorSummary:   "error_summary",
	CreatedAt:      "created_at",
}

// NewInvocationDao creates and returns a new DAO object for table data access.
func NewInvocationDao(handlers ...gdb.ModelHandler) *InvocationDao {
	return &InvocationDao{
		group:    "default",
		table:    "plugin_linapro_ai_invocation",
		columns:  invocationColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *InvocationDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *InvocationDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *InvocationDao) Columns() InvocationColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *InvocationDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *InvocationDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *InvocationDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

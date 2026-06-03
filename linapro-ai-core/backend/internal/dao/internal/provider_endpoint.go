// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProviderEndpointDao is the data access object for the table plugin_linapro_ai_provider_endpoint.
type ProviderEndpointDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  ProviderEndpointColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// ProviderEndpointColumns defines and stores column names for the table plugin_linapro_ai_provider_endpoint.
type ProviderEndpointColumns struct {
	Id           string // Provider endpoint ID
	ProviderId   string // Provider ID
	Protocol     string // Protocol: openai, anthropic, voyage, openai-compatible, or provider-specific
	BaseUrl      string // Provider protocol base URL
	SecretRef    string // Secret reference or masked secret reference
	Enabled      string // Enabled flag: 0=disabled 1=enabled
	MetadataJson string // Endpoint metadata JSON without secret values
	CreatedAt    string // Creation time
	UpdatedAt    string // Update time
	DeletedAt    string // Deletion time
}

// providerEndpointColumns holds the columns for the table plugin_linapro_ai_provider_endpoint.
var providerEndpointColumns = ProviderEndpointColumns{
	Id:           "id",
	ProviderId:   "provider_id",
	Protocol:     "protocol",
	BaseUrl:      "base_url",
	SecretRef:    "secret_ref",
	Enabled:      "enabled",
	MetadataJson: "metadata_json",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewProviderEndpointDao creates and returns a new DAO object for table data access.
func NewProviderEndpointDao(handlers ...gdb.ModelHandler) *ProviderEndpointDao {
	return &ProviderEndpointDao{
		group:    "default",
		table:    "plugin_linapro_ai_provider_endpoint",
		columns:  providerEndpointColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProviderEndpointDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProviderEndpointDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProviderEndpointDao) Columns() ProviderEndpointColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProviderEndpointDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProviderEndpointDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProviderEndpointDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

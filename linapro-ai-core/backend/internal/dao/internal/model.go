// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ModelDao is the data access object for the table plugin_linapro_ai_model.
type ModelDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ModelColumns       // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ModelColumns defines and stores column names for the table plugin_linapro_ai_model.
type ModelColumns struct {
	Id         string // Model ID
	ProviderId string // Provider ID
	ModelName  string // Provider model name
	Protocol   string // Protocol: openai or anthropic
	Source     string // Model source: manual or api
	Enabled    string // Enabled flag: 0=disabled 1=enabled
	CreatedAt  string // Creation time
	UpdatedAt  string // Update time
	DeletedAt  string // Deletion time
	EndpointId string // Provider endpoint ID used by the model
}

// modelColumns holds the columns for the table plugin_linapro_ai_model.
var modelColumns = ModelColumns{
	Id:         "id",
	ProviderId: "provider_id",
	ModelName:  "model_name",
	Protocol:   "protocol",
	Source:     "source",
	Enabled:    "enabled",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
	DeletedAt:  "deleted_at",
	EndpointId: "endpoint_id",
}

// NewModelDao creates and returns a new DAO object for table data access.
func NewModelDao(handlers ...gdb.ModelHandler) *ModelDao {
	return &ModelDao{
		group:    "default",
		table:    "plugin_linapro_ai_model",
		columns:  modelColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ModelDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ModelDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ModelDao) Columns() ModelColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ModelDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ModelDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ModelDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

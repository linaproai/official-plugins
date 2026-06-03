// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MethodDefaultParamDao is the data access object for the table plugin_linapro_ai_method_default_param.
type MethodDefaultParamDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  MethodDefaultParamColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// MethodDefaultParamColumns defines and stores column names for the table plugin_linapro_ai_method_default_param.
type MethodDefaultParamColumns struct {
	Id                string // Method default param ID
	CapabilityType    string // Capability type
	CapabilityMethod  string // Capability method
	DefaultParamsJson string // Method-specific default params JSON
	Enabled           string // Enabled flag: 0=disabled 1=enabled
	CreatedAt         string // Creation time
	UpdatedAt         string // Update time
}

// methodDefaultParamColumns holds the columns for the table plugin_linapro_ai_method_default_param.
var methodDefaultParamColumns = MethodDefaultParamColumns{
	Id:                "id",
	CapabilityType:    "capability_type",
	CapabilityMethod:  "capability_method",
	DefaultParamsJson: "default_params_json",
	Enabled:           "enabled",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewMethodDefaultParamDao creates and returns a new DAO object for table data access.
func NewMethodDefaultParamDao(handlers ...gdb.ModelHandler) *MethodDefaultParamDao {
	return &MethodDefaultParamDao{
		group:    "default",
		table:    "plugin_linapro_ai_method_default_param",
		columns:  methodDefaultParamColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MethodDefaultParamDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MethodDefaultParamDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MethodDefaultParamDao) Columns() MethodDefaultParamColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MethodDefaultParamDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MethodDefaultParamDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MethodDefaultParamDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

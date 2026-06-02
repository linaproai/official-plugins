// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TierBindingDao is the data access object for the table plugin_linapro_ai_tier_binding.
type TierBindingDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  TierBindingColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// TierBindingColumns defines and stores column names for the table plugin_linapro_ai_tier_binding.
type TierBindingColumns struct {
	Id         string // Binding ID
	TierId     string // Tier ID
	ProviderId string // Provider ID
	ModelId    string // Model ID
	Priority   string // Binding priority, 0 is primary
	Enabled    string // Enabled flag: 0=disabled 1=enabled
	CreatedAt  string // Creation time
	UpdatedAt  string // Update time
	DeletedAt  string // Deletion time
}

// tierBindingColumns holds the columns for the table plugin_linapro_ai_tier_binding.
var tierBindingColumns = TierBindingColumns{
	Id:         "id",
	TierId:     "tier_id",
	ProviderId: "provider_id",
	ModelId:    "model_id",
	Priority:   "priority",
	Enabled:    "enabled",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
	DeletedAt:  "deleted_at",
}

// NewTierBindingDao creates and returns a new DAO object for table data access.
func NewTierBindingDao(handlers ...gdb.ModelHandler) *TierBindingDao {
	return &TierBindingDao{
		group:    "default",
		table:    "plugin_linapro_ai_tier_binding",
		columns:  tierBindingColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TierBindingDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TierBindingDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TierBindingDao) Columns() TierBindingColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TierBindingDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TierBindingDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TierBindingDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

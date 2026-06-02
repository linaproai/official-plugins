// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TierDao is the data access object for the table plugin_linapro_ai_tier.
type TierDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  TierColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// TierColumns defines and stores column names for the table plugin_linapro_ai_tier.
type TierColumns struct {
	Id                   string // Tier ID
	CapabilityType       string // Capability type: text
	Code                 string // Tier code: basic, standard, advanced
	DisplayName          string // Tier display name
	Description          string // Tier description
	DefaultEffort        string // Default thinking effort
	Enabled              string // Enabled flag: 0=disabled 1=enabled
	SortOrder            string // Stable sort order
	LastTestStatus       string // Last tier test status
	LastTestLatencyMs    string // Last tier test latency in milliseconds
	LastTestErrorSummary string // Last tier test masked error summary
	LastTestAt           string // Last tier test time
	CreatedAt            string // Creation time
	UpdatedAt            string // Update time
}

// tierColumns holds the columns for the table plugin_linapro_ai_tier.
var tierColumns = TierColumns{
	Id:                   "id",
	CapabilityType:       "capability_type",
	Code:                 "code",
	DisplayName:          "display_name",
	Description:          "description",
	DefaultEffort:        "default_effort",
	Enabled:              "enabled",
	SortOrder:            "sort_order",
	LastTestStatus:       "last_test_status",
	LastTestLatencyMs:    "last_test_latency_ms",
	LastTestErrorSummary: "last_test_error_summary",
	LastTestAt:           "last_test_at",
	CreatedAt:            "created_at",
	UpdatedAt:            "updated_at",
}

// NewTierDao creates and returns a new DAO object for table data access.
func NewTierDao(handlers ...gdb.ModelHandler) *TierDao {
	return &TierDao{
		group:    "default",
		table:    "plugin_linapro_ai_tier",
		columns:  tierColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TierDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TierDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TierDao) Columns() TierColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TierDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TierDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TierDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

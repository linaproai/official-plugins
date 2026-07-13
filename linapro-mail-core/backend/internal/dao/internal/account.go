// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AccountDao is the data access object for the table plugin_linapro_mail_core_account.
type AccountDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AccountColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AccountColumns defines and stores column names for the table plugin_linapro_mail_core_account.
type AccountColumns struct {
	Id                   string // Primary key ID
	Name                 string // Account display name
	FromAddress          string // Default From address
	OutboundConnectionId string // Outbound connection ID; 0 means none
	InboundConnectionId  string // Inbound connection ID; 0 means none
	IsDefault            string // Default account flag: 1=default, 0=normal
	Status               string // Status: 1=enabled, 0=disabled
	TenantId             string // Tenant ID; 0 means platform scope
	Remark               string // Remark
	CreatedAt            string // Creation time
	UpdatedAt            string // Update time
	DeletedAt            string // Deletion time
}

// accountColumns holds the columns for the table plugin_linapro_mail_core_account.
var accountColumns = AccountColumns{
	Id:                   "id",
	Name:                 "name",
	FromAddress:          "from_address",
	OutboundConnectionId: "outbound_connection_id",
	InboundConnectionId:  "inbound_connection_id",
	IsDefault:            "is_default",
	Status:               "status",
	TenantId:             "tenant_id",
	Remark:               "remark",
	CreatedAt:            "created_at",
	UpdatedAt:            "updated_at",
	DeletedAt:            "deleted_at",
}

// NewAccountDao creates and returns a new DAO object for table data access.
func NewAccountDao(handlers ...gdb.ModelHandler) *AccountDao {
	return &AccountDao{
		group:    "default",
		table:    "plugin_linapro_mail_core_account",
		columns:  accountColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AccountDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AccountDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AccountDao) Columns() AccountColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AccountDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AccountDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AccountDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ConnectionDao is the data access object for the table plugin_linapro_mail_core_connection.
type ConnectionDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ConnectionColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ConnectionColumns defines and stores column names for the table plugin_linapro_mail_core_connection.
type ConnectionColumns struct {
	Id        string // Primary key ID
	Name      string // Connection display name
	Kind      string // Transport kind: smtp, imap, pop3
	Host      string // Mail server host
	Port      string // Mail server port
	Username  string // Authentication username
	SecretRef string // Secret reference for password or token
	TlsMode   string // TLS mode: disable, starttls, tls
	AuthMode  string // Auth mode: password, oauth2
	ExtraJson string // Protocol extension JSON without secrets
	Status    string // Status: 1=enabled, 0=disabled
	TenantId  string // Tenant ID; 0 means platform scope
	Remark    string // Remark
	CreatedAt string // Creation time
	UpdatedAt string // Update time
	DeletedAt string // Deletion time
}

// connectionColumns holds the columns for the table plugin_linapro_mail_core_connection.
var connectionColumns = ConnectionColumns{
	Id:        "id",
	Name:      "name",
	Kind:      "kind",
	Host:      "host",
	Port:      "port",
	Username:  "username",
	SecretRef: "secret_ref",
	TlsMode:   "tls_mode",
	AuthMode:  "auth_mode",
	ExtraJson: "extra_json",
	Status:    "status",
	TenantId:  "tenant_id",
	Remark:    "remark",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewConnectionDao creates and returns a new DAO object for table data access.
func NewConnectionDao(handlers ...gdb.ModelHandler) *ConnectionDao {
	return &ConnectionDao{
		group:    "default",
		table:    "plugin_linapro_mail_core_connection",
		columns:  connectionColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ConnectionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ConnectionDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ConnectionDao) Columns() ConnectionColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ConnectionDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ConnectionDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ConnectionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

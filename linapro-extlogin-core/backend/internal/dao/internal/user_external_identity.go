// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserExternalIdentityDao is the data access object for the table plugin_linapro_extlogin_core_user_external_identity.
type UserExternalIdentityDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  UserExternalIdentityColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// UserExternalIdentityColumns defines and stores column names for the table plugin_linapro_extlogin_core_user_external_identity.
type UserExternalIdentityColumns struct {
	Id                  string // External identity linkage ID
	UserId              string // Linked local user ID
	Provider            string // Stable external provider ID owned by the calling plugin, e.g. google, discord
	Subject             string // Immutable provider-issued subject identifier, e.g. OIDC sub
	SubjectKind         string // Subject classification
	AppContext          string // Multi-app context
	PluginId            string // Calling plugin ID stamped by the host when the linkage was created
	EmailSnapshot       string // Email captured at link time for audit only, never used as a resolution key
	PhoneSnapshot       string // Phone captured at link time
	DisplayNameSnapshot string // Display name captured at link time
	AvatarUrlSnapshot   string // Avatar URL captured at link time
	CreatedAt           string // Creation time
	UpdatedAt           string // Update time
	DeletedAt           string // Soft delete time; live rows keep NULL
}

// userExternalIdentityColumns holds the columns for the table plugin_linapro_extlogin_core_user_external_identity.
var userExternalIdentityColumns = UserExternalIdentityColumns{
	Id:                  "id",
	UserId:              "user_id",
	Provider:            "provider",
	Subject:             "subject",
	SubjectKind:         "subject_kind",
	AppContext:          "app_context",
	PluginId:            "plugin_id",
	EmailSnapshot:       "email_snapshot",
	PhoneSnapshot:       "phone_snapshot",
	DisplayNameSnapshot: "display_name_snapshot",
	AvatarUrlSnapshot:   "avatar_url_snapshot",
	CreatedAt:           "created_at",
	UpdatedAt:           "updated_at",
	DeletedAt:           "deleted_at",
}

// NewUserExternalIdentityDao creates and returns a new DAO object for table data access.
func NewUserExternalIdentityDao(handlers ...gdb.ModelHandler) *UserExternalIdentityDao {
	return &UserExternalIdentityDao{
		group:    "default",
		table:    "plugin_linapro_extlogin_core_user_external_identity",
		columns:  userExternalIdentityColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserExternalIdentityDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserExternalIdentityDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserExternalIdentityDao) Columns() UserExternalIdentityColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserExternalIdentityDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserExternalIdentityDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UserExternalIdentityDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

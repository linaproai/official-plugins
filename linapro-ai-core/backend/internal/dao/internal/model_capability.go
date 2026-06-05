// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ModelCapabilityDao is the data access object for the table plugin_linapro_ai_model_capability.
type ModelCapabilityDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  ModelCapabilityColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// ModelCapabilityColumns defines and stores column names for the table plugin_linapro_ai_model_capability.
type ModelCapabilityColumns struct {
	Id                string // Model capability ID
	ModelId           string // Model ID
	EndpointId        string // Preferred provider endpoint ID, 0 means model default
	CapabilityType    string // Capability type
	CapabilityMethod  string // Capability method
	InputModalities   string // Comma-separated input modality list
	OutputModalities  string // Comma-separated output modality list
	MaxInputTokens    string // Maximum input tokens, 0 means unspecified
	MaxOutputTokens   string // Maximum output tokens, 0 means unspecified
	MaxInputAssets    string // Maximum input assets, 0 means unspecified
	MaxOutputAssets   string // Maximum output assets, 0 means unspecified
	MaxAssetBytes     string // Maximum single asset bytes, 0 means unspecified
	SupportsThinking  string // Thinking effort support flag for this model method: 0=no 1=yes
	SupportedEfforts  string // Comma-separated thinking efforts supported by this model method
	SupportsStreaming string // Streaming support flag: 0=no 1=yes
	SupportsOperation string // Provider operation support flag: 0=no 1=yes
	Enabled           string // Enabled flag: 0=disabled 1=enabled
	CreatedAt         string // Creation time
	UpdatedAt         string // Update time
	DeletedAt         string // Deletion time
}

// modelCapabilityColumns holds the columns for the table plugin_linapro_ai_model_capability.
var modelCapabilityColumns = ModelCapabilityColumns{
	Id:                "id",
	ModelId:           "model_id",
	EndpointId:        "endpoint_id",
	CapabilityType:    "capability_type",
	CapabilityMethod:  "capability_method",
	InputModalities:   "input_modalities",
	OutputModalities:  "output_modalities",
	MaxInputTokens:    "max_input_tokens",
	MaxOutputTokens:   "max_output_tokens",
	MaxInputAssets:    "max_input_assets",
	MaxOutputAssets:   "max_output_assets",
	MaxAssetBytes:     "max_asset_bytes",
	SupportsThinking:  "supports_thinking",
	SupportedEfforts:  "supported_efforts",
	SupportsStreaming: "supports_streaming",
	SupportsOperation: "supports_operation",
	Enabled:           "enabled",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
	DeletedAt:         "deleted_at",
}

// NewModelCapabilityDao creates and returns a new DAO object for table data access.
func NewModelCapabilityDao(handlers ...gdb.ModelHandler) *ModelCapabilityDao {
	return &ModelCapabilityDao{
		group:    "default",
		table:    "plugin_linapro_ai_model_capability",
		columns:  modelCapabilityColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ModelCapabilityDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ModelCapabilityDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ModelCapabilityDao) Columns() ModelCapabilityColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ModelCapabilityDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ModelCapabilityDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ModelCapabilityDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

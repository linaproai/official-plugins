// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Model is the golang structure of table plugin_linapro_ai_model for DAO operations like Where/Data.
type Model struct {
	g.Meta     `orm:"table:plugin_linapro_ai_model, do:true"`
	Id         any        // Model ID
	ProviderId any        // Provider ID
	ModelName  any        // Provider model name
	Protocol   any        // Protocol: openai or anthropic
	Source     any        // Model source: manual or api
	Enabled    any        // Enabled flag: 0=disabled 1=enabled
	CreatedAt  *time.Time // Creation time
	UpdatedAt  *time.Time // Update time
	DeletedAt  *time.Time // Deletion time
	EndpointId any        // Provider endpoint ID used by the model
}

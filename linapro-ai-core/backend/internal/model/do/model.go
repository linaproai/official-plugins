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
	g.Meta           `orm:"table:plugin_linapro_ai_model, do:true"`
	Id               any        // Model ID
	ProviderId       any        // Provider ID
	CapabilityType   any        // Capability type: text
	ModelName        any        // Provider model name
	Protocol         any        // Protocol: openai or anthropic
	Source           any        // Model source: manual or api
	SupportsThinking any        // Thinking effort support flag: 0=no 1=yes
	SupportedEfforts any        // Comma-separated supported thinking efforts
	MaxInputTokens   any        // Maximum input tokens, 0 means unspecified
	MaxOutputTokens  any        // Maximum output tokens, 0 means unspecified
	Enabled          any        // Enabled flag: 0=disabled 1=enabled
	CreatedAt        *time.Time // Creation time
	UpdatedAt        *time.Time // Update time
	DeletedAt        *time.Time // Deletion time
}

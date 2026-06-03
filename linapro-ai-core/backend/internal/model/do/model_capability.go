// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ModelCapability is the golang structure of table plugin_linapro_ai_model_capability for DAO operations like Where/Data.
type ModelCapability struct {
	g.Meta            `orm:"table:plugin_linapro_ai_model_capability, do:true"`
	Id                any        // Model capability ID
	ModelId           any        // Model ID
	EndpointId        any        // Preferred provider endpoint ID, 0 means model default
	CapabilityType    any        // Capability type
	CapabilityMethod  any        // Capability method
	InputModalities   any        // Comma-separated input modality list
	OutputModalities  any        // Comma-separated output modality list
	MaxInputTokens    any        // Maximum input tokens, 0 means unspecified
	MaxOutputTokens   any        // Maximum output tokens, 0 means unspecified
	MaxInputAssets    any        // Maximum input assets, 0 means unspecified
	MaxOutputAssets   any        // Maximum output assets, 0 means unspecified
	MaxAssetBytes     any        // Maximum single asset bytes, 0 means unspecified
	SupportsStreaming any        // Streaming support flag: 0=no 1=yes
	SupportsOperation any        // Provider operation support flag: 0=no 1=yes
	DefaultParamsJson any        // Method-specific default params JSON
	Enabled           any        // Enabled flag: 0=disabled 1=enabled
	CreatedAt         *time.Time // Creation time
	UpdatedAt         *time.Time // Update time
	DeletedAt         *time.Time // Deletion time
	SupportsThinking  any        // Thinking effort support flag for this model method: 0=no 1=yes
	SupportedEfforts  any        // Comma-separated thinking efforts supported by this model method
}

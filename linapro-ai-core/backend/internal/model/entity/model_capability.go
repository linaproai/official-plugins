// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// ModelCapability is the golang structure for table model_capability.
type ModelCapability struct {
	Id                int64      `json:"id"                orm:"id"                 description:"Model capability ID"`
	ModelId           int64      `json:"modelId"           orm:"model_id"           description:"Model ID"`
	EndpointId        int64      `json:"endpointId"        orm:"endpoint_id"        description:"Preferred provider endpoint ID, 0 means model default"`
	CapabilityType    string     `json:"capabilityType"    orm:"capability_type"    description:"Capability type"`
	CapabilityMethod  string     `json:"capabilityMethod"  orm:"capability_method"  description:"Capability method"`
	InputModalities   string     `json:"inputModalities"   orm:"input_modalities"   description:"Comma-separated input modality list"`
	OutputModalities  string     `json:"outputModalities"  orm:"output_modalities"  description:"Comma-separated output modality list"`
	MaxInputTokens    int        `json:"maxInputTokens"    orm:"max_input_tokens"   description:"Maximum input tokens, 0 means unspecified"`
	MaxOutputTokens   int        `json:"maxOutputTokens"   orm:"max_output_tokens"  description:"Maximum output tokens, 0 means unspecified"`
	MaxInputAssets    int        `json:"maxInputAssets"    orm:"max_input_assets"   description:"Maximum input assets, 0 means unspecified"`
	MaxOutputAssets   int        `json:"maxOutputAssets"   orm:"max_output_assets"  description:"Maximum output assets, 0 means unspecified"`
	MaxAssetBytes     int64      `json:"maxAssetBytes"     orm:"max_asset_bytes"    description:"Maximum single asset bytes, 0 means unspecified"`
	SupportsThinking  int        `json:"supportsThinking"  orm:"supports_thinking"  description:"Thinking effort support flag for this model method: 0=no 1=yes"`
	SupportedEfforts  string     `json:"supportedEfforts"  orm:"supported_efforts"  description:"Comma-separated thinking efforts supported by this model method"`
	SupportsStreaming int        `json:"supportsStreaming" orm:"supports_streaming" description:"Streaming support flag: 0=no 1=yes"`
	SupportsOperation int        `json:"supportsOperation" orm:"supports_operation" description:"Provider operation support flag: 0=no 1=yes"`
	Enabled           int        `json:"enabled"           orm:"enabled"            description:"Enabled flag: 0=disabled 1=enabled"`
	CreatedAt         *time.Time `json:"createdAt"         orm:"created_at"         description:"Creation time"`
	UpdatedAt         *time.Time `json:"updatedAt"         orm:"updated_at"         description:"Update time"`
	DeletedAt         *time.Time `json:"deletedAt"         orm:"deleted_at"         description:"Deletion time"`
}

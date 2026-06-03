// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ProviderEndpoint is the golang structure of table plugin_linapro_ai_provider_endpoint for DAO operations like Where/Data.
type ProviderEndpoint struct {
	g.Meta       `orm:"table:plugin_linapro_ai_provider_endpoint, do:true"`
	Id           any        // Provider endpoint ID
	ProviderId   any        // Provider ID
	Protocol     any        // Protocol: openai, anthropic, voyage, openai-compatible, or provider-specific
	BaseUrl      any        // Provider protocol base URL
	SecretRef    any        // Secret reference or masked secret reference
	Enabled      any        // Enabled flag: 0=disabled 1=enabled
	MetadataJson any        // Endpoint metadata JSON without secret values
	CreatedAt    *time.Time // Creation time
	UpdatedAt    *time.Time // Update time
	DeletedAt    *time.Time // Deletion time
}

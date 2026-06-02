// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Invocation is the golang structure of table plugin_linapro_ai_invocation for DAO operations like Where/Data.
type Invocation struct {
	g.Meta         `orm:"table:plugin_linapro_ai_invocation, do:true"`
	Id             any        // Invocation ID
	RequestId      any        // Request correlation ID
	CapabilityType any        // Capability type
	Purpose        any        // Governed AI purpose
	TierCode       any        // Tier code
	SourcePluginId any        // Source plugin ID
	TenantId       any        // Tenant ID
	UserId         any        // User ID
	ProviderId     any        // Provider ID
	ModelId        any        // Model ID
	ProviderName   any        // Provider display name snapshot
	ModelName      any        // Model name snapshot
	Protocol       any        // Protocol snapshot
	ThinkingEffort any        // Requested or applied thinking effort
	Status         any        // Invocation status: success or failed
	InputTokens    any        // Input token count
	OutputTokens   any        // Output token count
	LatencyMs      any        // Provider call latency in milliseconds
	ErrorCode      any        // Stable error code
	ErrorSummary   any        // Masked error summary
	CreatedAt      *time.Time // Creation time
}

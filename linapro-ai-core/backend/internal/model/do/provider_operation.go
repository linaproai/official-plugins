// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ProviderOperation is the golang structure of table plugin_linapro_ai_provider_operation for DAO operations like Where/Data.
type ProviderOperation struct {
	g.Meta           `orm:"table:plugin_linapro_ai_provider_operation, do:true"`
	Id               any        // Provider operation row ID
	OperationRef     any        // Opaque provider operation reference
	CapabilityType   any        // Capability type
	CapabilityMethod any        // Capability method
	Purpose          any        // Governed AI purpose
	SourcePluginId   any        // Source plugin ID
	ProviderId       any        // Provider ID
	ModelId          any        // Model ID
	ProviderName     any        // Provider display name snapshot
	ModelName        any        // Model name snapshot
	Protocol         any        // Protocol snapshot
	Status           any        // Provider operation status
	NextPollAfterMs  any        // Recommended next poll delay in milliseconds
	ExpiresAt        *time.Time // Operation reference expiration time
	AssetSummaryJson any        // Asset reference summary JSON without file contents
	ErrorCode        any        // Stable error code
	ErrorSummary     any        // Masked error summary
	CreatedAt        *time.Time // Creation time
	UpdatedAt        *time.Time // Update time
	DeletedAt        *time.Time // Deletion time
}

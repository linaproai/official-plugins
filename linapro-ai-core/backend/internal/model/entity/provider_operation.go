// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// ProviderOperation is the golang structure for table provider_operation.
type ProviderOperation struct {
	Id               int64      `json:"id"               orm:"id"                 description:"Provider operation row ID"`
	OperationRef     string     `json:"operationRef"     orm:"operation_ref"      description:"Opaque provider operation reference"`
	CapabilityType   string     `json:"capabilityType"   orm:"capability_type"    description:"Capability type"`
	CapabilityMethod string     `json:"capabilityMethod" orm:"capability_method"  description:"Capability method"`
	Purpose          string     `json:"purpose"          orm:"purpose"            description:"Governed AI purpose"`
	SourcePluginId   string     `json:"sourcePluginId"   orm:"source_plugin_id"   description:"Source plugin ID"`
	ProviderId       int64      `json:"providerId"       orm:"provider_id"        description:"Provider ID"`
	ModelId          int64      `json:"modelId"          orm:"model_id"           description:"Model ID"`
	ProviderName     string     `json:"providerName"     orm:"provider_name"      description:"Provider display name snapshot"`
	ModelName        string     `json:"modelName"        orm:"model_name"         description:"Model name snapshot"`
	Protocol         string     `json:"protocol"         orm:"protocol"           description:"Protocol snapshot"`
	Status           string     `json:"status"           orm:"status"             description:"Provider operation status"`
	NextPollAfterMs  int64      `json:"nextPollAfterMs"  orm:"next_poll_after_ms" description:"Recommended next poll delay in milliseconds"`
	ExpiresAt        *time.Time `json:"expiresAt"        orm:"expires_at"         description:"Operation reference expiration time"`
	AssetSummaryJson string     `json:"assetSummaryJson" orm:"asset_summary_json" description:"Asset reference summary JSON without file contents"`
	ErrorCode        string     `json:"errorCode"        orm:"error_code"         description:"Stable error code"`
	ErrorSummary     string     `json:"errorSummary"     orm:"error_summary"      description:"Masked error summary"`
	CreatedAt        *time.Time `json:"createdAt"        orm:"created_at"         description:"Creation time"`
	UpdatedAt        *time.Time `json:"updatedAt"        orm:"updated_at"         description:"Update time"`
	DeletedAt        *time.Time `json:"deletedAt"        orm:"deleted_at"         description:"Deletion time"`
}

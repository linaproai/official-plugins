// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Invocation is the golang structure for table invocation.
type Invocation struct {
	Id                   int64      `json:"id"                   orm:"id"                     description:"Invocation ID"`
	RequestId            string     `json:"requestId"            orm:"request_id"             description:"Request correlation ID"`
	CapabilityType       string     `json:"capabilityType"       orm:"capability_type"        description:"Capability type"`
	CapabilityMethod     string     `json:"capabilityMethod"     orm:"capability_method"      description:"Capability method"`
	Purpose              string     `json:"purpose"              orm:"purpose"                description:"Governed AI purpose"`
	TierCode             string     `json:"tierCode"             orm:"tier_code"              description:"Tier code"`
	SourcePluginId       string     `json:"sourcePluginId"       orm:"source_plugin_id"       description:"Source plugin ID"`
	TenantId             int        `json:"tenantId"             orm:"tenant_id"              description:"Tenant ID"`
	UserId               int        `json:"userId"               orm:"user_id"                description:"User ID"`
	ProviderId           int64      `json:"providerId"           orm:"provider_id"            description:"Provider ID"`
	ModelId              int64      `json:"modelId"              orm:"model_id"               description:"Model ID"`
	ProviderName         string     `json:"providerName"         orm:"provider_name"          description:"Provider display name snapshot"`
	ModelName            string     `json:"modelName"            orm:"model_name"             description:"Model name snapshot"`
	Protocol             string     `json:"protocol"             orm:"protocol"               description:"Protocol snapshot"`
	ThinkingEffort       string     `json:"thinkingEffort"       orm:"thinking_effort"        description:"Requested or applied thinking effort"`
	Status               string     `json:"status"               orm:"status"                 description:"Invocation status: success or failed"`
	InputTokens          int        `json:"inputTokens"          orm:"input_tokens"           description:"Input token count"`
	OutputTokens         int        `json:"outputTokens"         orm:"output_tokens"          description:"Output token count"`
	LatencyMs            int        `json:"latencyMs"            orm:"latency_ms"             description:"Provider call latency in milliseconds"`
	ErrorCode            string     `json:"errorCode"            orm:"error_code"             description:"Stable error code"`
	ErrorSummary         string     `json:"errorSummary"         orm:"error_summary"          description:"Masked error summary"`
	CreatedAt            *time.Time `json:"createdAt"            orm:"created_at"             description:"Creation time"`
	AssetSummaryJson     string     `json:"assetSummaryJson"     orm:"asset_summary_json"     description:"Asset reference summary JSON without file contents"`
	OperationSummaryJson string     `json:"operationSummaryJson" orm:"operation_summary_json" description:"Provider operation summary JSON without provider secrets"`
	MetadataSummaryJson  string     `json:"metadataSummaryJson"  orm:"metadata_summary_json"  description:"Bounded metadata summary JSON without request or response bodies"`
}

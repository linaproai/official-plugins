// This file declares shared AI invocation log API DTO projections.

package v1

// InvocationItem is the masked invocation log projection returned by the list API.
type InvocationItem struct {
	Id                   int64  `json:"id" dc:"Invocation ID" eg:"1"`
	RequestId            string `json:"requestId" dc:"Request correlation ID" eg:"req_123"`
	CapabilityType       string `json:"capabilityType" dc:"Capability type" eg:"text"`
	CapabilityMethod     string `json:"capabilityMethod" dc:"Capability method within the type" eg:"generate"`
	Purpose              string `json:"purpose" dc:"Governed AI purpose" eg:"git.commitMessage"`
	TierCode             string `json:"tierCode" dc:"Tier code: basic, standard, advanced" eg:"basic"`
	SourcePluginId       string `json:"sourcePluginId" dc:"Source plugin ID that initiated the call" eg:"linapro-demo-source"`
	TenantId             int    `json:"tenantId" dc:"Tenant ID associated with the call" eg:"0"`
	UserId               int    `json:"userId" dc:"User ID associated with the call" eg:"1"`
	ProviderId           int64  `json:"providerId" dc:"Provider ID snapshot" eg:"1"`
	ModelId              int64  `json:"modelId" dc:"Model ID snapshot" eg:"1"`
	ProviderName         string `json:"providerName" dc:"Provider display name snapshot" eg:"OpenAI"`
	ModelName            string `json:"modelName" dc:"Model name snapshot" eg:"gpt-4.1-mini"`
	Protocol             string `json:"protocol" dc:"Provider protocol snapshot" eg:"openai"`
	ThinkingEffort       string `json:"thinkingEffort" dc:"Requested or applied thinking effort" eg:"low"`
	Status               string `json:"status" dc:"Invocation status: success or failed" eg:"success"`
	InputTokens          int    `json:"inputTokens" dc:"Input token count" eg:"32"`
	OutputTokens         int    `json:"outputTokens" dc:"Output token count" eg:"64"`
	LatencyMs            int    `json:"latencyMs" dc:"Provider call latency in milliseconds" eg:"300"`
	AssetSummaryJson     string `json:"assetSummaryJson" dc:"Asset reference summary JSON without file content, provider URLs, or secrets" eg:"{}"`
	OperationSummaryJson string `json:"operationSummaryJson" dc:"Provider operation summary JSON without provider request IDs that expose secrets" eg:"{}"`
	MetadataSummaryJson  string `json:"metadataSummaryJson" dc:"Non-sensitive invocation metadata summary JSON" eg:"{}"`
	ErrorCode            string `json:"errorCode" dc:"Stable error code for failed invocations" eg:"AI_CORE_TIER_BINDING_UNAVAILABLE"`
	ErrorSummary         string `json:"errorSummary" dc:"Masked error summary. Full prompts, responses, and secrets are never returned." eg:"Provider returned 401"`
	CreatedAt            int64  `json:"createdAt" dc:"Creation time, Unix timestamp in milliseconds" eg:"1717200000000"`
}

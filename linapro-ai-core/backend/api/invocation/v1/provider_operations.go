// This file declares provider operation query DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ProviderOperationItem is one masked provider operation projection.
type ProviderOperationItem struct {
	Id               int64  `json:"id" dc:"Provider operation ID" eg:"1"`
	OperationRef     string `json:"operationRef" dc:"Opaque provider operation reference" eg:"op_opaque_123"`
	CapabilityType   string `json:"capabilityType" dc:"Capability type such as image, audio, or video" eg:"video"`
	CapabilityMethod string `json:"capabilityMethod" dc:"Capability method such as generate, edit, extend, operation.get, or operation.cancel" eg:"generate"`
	Purpose          string `json:"purpose" dc:"Governed AI purpose" eg:"media.storyboard"`
	SourcePluginId   string `json:"sourcePluginId" dc:"Source plugin ID that initiated the operation" eg:"linapro-demo-source"`
	ProviderId       int64  `json:"providerId" dc:"Provider ID snapshot" eg:"1"`
	ModelId          int64  `json:"modelId" dc:"Model ID snapshot" eg:"1"`
	ProviderName     string `json:"providerName" dc:"Provider display name snapshot" eg:"OpenAI"`
	ModelName        string `json:"modelName" dc:"Model name snapshot" eg:"sora-2"`
	Protocol         string `json:"protocol" dc:"Provider protocol snapshot" eg:"openai"`
	Status           string `json:"status" dc:"Provider operation status" eg:"running"`
	NextPollAfterMs  int64  `json:"nextPollAfterMs" dc:"Suggested next provider poll delay in milliseconds" eg:"3000"`
	ExpiresAt        int64  `json:"expiresAt" dc:"Operation expiration time, Unix timestamp in milliseconds" eg:"1717203600000"`
	AssetSummaryJson string `json:"assetSummaryJson" dc:"Asset reference summary JSON without file content, provider URLs, or secrets" eg:"{}"`
	ErrorCode        string `json:"errorCode" dc:"Stable error code for failed operations" eg:"AI_CORE_PROVIDER_UNAVAILABLE"`
	ErrorSummary     string `json:"errorSummary" dc:"Masked error summary. Provider responses and secrets are never returned." eg:"Provider operation failed"`
	CreatedAt        int64  `json:"createdAt" dc:"Creation time, Unix timestamp in milliseconds" eg:"1717200000000"`
	UpdatedAt        int64  `json:"updatedAt" dc:"Update time, Unix timestamp in milliseconds" eg:"1717200000000"`
}

// ListProviderOperationsReq defines the request for querying provider operations.
type ListProviderOperationsReq struct {
	g.Meta           `path:"/ai/provider-operations" method:"get" tags:"AI Provider Operations" summary:"List AI provider operations" dc:"Query masked provider operation projections by page with filters for capability method, purpose, status, provider, model, source plugin, and time range. Business task state and provider secret data are never returned." permission:"ai:invocation:list"`
	PageNum          int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize         int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	CapabilityType   string `json:"capabilityType" dc:"Capability type filter" eg:"video"`
	CapabilityMethod string `json:"capabilityMethod" dc:"Capability method filter" eg:"generate"`
	Purpose          string `json:"purpose" dc:"Purpose filter with exact match" eg:"media.storyboard"`
	Status           string `json:"status" dc:"Status filter" eg:"running"`
	ProviderId       int64  `json:"providerId" dc:"Provider ID filter" eg:"1"`
	ModelId          int64  `json:"modelId" dc:"Model ID filter" eg:"1"`
	SourcePluginId   string `json:"sourcePluginId" dc:"Source plugin ID filter" eg:"linapro-demo-source"`
	StartedAt        int64  `json:"startedAt" dc:"Start time filter, Unix timestamp in milliseconds" eg:"1717200000000"`
	EndedAt          int64  `json:"endedAt" dc:"End time filter, Unix timestamp in milliseconds" eg:"1717286400000"`
}

// ListProviderOperationsRes defines the response for querying provider operations.
type ListProviderOperationsRes struct {
	List  []*ProviderOperationItem `json:"list" dc:"Masked provider operation list" eg:"[]"`
	Total int                      `json:"total" dc:"Total number of provider operations matching filters" eg:"20"`
}

// GetProviderOperationReq defines the request for reading one provider operation.
type GetProviderOperationReq struct {
	g.Meta       `path:"/ai/provider-operations/{operationRef}" method:"get" tags:"AI Provider Operations" summary:"Get AI provider operation" dc:"Get one masked provider operation projection by opaque operation reference. Business task state and provider secret data are never returned." permission:"ai:invocation:list"`
	OperationRef string `json:"operationRef" v:"required" dc:"Opaque provider operation reference" eg:"op_opaque_123"`
}

// GetProviderOperationRes defines the response for reading one provider operation.
type GetProviderOperationRes struct {
	ProviderOperationItem
}

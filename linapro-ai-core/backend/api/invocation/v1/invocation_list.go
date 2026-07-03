// This file declares the invocation-log list request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq defines the request for querying masked AI invocation logs.
type ListReq struct {
	g.Meta           `path:"/ai/invocations" method:"get" tags:"Smart Center/Invocation Logs" summary:"List AI invocation logs" dc:"Query masked AI invocation logs by page with filters for capability type, capability method, purpose, tier, status, provider, model, source plugin, and time range. Full prompts and responses are never returned." permission:"ai:invocation:list"`
	PageNum          int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize         int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	CapabilityType   string `json:"capabilityType" dc:"Capability type filter" eg:"text"`
	CapabilityMethod string `json:"capabilityMethod" dc:"Capability method filter" eg:"generate"`
	Purpose          string `json:"purpose" dc:"Purpose filter with exact match" eg:"git.commitMessage"`
	TierCode         string `json:"tierCode" dc:"Tier code filter: basic, standard, advanced" eg:"basic"`
	Status           string `json:"status" dc:"Status filter: success or failed" eg:"success"`
	ProviderId       int64  `json:"providerId" dc:"Provider ID filter" eg:"1"`
	ModelId          int64  `json:"modelId" dc:"Model ID filter" eg:"1"`
	SourcePluginId   string `json:"sourcePluginId" dc:"Source plugin ID filter" eg:"linapro-demo-source"`
	StartedAt        int64  `json:"startedAt" dc:"Start time filter, Unix timestamp in milliseconds" eg:"1717200000000"`
	EndedAt          int64  `json:"endedAt" dc:"End time filter, Unix timestamp in milliseconds" eg:"1717286400000"`
}

// ListRes defines the response for querying masked AI invocation logs.
type ListRes struct {
	List  []*InvocationItem `json:"list" dc:"Masked AI invocation log list" eg:"[]"`
	Total int               `json:"total" dc:"Total number of invocation logs matching filters" eg:"20"`
}

// This file declares provider-owned model list, create, and sync DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListModelsReq defines the request for listing provider models.
type ListModelsReq struct {
	g.Meta         `path:"/ai/providers/{providerId}/models" method:"get" tags:"AI Provider Models" summary:"List provider models" dc:"List models belonging to one AI provider with capability and enabled filters." permission:"ai:provider:list"`
	ProviderId     int64  `json:"providerId" v:"required|min:1" dc:"Provider ID" eg:"1"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Page size, capped at 100" eg:"10"`
	CapabilityType string `json:"capabilityType" d:"text" dc:"Capability type filter; first release supports text" eg:"text"`
	Enabled        *int   `json:"enabled" dc:"Optional enabled filter: 0=disabled 1=enabled" eg:"1"`
}

// ListModelsRes defines the response for listing provider models.
type ListModelsRes struct {
	List  []*ModelItem `json:"list" dc:"Provider model list" eg:"[]"`
	Total int          `json:"total" dc:"Total provider models matching filters" eg:"3"`
}

// CreateModelReq defines the request for creating a provider model.
type CreateModelReq struct {
	g.Meta           `path:"/ai/providers/{providerId}/models" method:"post" tags:"AI Provider Models" summary:"Create provider model" dc:"Create one model under an AI provider and declare text capability metadata, token bounds, and thinking effort support." permission:"ai:provider:create"`
	ProviderId       int64    `json:"providerId" v:"required|min:1" dc:"Provider ID" eg:"1"`
	CapabilityType   string   `json:"capabilityType" d:"text" dc:"Capability type; first release supports text" eg:"text"`
	ModelName        string   `json:"modelName" v:"required|max-length:128" dc:"Provider model name" eg:"gpt-4.1-mini"`
	Protocol         string   `json:"protocol" v:"required|in:openai,anthropic" dc:"Provider protocol: openai or anthropic" eg:"openai"`
	SupportsThinking int      `json:"supportsThinking" dc:"Thinking effort support flag: 0=no 1=yes" eg:"1"`
	SupportedEfforts []string `json:"supportedEfforts" dc:"Supported thinking efforts: low, medium, high, xhigh, max" eg:"low,medium,high"`
	MaxInputTokens   int      `json:"maxInputTokens" dc:"Maximum input tokens; 0 means unspecified" eg:"128000"`
	MaxOutputTokens  int      `json:"maxOutputTokens" dc:"Maximum output tokens; 0 means unspecified" eg:"8192"`
	Enabled          int      `json:"enabled" d:"1" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
}

// CreateModelRes defines the response for creating a provider model.
type CreateModelRes struct {
	Id int64 `json:"id" dc:"Created model ID" eg:"1"`
}

// SyncModelsReq defines the request for synchronizing provider models.
type SyncModelsReq struct {
	g.Meta     `path:"/ai/providers/{providerId}/models/sync" method:"post" tags:"AI Provider Models" summary:"Sync provider models" dc:"Synchronize public model metadata from the provider protocol. Failures keep existing manual and referenced models unchanged." permission:"ai:provider:update"`
	ProviderId int64  `json:"providerId" v:"required|min:1" dc:"Provider ID" eg:"1"`
	Protocol   string `json:"protocol" v:"required|in:openai,anthropic" dc:"Protocol used for model synchronization: openai or anthropic" eg:"openai"`
}

// SyncModelsRes defines the response for synchronizing provider models.
type SyncModelsRes struct {
	Created int `json:"created" dc:"Number of models created by synchronization" eg:"2"`
	Kept    int `json:"kept" dc:"Number of existing models kept unchanged" eg:"3"`
}

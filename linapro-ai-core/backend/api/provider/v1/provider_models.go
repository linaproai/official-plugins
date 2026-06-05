// This file declares provider-owned model list, create, and sync DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListModelsReq defines the request for listing provider models.
type ListModelsReq struct {
	g.Meta     `path:"/ai/providers/{providerId}/models" method:"get" tags:"AI Provider Models" summary:"List provider models" dc:"List model identities belonging to one AI provider with bounded pagination. Model candidates are not filtered by capability method declarations." permission:"ai:provider:list"`
	ProviderId int64 `json:"providerId" v:"required|min:1" dc:"Provider ID" eg:"1"`
	PageNum    int   `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize   int   `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Page size, capped at 100" eg:"10"`
	Enabled    *int  `json:"enabled" dc:"Optional enabled filter: 0=disabled 1=enabled" eg:"1"`
}

// ListModelsRes defines the response for listing provider models.
type ListModelsRes struct {
	List  []*ModelItem `json:"list" dc:"Provider model list" eg:"[]"`
	Total int          `json:"total" dc:"Total provider models matching filters" eg:"3"`
}

// CreateModelReq defines the request for creating a provider model.
type CreateModelReq struct {
	g.Meta     `path:"/ai/providers/{providerId}/models" method:"post" tags:"AI Provider Models" summary:"Create provider model" dc:"Create one model identity under an AI provider. Capability method support is decided by administrators through tier selection and tests, not by model declarations." permission:"ai:provider:create"`
	ProviderId int64  `json:"providerId" v:"required|min:1" dc:"Provider ID" eg:"1"`
	EndpointId int64  `json:"endpointId" v:"required|min:1" dc:"Provider endpoint ID used for this model" eg:"1"`
	ModelName  string `json:"modelName" v:"required|max-length:128" dc:"Provider model name" eg:"gpt-4.1-mini"`
	Protocol   string `json:"protocol" v:"required|in:openai,anthropic,voyage,openai-compatible,anthropic-compatible" dc:"Provider protocol: openai, anthropic, voyage, openai-compatible, or anthropic-compatible" eg:"openai-compatible"`
	Enabled    int    `json:"enabled" d:"1" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
}

// CreateModelRes defines the response for creating a provider model.
type CreateModelRes struct {
	Id int64 `json:"id" dc:"Created model ID" eg:"1"`
}

// SyncModelsReq defines the request for synchronizing provider models.
type SyncModelsReq struct {
	g.Meta     `path:"/ai/providers/{providerId}/models/sync" method:"post" tags:"AI Provider Models" summary:"Sync provider models" dc:"Synchronize public model metadata from enabled provider endpoints. When protocol is omitted, all enabled syncable endpoints are queried and partial endpoint failures keep successful endpoint results." permission:"ai:provider:update"`
	ProviderId int64  `json:"providerId" v:"required|min:1" dc:"Provider ID" eg:"1"`
	Protocol   string `json:"protocol" v:"in:openai,anthropic,voyage,openai-compatible,anthropic-compatible" dc:"Optional protocol used for narrow model synchronization; empty queries all enabled syncable endpoints" eg:"openai"`
}

// SyncModelsRes defines the response for synchronizing provider models.
type SyncModelsRes struct {
	Created int `json:"created" dc:"Number of models created by synchronization" eg:"2"`
	Kept    int `json:"kept" dc:"Number of existing models kept unchanged" eg:"3"`
}

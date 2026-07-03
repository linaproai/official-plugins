// This file declares model-dimension list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq defines the request for listing all AI models.
type ListReq struct {
	g.Meta     `path:"/ai/models" method:"get" tags:"Smart Center/Model Management" summary:"List AI models" dc:"List all AI model identities from the platform Smart Center with bounded pagination and provider and endpoint projections assembled in batches. Model management does not filter by capability method declarations." permission:"ai:provider:list"`
	PageNum    int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize   int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Page size, capped at 100" eg:"10"`
	Keyword    string `json:"keyword" dc:"Optional model name keyword filter" eg:"gpt"`
	ProviderId int64  `json:"providerId" dc:"Optional provider ID filter; zero means all providers" eg:"1"`
	Enabled    *int   `json:"enabled" dc:"Optional enabled filter: 0=disabled 1=enabled" eg:"1"`
}

// ListRes defines the response for listing all AI models.
type ListRes struct {
	List  []*ModelItem `json:"list" dc:"AI model list" eg:"[]"`
	Total int          `json:"total" dc:"Total AI models matching filters" eg:"3"`
}

// ModelItem is the AI model projection returned by model-dimension APIs.
type ModelItem struct {
	Id              int64  `json:"id" dc:"Model ID" eg:"1"`
	ProviderId      int64  `json:"providerId" dc:"Owning provider ID" eg:"1"`
	ProviderName    string `json:"providerName" dc:"Owning provider display name" eg:"OpenAI"`
	EndpointId      int64  `json:"endpointId" dc:"Default provider endpoint ID used by this model" eg:"1"`
	EndpointBaseUrl string `json:"endpointBaseUrl" dc:"Protocol endpoint base URL" eg:"https://api.openai.com/v1"`
	ModelName       string `json:"modelName" dc:"Provider model name" eg:"gpt-4.1-mini"`
	Protocol        string `json:"protocol" dc:"Provider protocol: openai, anthropic, voyage, openai-compatible, or anthropic-compatible" eg:"openai"`
	Source          string `json:"source" dc:"Model source: manual or api" eg:"manual"`
	Enabled         int    `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	CreatedAt       int64  `json:"createdAt" dc:"Creation time, Unix timestamp in milliseconds" eg:"1717200000000"`
	UpdatedAt       int64  `json:"updatedAt" dc:"Update time, Unix timestamp in milliseconds" eg:"1717200000000"`
}

// This file declares shared AI provider and model API DTO projections.

package v1

// ProviderItem is the provider projection returned by list and detail APIs.
type ProviderItem struct {
	Id                   int64                       `json:"id" dc:"Provider ID" eg:"1"`
	Name                 string                      `json:"name" dc:"Provider display name" eg:"OpenAI"`
	WebsiteUrl           string                      `json:"websiteUrl" dc:"Provider website URL" eg:"https://openai.com"`
	Remark               string                      `json:"remark" dc:"Provider remark" eg:"Production multimodal models"`
	Enabled              int                         `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	ModelCount           int                         `json:"modelCount" dc:"Number of models under this provider" eg:"3"`
	EnabledModelCount    int                         `json:"enabledModelCount" dc:"Number of enabled models under this provider" eg:"2"`
	EndpointCount        int                         `json:"endpointCount" dc:"Number of protocol endpoints under this provider" eg:"2"`
	EnabledEndpointCount int                         `json:"enabledEndpointCount" dc:"Number of enabled protocol endpoints under this provider" eg:"1"`
	Models               []*ProviderModelSummaryItem `json:"models" dc:"Provider-owned model summaries for current list rendering" eg:"[]"`
	Endpoints            []*ProviderEndpointItem     `json:"endpoints" dc:"Provider-owned endpoint summaries for current list rendering" eg:"[]"`
	CreatedAt            int64                       `json:"createdAt" dc:"Creation time, Unix timestamp in milliseconds" eg:"1717200000000"`
	UpdatedAt            int64                       `json:"updatedAt" dc:"Update time, Unix timestamp in milliseconds" eg:"1717200000000"`
}

// ProviderEndpointItem is the provider protocol endpoint projection returned by provider APIs.
type ProviderEndpointItem struct {
	Id           int64  `json:"id" dc:"Provider endpoint ID" eg:"1"`
	ProviderId   int64  `json:"providerId" dc:"Owning provider ID" eg:"1"`
	Protocol     string `json:"protocol" dc:"Provider protocol: openai, anthropic, voyage, openai-compatible, or anthropic-compatible" eg:"openai-compatible"`
	BaseUrl      string `json:"baseUrl" dc:"Protocol endpoint base URL" eg:"https://api.openai.com/v1"`
	SecretRef    string `json:"secretRef" dc:"Masked endpoint secret reference; plaintext API keys are never returned" eg:"sk-**********cd"`
	Enabled      int    `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	MetadataJson string `json:"metadataJson" dc:"Endpoint metadata JSON without provider secrets" eg:"{}"`
	CreatedAt    int64  `json:"createdAt" dc:"Creation time, Unix timestamp in milliseconds" eg:"1717200000000"`
	UpdatedAt    int64  `json:"updatedAt" dc:"Update time, Unix timestamp in milliseconds" eg:"1717200000000"`
}

// ProviderEndpointSaveItem is the fixed provider-form endpoint payload.
type ProviderEndpointSaveItem struct {
	Id           int64  `json:"id" dc:"Existing provider endpoint ID; zero creates a new endpoint" eg:"1"`
	Protocol     string `json:"protocol" v:"required|in:openai,anthropic" dc:"Provider form endpoint protocol: openai or anthropic" eg:"openai"`
	BaseUrl      string `json:"baseUrl" v:"max-length:512" dc:"Protocol endpoint base URL; empty with an existing ID removes that endpoint after reference checks" eg:"https://api.openai.com/v1"`
	SecretRef    string `json:"secretRef" v:"max-length:512" dc:"Endpoint secret reference; empty or masked values keep the existing reference" eg:"sk-live-secret"`
	Enabled      int    `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	MetadataJson string `json:"metadataJson" v:"max-length:2048" dc:"Endpoint metadata JSON without provider secrets" eg:"{}"`
}

// ProviderModelSummaryItem is the compact model projection embedded in provider lists.
type ProviderModelSummaryItem struct {
	Id        int64  `json:"id" dc:"Model ID" eg:"1"`
	ModelName string `json:"modelName" dc:"Provider model name" eg:"gpt-4.1-mini"`
	Protocol  string `json:"protocol" dc:"Provider protocol: openai, anthropic, voyage, openai-compatible, or anthropic-compatible" eg:"openai"`
	Enabled   int    `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
}

// ModelItem is the AI model projection returned by provider model APIs.
type ModelItem struct {
	Id         int64  `json:"id" dc:"Model ID" eg:"1"`
	ProviderId int64  `json:"providerId" dc:"Owning provider ID" eg:"1"`
	EndpointId int64  `json:"endpointId" dc:"Default provider endpoint ID used by this model" eg:"1"`
	ModelName  string `json:"modelName" dc:"Provider model name" eg:"gpt-4.1-mini"`
	Protocol   string `json:"protocol" dc:"Provider protocol: openai, anthropic, voyage, openai-compatible, or anthropic-compatible" eg:"openai"`
	Source     string `json:"source" dc:"Model source: manual or api" eg:"manual"`
	Enabled    int    `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	CreatedAt  int64  `json:"createdAt" dc:"Creation time, Unix timestamp in milliseconds" eg:"1717200000000"`
	UpdatedAt  int64  `json:"updatedAt" dc:"Update time, Unix timestamp in milliseconds" eg:"1717200000000"`
}

// This file declares provider endpoint management DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListProviderEndpointsReq defines the request for listing provider endpoints.
type ListProviderEndpointsReq struct {
	g.Meta     `path:"/ai/providers/{providerId}/endpoints" method:"get" tags:"Smart Center/Channel Management" summary:"List provider endpoints" dc:"List protocol endpoints under one AI provider with optional protocol and enabled filters. Plaintext API keys are never returned." permission:"ai:provider:list"`
	ProviderId int64  `json:"providerId" v:"required|min:1" dc:"Provider ID" eg:"1"`
	Protocol   string `json:"protocol" dc:"Optional protocol filter: openai, anthropic, voyage, openai-compatible, or anthropic-compatible" eg:"openai-compatible"`
	Enabled    *int   `json:"enabled" dc:"Optional enabled filter: 0=disabled 1=enabled" eg:"1"`
}

// ListProviderEndpointsRes defines the response for listing provider endpoints.
type ListProviderEndpointsRes struct {
	List []*ProviderEndpointItem `json:"list" dc:"Provider endpoint list" eg:"[]"`
}

// CreateProviderEndpointReq defines the request for creating a provider endpoint.
type CreateProviderEndpointReq struct {
	g.Meta       `path:"/ai/providers/{providerId}/endpoints" method:"post" tags:"Smart Center/Channel Management" summary:"Create provider endpoint" dc:"Create one protocol endpoint under an AI provider. The response returns only the generated endpoint ID and never returns plaintext secrets." permission:"ai:provider:create"`
	ProviderId   int64  `json:"providerId" v:"required|min:1" dc:"Provider ID" eg:"1"`
	Protocol     string `json:"protocol" v:"required|in:openai,anthropic,voyage,openai-compatible,anthropic-compatible" dc:"Provider protocol: openai, anthropic, voyage, openai-compatible, or anthropic-compatible" eg:"openai-compatible"`
	BaseUrl      string `json:"baseUrl" v:"required|max-length:512" dc:"Protocol endpoint base URL" eg:"https://api.openai.com/v1"`
	SecretRef    string `json:"secretRef" v:"max-length:512" dc:"Endpoint secret reference or secret value to store through the configured secret handling path" eg:"sk-live-secret"`
	Enabled      int    `json:"enabled" d:"1" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	MetadataJson string `json:"metadataJson" dc:"Endpoint metadata JSON without provider secrets" eg:"{}"`
}

// CreateProviderEndpointRes defines the response for creating a provider endpoint.
type CreateProviderEndpointRes struct {
	Id int64 `json:"id" dc:"Created provider endpoint ID" eg:"1"`
}

// UpdateProviderEndpointReq defines the request for updating a provider endpoint.
type UpdateProviderEndpointReq struct {
	g.Meta       `path:"/ai/providers/{providerId}/endpoints/{id}" method:"put" tags:"Smart Center/Channel Management" summary:"Update provider endpoint" dc:"Update one protocol endpoint under an AI provider. Empty or masked secret references keep the existing secret reference." permission:"ai:provider:update"`
	ProviderId   int64  `json:"providerId" v:"required|min:1" dc:"Provider ID" eg:"1"`
	Id           int64  `json:"id" v:"required|min:1" dc:"Provider endpoint ID" eg:"1"`
	Protocol     string `json:"protocol" v:"required|in:openai,anthropic,voyage,openai-compatible,anthropic-compatible" dc:"Provider protocol: openai, anthropic, voyage, openai-compatible, or anthropic-compatible" eg:"openai-compatible"`
	BaseUrl      string `json:"baseUrl" v:"required|max-length:512" dc:"Protocol endpoint base URL" eg:"https://api.openai.com/v1"`
	SecretRef    string `json:"secretRef" v:"max-length:512" dc:"Endpoint secret reference; empty or masked values keep the existing reference" eg:"sk-**********cd"`
	Enabled      int    `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	MetadataJson string `json:"metadataJson" dc:"Endpoint metadata JSON without provider secrets" eg:"{}"`
}

// UpdateProviderEndpointRes defines the response for updating a provider endpoint.
type UpdateProviderEndpointRes struct{}

// DeleteProviderEndpointReq defines the request for deleting a provider endpoint.
type DeleteProviderEndpointReq struct {
	g.Meta     `path:"/ai/providers/{providerId}/endpoints/{id}" method:"delete" tags:"Smart Center/Channel Management" summary:"Delete provider endpoint" dc:"Delete one provider endpoint after verifying no model references it." permission:"ai:provider:delete"`
	ProviderId int64 `json:"providerId" v:"required|min:1" dc:"Provider ID" eg:"1"`
	Id         int64 `json:"id" v:"required|min:1" dc:"Provider endpoint ID" eg:"1"`
}

// DeleteProviderEndpointRes defines the response for deleting a provider endpoint.
type DeleteProviderEndpointRes struct{}

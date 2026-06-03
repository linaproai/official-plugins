// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// ProviderEndpoint is the golang structure for table provider_endpoint.
type ProviderEndpoint struct {
	Id           int64      `json:"id"           orm:"id"            description:"Provider endpoint ID"`
	ProviderId   int64      `json:"providerId"   orm:"provider_id"   description:"Provider ID"`
	Protocol     string     `json:"protocol"     orm:"protocol"      description:"Protocol: openai, anthropic, voyage, openai-compatible, or provider-specific"`
	BaseUrl      string     `json:"baseUrl"      orm:"base_url"      description:"Provider protocol base URL"`
	SecretRef    string     `json:"secretRef"    orm:"secret_ref"    description:"Secret reference or masked secret reference"`
	Enabled      int        `json:"enabled"      orm:"enabled"       description:"Enabled flag: 0=disabled 1=enabled"`
	MetadataJson string     `json:"metadataJson" orm:"metadata_json" description:"Endpoint metadata JSON without secret values"`
	CreatedAt    *time.Time `json:"createdAt"    orm:"created_at"    description:"Creation time"`
	UpdatedAt    *time.Time `json:"updatedAt"    orm:"updated_at"    description:"Update time"`
	DeletedAt    *time.Time `json:"deletedAt"    orm:"deleted_at"    description:"Deletion time"`
}

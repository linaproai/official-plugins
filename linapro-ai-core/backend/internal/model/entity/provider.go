// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Provider is the golang structure for table provider.
type Provider struct {
	Id               int64      `json:"id"               orm:"id"                 description:"Provider ID"`
	Name             string     `json:"name"             orm:"name"               description:"Provider display name"`
	WebsiteUrl       string     `json:"websiteUrl"       orm:"website_url"        description:"Provider website URL"`
	Remark           string     `json:"remark"           orm:"remark"             description:"Provider remark"`
	OpenaiBaseUrl    string     `json:"openaiBaseUrl"    orm:"openai_base_url"    description:"OpenAI-compatible base URL"`
	AnthropicBaseUrl string     `json:"anthropicBaseUrl" orm:"anthropic_base_url" description:"Anthropic-compatible base URL"`
	ApiKeySecretRef  string     `json:"apiKeySecretRef"  orm:"api_key_secret_ref" description:"API key secret reference or masked secret reference"`
	Enabled          int        `json:"enabled"          orm:"enabled"            description:"Enabled flag: 0=disabled 1=enabled"`
	CreatedAt        *time.Time `json:"createdAt"        orm:"created_at"         description:"Creation time"`
	UpdatedAt        *time.Time `json:"updatedAt"        orm:"updated_at"         description:"Update time"`
	DeletedAt        *time.Time `json:"deletedAt"        orm:"deleted_at"         description:"Deletion time"`
}

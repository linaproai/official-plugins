// This file declares the update-provider request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateReq defines the request for updating an AI provider.
type UpdateReq struct {
	g.Meta           `path:"/ai/providers/{id}" method:"put" tags:"AI Providers" summary:"Update AI provider" dc:"Update one AI provider. Empty API key secret references keep the existing secret reference." permission:"ai:provider:update"`
	Id               int64  `json:"id" v:"required|min:1" dc:"Provider ID" eg:"1"`
	Name             string `json:"name" v:"required|max-length:128" dc:"Provider display name" eg:"OpenAI"`
	WebsiteUrl       string `json:"websiteUrl" v:"max-length:512" dc:"Provider website URL" eg:"https://openai.com"`
	Remark           string `json:"remark" v:"max-length:512" dc:"Provider remark" eg:"Production text models"`
	OpenaiBaseUrl    string `json:"openaiBaseUrl" v:"max-length:512" dc:"OpenAI-compatible base URL" eg:"https://api.openai.com/v1"`
	AnthropicBaseUrl string `json:"anthropicBaseUrl" v:"max-length:512" dc:"Anthropic-compatible base URL" eg:"https://api.anthropic.com/v1"`
	ApiKeySecretRef  string `json:"apiKeySecretRef" v:"max-length:256" dc:"API key secret reference or masked key placeholder" eg:"sk-***abcd"`
	Enabled          int    `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
}

// UpdateRes defines the response for updating an AI provider.
type UpdateRes struct{}

// This file declares the create-provider request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateReq defines the request for creating an AI provider.
type CreateReq struct {
	g.Meta     `path:"/ai/providers" method:"post" tags:"Smart Center/Channel Management" summary:"Create AI provider" dc:"Create one AI provider metadata row and its OpenAI or Anthropic endpoint configuration in one database transaction. Plaintext endpoint secrets are never returned." permission:"ai:provider:create"`
	Name       string                      `json:"name" v:"required|max-length:128" dc:"Provider display name" eg:"OpenAI"`
	WebsiteUrl string                      `json:"websiteUrl" v:"max-length:512" dc:"Provider website URL" eg:"https://openai.com"`
	Remark     string                      `json:"remark" v:"max-length:512" dc:"Provider remark" eg:"Production text models"`
	Enabled    int                         `json:"enabled" d:"1" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	Endpoints  []*ProviderEndpointSaveItem `json:"endpoints" dc:"OpenAI and Anthropic endpoint configuration saved with the provider" eg:"[]"`
}

// CreateRes defines the response for creating an AI provider.
type CreateRes struct {
	Id int64 `json:"id" dc:"Created provider ID" eg:"1"`
}

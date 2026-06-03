// This file declares the create-provider request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateReq defines the request for creating an AI provider.
type CreateReq struct {
	g.Meta     `path:"/ai/providers" method:"post" tags:"AI Providers" summary:"Create AI provider" dc:"Create one AI provider metadata row. Protocol endpoints and secrets are managed by the provider endpoint APIs." permission:"ai:provider:create"`
	Name       string `json:"name" v:"required|max-length:128" dc:"Provider display name" eg:"OpenAI"`
	WebsiteUrl string `json:"websiteUrl" v:"max-length:512" dc:"Provider website URL" eg:"https://openai.com"`
	Remark     string `json:"remark" v:"max-length:512" dc:"Provider remark" eg:"Production text models"`
	Enabled    int    `json:"enabled" d:"1" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
}

// CreateRes defines the response for creating an AI provider.
type CreateRes struct {
	Id int64 `json:"id" dc:"Created provider ID" eg:"1"`
}

// This file declares the update-provider request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateReq defines the request for updating an AI provider.
type UpdateReq struct {
	g.Meta     `path:"/ai/providers/{id}" method:"put" tags:"AI Providers" summary:"Update AI provider" dc:"Update one AI provider metadata row. Protocol endpoints and secrets are managed by the provider endpoint APIs." permission:"ai:provider:update"`
	Id         int64  `json:"id" v:"required|min:1" dc:"Provider ID" eg:"1"`
	Name       string `json:"name" v:"required|max-length:128" dc:"Provider display name" eg:"OpenAI"`
	WebsiteUrl string `json:"websiteUrl" v:"max-length:512" dc:"Provider website URL" eg:"https://openai.com"`
	Remark     string `json:"remark" v:"max-length:512" dc:"Provider remark" eg:"Production text models"`
	Enabled    int    `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
}

// UpdateRes defines the response for updating an AI provider.
type UpdateRes struct{}

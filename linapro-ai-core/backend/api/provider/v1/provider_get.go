// This file declares the get-provider request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetReq defines the request for reading an AI provider.
type GetReq struct {
	g.Meta `path:"/ai/providers/{id}" method:"get" tags:"Smart Center/Channel Management" summary:"Get AI provider" dc:"Get one AI provider detail and aggregated model counts by provider ID." permission:"ai:provider:list"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Provider ID" eg:"1"`
}

// GetRes defines the response for reading an AI provider.
type GetRes struct {
	ProviderItem
}

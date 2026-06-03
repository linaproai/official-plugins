// This file declares model update and delete DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateReq defines the request for updating an AI model.
type UpdateReq struct {
	g.Meta     `path:"/ai/models/{id}" method:"put" tags:"AI Models" summary:"Update AI model" dc:"Update one AI model identity after verifying the default endpoint still belongs to the provider and protocol. Capability method declarations are maintained through the model capability API." permission:"ai:provider:update"`
	Id         int64  `json:"id" v:"required|min:1" dc:"Model ID" eg:"1"`
	EndpointId int64  `json:"endpointId" v:"required|min:1" dc:"Default provider endpoint ID used when a capability declaration has no endpoint override" eg:"1"`
	ModelName  string `json:"modelName" v:"required|max-length:128" dc:"Provider model name" eg:"gpt-4.1-mini"`
	Protocol   string `json:"protocol" v:"required|in:openai,anthropic,voyage,openai-compatible,anthropic-compatible" dc:"Provider protocol: openai, anthropic, voyage, openai-compatible, or anthropic-compatible" eg:"openai-compatible"`
	Enabled    int    `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
}

// UpdateRes defines the response for updating an AI model.
type UpdateRes struct{}

// DeleteReq defines the request for deleting an AI model.
type DeleteReq struct {
	g.Meta `path:"/ai/models/{id}" method:"delete" tags:"AI Models" summary:"Delete AI model" dc:"Delete one AI model after verifying no AI capability tier binding references it." permission:"ai:provider:delete"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Model ID" eg:"1"`
}

// DeleteRes defines the response for deleting an AI model.
type DeleteRes struct{}

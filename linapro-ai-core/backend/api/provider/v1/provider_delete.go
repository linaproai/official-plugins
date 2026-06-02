// This file declares the delete-provider request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteReq defines the request for deleting an AI provider.
type DeleteReq struct {
	g.Meta `path:"/ai/providers/{id}" method:"delete" tags:"AI Providers" summary:"Delete AI provider" dc:"Delete one AI provider after verifying no AI capability tier references any of its models." permission:"ai:provider:delete"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Provider ID" eg:"1"`
}

// DeleteRes defines the response for deleting an AI provider.
type DeleteRes struct{}

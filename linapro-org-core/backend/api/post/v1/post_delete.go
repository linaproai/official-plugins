package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteReq defines the request for deleting posts.
type DeleteReq struct {
	g.Meta `path:"/post" method:"delete" tags:"Position Management" summary:"Delete position" dc:"Delete one or more positions by query array ids[]. If the position has been assigned to the user, deletion is not allowed." permission:"system:post:remove"`
	Ids    []int `json:"ids" v:"required|min-length:1" dc:"Position ID list as a query array, e.g. ids[]=1&ids[]=2&ids[]=3" eg:"[1,2,3]"`
}

// DeleteRes defines the response for deleting posts.
type DeleteRes struct{}

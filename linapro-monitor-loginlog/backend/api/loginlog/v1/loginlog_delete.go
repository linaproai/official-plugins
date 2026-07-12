package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog Delete API

// DeleteReq defines the request for deleting login logs.
type DeleteReq struct {
	g.Meta `path:"/loginlog" method:"delete" tags:"Login Logs" summary:"Delete login logs" dc:"Delete one or more login log records by query array ids[]" permission:"monitor:loginlog:remove"`
	Ids    []int `json:"ids" v:"required|min-length:1" dc:"Log ID list as a query array, e.g. ids[]=1&ids[]=2&ids[]=3" eg:"[1,2,3]"`
}

// DeleteRes is the login-log delete response.
type DeleteRes struct {
	Deleted int `json:"deleted" dc:"Number of records actually deleted" eg:"3"`
}

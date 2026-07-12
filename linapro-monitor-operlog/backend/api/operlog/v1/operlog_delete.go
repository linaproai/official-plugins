package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Delete API

// DeleteReq defines the request for deleting operation logs.
type DeleteReq struct {
	g.Meta `path:"/operlog" method:"delete" tags:"Operation Logs" summary:"Delete operation log" dc:"Delete one or more operation log records by query array ids[]" permission:"monitor:operlog:remove"`
	Ids    []int `json:"ids" v:"required|min-length:1" dc:"Log ID list as a query array, e.g. ids[]=1&ids[]=2&ids[]=3" eg:"[1,2,3]"`
}

// DeleteRes is the operation-log delete response.
type DeleteRes struct {
	Deleted int `json:"deleted" dc:"Number of records actually deleted" eg:"3"`
}

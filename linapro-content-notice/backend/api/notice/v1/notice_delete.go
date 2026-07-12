// This file declares the delete-notice request/response DTOs used by the
// linapro-content-notice source plugin.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Notice Delete API

// DeleteReq defines the request for deleting notices.
type DeleteReq struct {
	g.Meta `path:"/notice" method:"delete" tags:"Notices" summary:"Delete notification or announcement" dc:"Delete one or more notifications or announcements by query array ids[]" permission:"system:notice:remove"`
	Ids    []int64 `json:"ids" v:"required|min-length:1" dc:"Announcement ID list as a query array, e.g. ids[]=1&ids[]=2&ids[]=3" eg:"[1,2,3]"`
}

// DeleteRes Notice delete response
type DeleteRes struct{}

// This file defines shared post response DTOs for the org-center API.
package v1

import "github.com/gogf/gf/v2/os/gtime"

// PostItem exposes position fields visible through org-center APIs.
type PostItem struct {
	Id        int         `json:"id" dc:"Position ID" eg:"1"`
	DeptId    int         `json:"deptId" dc:"Department ID" eg:"100"`
	Code      string      `json:"code" dc:"Position code" eg:"dev"`
	Name      string      `json:"name" dc:"Position name" eg:"Development Engineer"`
	Sort      int         `json:"sort" dc:"Sort order" eg:"1"`
	Status    int         `json:"status" dc:"Status: 1=normal 0=disabled" eg:"1"`
	Remark    string      `json:"remark" dc:"Remark" eg:"Responsible for system development"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"Creation time" eg:"2026-04-21 10:00:00"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"Update time" eg:"2026-04-21 10:30:00"`
}

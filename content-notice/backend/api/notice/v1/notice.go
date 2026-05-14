// This file defines shared notice response DTOs for the content-notice API.
package v1

import "github.com/gogf/gf/v2/os/gtime"

// NoticeItem exposes notification and announcement fields visible to callers.
type NoticeItem struct {
	Id        int64       `json:"id" dc:"Announcement ID" eg:"1"`
	Title     string      `json:"title" dc:"Announcement title" eg:"System maintenance notification"`
	Type      int         `json:"type" dc:"Announcement type: 1=Notice 2=Announcement" eg:"1"`
	Content   string      `json:"content" dc:"Announcement content, supports rich text HTML" eg:"<p>The system will be undergoing maintenance and upgrade tonight</p>"`
	FileIds   string      `json:"fileIds" dc:"Attachment file ID list, comma-separated" eg:"1,2,3"`
	Status    int         `json:"status" dc:"Announcement status: 0=Draft 1=Published" eg:"1"`
	Remark    string      `json:"remark" dc:"Remark" eg:"Emergency notification"`
	CreatedBy int64       `json:"createdBy" dc:"Creator user ID" eg:"1"`
	UpdatedBy int64       `json:"updatedBy" dc:"Last updated user ID" eg:"1"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"Creation time" eg:"2026-04-21 10:00:00"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"Last updated time" eg:"2026-04-21 10:30:00"`
}

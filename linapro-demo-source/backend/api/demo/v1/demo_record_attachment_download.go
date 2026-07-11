// demo_record_attachment_download.go defines the request DTO for downloading
// one linapro-demo-source attachment.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DownloadAttachmentReq is the request for downloading one linapro-demo-source attachment.
type DownloadAttachmentReq struct {
	g.Meta `path:"/records/{id}/attachment" method:"get" tags:"Source Plugin Demo" summary:"Download the source plugin sample attachment" dc:"Download the attachment currently associated with a linapro-demo-source sample record, demonstrating reads from plugin-owned storage files." permission:"linapro-demo-source:example:view"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Record ID" eg:"1"`
}

// DownloadAttachmentRes is the response placeholder for attachment downloads streamed by the controller.
type DownloadAttachmentRes struct{}

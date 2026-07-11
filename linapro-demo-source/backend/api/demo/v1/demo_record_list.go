// demo_record_list.go defines the request and response DTOs for querying
// linapro-demo-source records.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListRecordsReq is the request for querying linapro-demo-source records.
type ListRecordsReq struct {
	g.Meta   `path:"/records" method:"get" tags:"Source Plugin Demo" summary:"List source plugin sample records" dc:"List business records for the linapro-demo-source sample page, with optional fuzzy title filtering. This demonstrates CRUD operations against a table created by the source plugin installation SQL." permission:"linapro-demo-source:example:view"`
	PageNum  int    `json:"pageNum" dc:"Page number, defaults to page 1 when omitted" eg:"1"`
	PageSize int    `json:"pageSize" dc:"Number of items per page; defaults to 10 when omitted" eg:"10"`
	Keyword  string `json:"keyword" dc:"Fuzzy filter by Record title; queries all records when omitted" eg:"Example"`
}

// ListRecordsRes is the response for querying linapro-demo-source records.
type ListRecordsRes struct {
	List  []*RecordItem `json:"list" dc:"Records on the current page" eg:"[]"`
	Total int           `json:"total" dc:"Total number of matching records" eg:"1"`
}

// RecordItem defines one linapro-demo-source record row.
type RecordItem struct {
	Id             int64  `json:"id" dc:"Record ID" eg:"1"`
	Title          string `json:"title" dc:"Record title" eg:"Source plugin SQL sample record"`
	Content        string `json:"content" dc:"Summary of record content" eg:"This record is used to demonstrate how the source plugin page operates the data table created by installing SQL."`
	AttachmentName string `json:"attachmentName" dc:"The original file name of the attachment. If there is no attachment, an empty string is returned." eg:"linapro-demo-source-note.txt"`
	HasAttachment  int    `json:"hasAttachment" dc:"Whether the attachment exists: 1=exists 0=does not exist" eg:"1"`
	CreatedAt      *int64 `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1776333600000"`
	UpdatedAt      *int64 `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1776333900000"`
}

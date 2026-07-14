// Package v1 declares Account management API DTOs for linapro-mail-core.
package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq lists mail accounts.
type ListReq struct {
	g.Meta   `path:"/mail/accounts" method:"get" tags:"Mail Account" summary:"List mail accounts" permission:"linapro-mail-core:account:list"`
	PageNum  int    `json:"pageNum" d:"1" dc:"Page number starting from 1; omit to use default 1"`
	PageSize int    `json:"pageSize" d:"20" dc:"Number of items per page; omit to use default 20"`
	Name     string `json:"name" dc:"Optional name fuzzy filter; empty means no name filter"`
	Status   *int   `json:"status" dc:"Optional status filter: 1=enabled 0=disabled; omit to return all statuses"`
}

// ListRes is the account list response.
type ListRes struct {
	g.Meta `mime:"application/json"`
	List   []*AccountItem `json:"list" dc:"Account list items for the current page"`
	Total  int            `json:"total" dc:"Total matching account count before pagination"`
}

// GetReq loads one account.
type GetReq struct {
	g.Meta `path:"/mail/accounts/{id}" method:"get" tags:"Mail Account" summary:"Get mail account" permission:"linapro-mail-core:account:query"`
	Id     int64 `json:"id" in:"path" v:"required|min:1" dc:"Account primary key ID"`
}

// GetRes is the account detail response.
type GetRes struct {
	g.Meta `mime:"application/json"`
	*AccountItem
}

// CreateReq creates one account.
type CreateReq struct {
	g.Meta               `path:"/mail/accounts" method:"post" tags:"Mail Account" summary:"Create mail account" permission:"linapro-mail-core:account:add"`
	Name                 string `json:"name" v:"required" dc:"Display name of the mail account"`
	FromAddress          string `json:"fromAddress" dc:"Default From address used when sending mail"`
	OutboundConnectionId int64  `json:"outboundConnectionId" dc:"Outbound (SMTP) connection ID; 0 means outbound is not bound"`
	InboundConnectionId  int64  `json:"inboundConnectionId" dc:"Inbound (IMAP/POP3) connection ID; 0 means outbound-only account"`
	IsDefault            bool   `json:"isDefault" dc:"Whether this account is the platform default mail account"`
	Status               int    `json:"status" d:"1" dc:"Account status: 1=enabled 0=disabled"`
	Remark               string `json:"remark" dc:"Optional free-form remark"`
}

// CreateRes returns the created account ID.
type CreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int64 `json:"id" dc:"Created account primary key ID"`
}

// UpdateReq updates one account.
type UpdateReq struct {
	g.Meta               `path:"/mail/accounts/{id}" method:"put" tags:"Mail Account" summary:"Update mail account" permission:"linapro-mail-core:account:edit"`
	Id                   int64  `json:"id" in:"path" v:"required|min:1" dc:"Account primary key ID"`
	Name                 string `json:"name" v:"required" dc:"Display name of the mail account"`
	FromAddress          string `json:"fromAddress" dc:"Default From address used when sending mail"`
	OutboundConnectionId int64  `json:"outboundConnectionId" dc:"Outbound (SMTP) connection ID; 0 means outbound is not bound"`
	InboundConnectionId  int64  `json:"inboundConnectionId" dc:"Inbound (IMAP/POP3) connection ID; 0 means outbound-only account"`
	IsDefault            bool   `json:"isDefault" dc:"Whether this account is the platform default mail account"`
	Status               int    `json:"status" d:"1" dc:"Account status: 1=enabled 0=disabled"`
	Remark               string `json:"remark" dc:"Optional free-form remark"`
}

// UpdateRes is empty success body.
type UpdateRes struct {
	g.Meta `mime:"application/json"`
}

// DeleteReq deletes accounts.
type DeleteReq struct {
	g.Meta `path:"/mail/accounts" method:"delete" tags:"Mail Account" summary:"Delete mail accounts" permission:"linapro-mail-core:account:remove"`
	Ids    []int64 `json:"ids" v:"required" dc:"Account primary key IDs to delete"`
}

// DeleteRes is empty success body.
type DeleteRes struct {
	g.Meta `mime:"application/json"`
}

// AccountItem is one account projection.
type AccountItem struct {
	Id                   int64  `json:"id" dc:"Account primary key ID"`
	Name                 string `json:"name" dc:"Display name of the mail account"`
	FromAddress          string `json:"fromAddress" dc:"Default From address used when sending mail"`
	OutboundConnectionId int64  `json:"outboundConnectionId" dc:"Outbound (SMTP) connection ID; 0 means outbound is not bound"`
	InboundConnectionId  int64  `json:"inboundConnectionId" dc:"Inbound (IMAP/POP3) connection ID; 0 means outbound-only account"`
	IsDefault            bool   `json:"isDefault" dc:"Whether this account is the platform default mail account"`
	Status               int    `json:"status" dc:"Account status: 1=enabled 0=disabled"`
	Remark               string `json:"remark" dc:"Optional free-form remark"`
	CreatedAt            int64  `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds"`
	UpdatedAt            int64  `json:"updatedAt" dc:"Last update time as Unix timestamp in milliseconds"`
}

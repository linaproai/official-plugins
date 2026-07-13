// Package v1 declares Connection management API DTOs for linapro-mail-core.
package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq lists mail connections.
type ListReq struct {
	g.Meta   `path:"/mail/connections" method:"get" tags:"Mail Connection" summary:"List mail connections" permission:"linapro-mail-core:connection:list"`
	PageNum  int    `json:"pageNum" d:"1" dc:"Page number" eg:"1"`
	PageSize int    `json:"pageSize" d:"20" dc:"Page size" eg:"20"`
	Name     string `json:"name" dc:"Name fuzzy filter"`
	Kind     string `json:"kind" dc:"Transport kind filter: smtp, imap, pop3"`
	Status   *int   `json:"status" dc:"Status filter: 1=enabled 0=disabled"`
}

// ListRes is the connection list response.
type ListRes struct {
	g.Meta `mime:"application/json"`
	List   []*ConnectionItem `json:"list"`
	Total  int               `json:"total"`
}

// GetReq loads one connection.
type GetReq struct {
	g.Meta `path:"/mail/connections/{id}" method:"get" tags:"Mail Connection" summary:"Get mail connection" permission:"linapro-mail-core:connection:query"`
	Id     int64 `json:"id" in:"path" v:"required|min:1" dc:"Connection ID"`
}

// GetRes is the connection detail response.
type GetRes struct {
	g.Meta `mime:"application/json"`
	*ConnectionItem
}

// CreateReq creates one connection.
type CreateReq struct {
	g.Meta    `path:"/mail/connections" method:"post" tags:"Mail Connection" summary:"Create mail connection" permission:"linapro-mail-core:connection:add"`
	Name      string `json:"name" v:"required" dc:"Display name"`
	Kind      string `json:"kind" v:"required" dc:"Transport kind: smtp, imap, pop3"`
	Host      string `json:"host" v:"required" dc:"Server host"`
	Port      int    `json:"port" v:"required|min:1|max:65535" dc:"Server port"`
	Username  string `json:"username" dc:"Auth username"`
	SecretRef string `json:"secretRef" dc:"Secret reference or password material reference"`
	TlsMode   string `json:"tlsMode" dc:"TLS mode: disable, starttls, tls"`
	AuthMode  string `json:"authMode" dc:"Auth mode: password"`
	ExtraJson string `json:"extraJson" dc:"Extension JSON"`
	Status    int    `json:"status" d:"1" dc:"Status: 1=enabled 0=disabled"`
	Remark    string `json:"remark" dc:"Remark"`
}

// CreateRes returns the created connection ID.
type CreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int64 `json:"id"`
}

// UpdateReq updates one connection.
type UpdateReq struct {
	g.Meta    `path:"/mail/connections/{id}" method:"put" tags:"Mail Connection" summary:"Update mail connection" permission:"linapro-mail-core:connection:edit"`
	Id        int64  `json:"id" in:"path" v:"required|min:1" dc:"Connection ID"`
	Name      string `json:"name" v:"required" dc:"Display name"`
	Kind      string `json:"kind" v:"required" dc:"Transport kind"`
	Host      string `json:"host" v:"required" dc:"Server host"`
	Port      int    `json:"port" v:"required|min:1|max:65535" dc:"Server port"`
	Username  string `json:"username" dc:"Auth username"`
	SecretRef string `json:"secretRef" dc:"Secret reference"`
	TlsMode   string `json:"tlsMode" dc:"TLS mode"`
	AuthMode  string `json:"authMode" dc:"Auth mode"`
	ExtraJson string `json:"extraJson" dc:"Extension JSON"`
	Status    int    `json:"status" d:"1" dc:"Status"`
	Remark    string `json:"remark" dc:"Remark"`
}

// UpdateRes is empty success body.
type UpdateRes struct {
	g.Meta `mime:"application/json"`
}

// DeleteReq deletes connections.
type DeleteReq struct {
	g.Meta `path:"/mail/connections" method:"delete" tags:"Mail Connection" summary:"Delete mail connections" permission:"linapro-mail-core:connection:remove"`
	Ids    []int64 `json:"ids" v:"required" dc:"Connection IDs"`
}

// DeleteRes is empty success body.
type DeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ProbeReq probes one connection.
type ProbeReq struct {
	g.Meta `path:"/mail/connections/{id}/probe" method:"post" tags:"Mail Connection" summary:"Probe mail connection" permission:"linapro-mail-core:connection:probe"`
	Id     int64 `json:"id" in:"path" v:"required|min:1" dc:"Connection ID"`
}

// ProbeRes is empty success body.
type ProbeRes struct {
	g.Meta `mime:"application/json"`
}

// ConnectionItem is one connection projection without secret plaintext expansion.
type ConnectionItem struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	SecretRef string `json:"secretRef"`
	TlsMode   string `json:"tlsMode"`
	AuthMode  string `json:"authMode"`
	ExtraJson string `json:"extraJson"`
	Status    int    `json:"status"`
	Remark    string `json:"remark"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

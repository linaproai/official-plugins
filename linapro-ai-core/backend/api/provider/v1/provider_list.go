// This file declares the list-provider request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq defines the request for paged AI provider listing.
type ListReq struct {
	g.Meta   `path:"/ai/providers" method:"get" tags:"Smart Center/Channel Management" summary:"List AI providers" dc:"Query AI providers by page with keyword and enabled-state filters. Each row includes batched model summaries and masked secrets to avoid per-row follow-up requests." permission:"ai:provider:list"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	Keyword  string `json:"keyword" dc:"Provider name keyword for fuzzy search" eg:"openai"`
	Enabled  *int   `json:"enabled" dc:"Optional enabled filter: 0=disabled 1=enabled" eg:"1"`
}

// ListRes defines the response for paged AI provider listing.
type ListRes struct {
	List  []*ProviderItem `json:"list" dc:"AI provider list" eg:"[]"`
	Total int             `json:"total" dc:"Total number of providers matching filters" eg:"20"`
}

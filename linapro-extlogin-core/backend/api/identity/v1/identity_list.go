// identity_list.go defines list DTOs for the current user's external identities.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListIdentitiesReq lists external identities bound to the current session user.
type ListIdentitiesReq struct {
	g.Meta `path:"/identities" method:"get" tags:"Auth Login / External Identity" summary:"List bound external identities" dc:"Returns external identities bound to the current session user only."`
}

// IdentityItem is one bound external identity projection.
type IdentityItem struct {
	Provider    string `json:"provider" dc:"Stable external provider ID" eg:"google"`
	Subject     string `json:"subject" dc:"Immutable provider subject" eg:"110169484474386276334"`
	SubjectKind string `json:"subjectKind" dc:"Subject classification" eg:"oidc_sub"`
	AppContext  string `json:"appContext" dc:"Optional multi-app context" eg:""`
	Email       string `json:"email" dc:"Email snapshot" eg:"user@example.com"`
	Phone       string `json:"phone" dc:"Phone snapshot" eg:""`
	DisplayName string `json:"displayName" dc:"Display name snapshot" eg:"Ada"`
	AvatarURL   string `json:"avatarUrl" dc:"Avatar URL snapshot" eg:"https://example.com/a.png"`
}

// ListIdentitiesRes is the list response.
type ListIdentitiesRes struct {
	List []IdentityItem `json:"list" dc:"Bound external identities"`
}

// identity_list.go defines the request and response DTOs for listing the
// current session user's bound external identities.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListIdentitiesReq is the request for listing the current session user's bound
// external identities.
type ListIdentitiesReq struct {
	g.Meta `path:"/plugins/linapro-oidc-core/identities" method:"get" tags:"External Identity" summary:"List bound external identities" dc:"Return the external identities bound to the current session user. The result is strictly self-isolated: it never exposes other accounts' linkages, so no extra permission node is required beyond authentication."`
}

// ListIdentitiesRes is the response for listing the current session user's
// bound external identities.
type ListIdentitiesRes struct {
	// Items contains the current user's bound external identities.
	Items []BoundIdentityItem `json:"items" dc:"Bound external identities of the current session user"`
}

// BoundIdentityItem is one bound external identity projection.
type BoundIdentityItem struct {
	Provider string `json:"provider" dc:"Stable external provider ID, e.g. google or discord" eg:"google"`
	Subject  string `json:"subject" dc:"Immutable provider-issued subject identifier" eg:"110169484474386276334"`
	Email    string `json:"email" dc:"Email snapshot captured at link time for display only; may be empty for email-less providers" eg:"user@example.com"`
}

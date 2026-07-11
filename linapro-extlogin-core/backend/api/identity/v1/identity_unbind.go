// identity_unbind.go defines unbind DTOs for the current session user.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UnbindIdentityReq unbinds one external identity from the current session user.
type UnbindIdentityReq struct {
	g.Meta   `path:"/identities" method:"delete" tags:"Auth Login / External Identity" summary:"Unbind an external identity" dc:"Removes one external identity linkage from the current session user. Cross-user targets report not found without leaking existence."`
	Provider string `json:"provider" v:"required|length:1,64" dc:"Stable external provider ID" eg:"google"`
	Subject  string `json:"subject" v:"required|length:1,191" dc:"Immutable provider-issued subject" eg:"110169484474386276334"`
}

// UnbindIdentityRes is the unbind response.
type UnbindIdentityRes struct{}

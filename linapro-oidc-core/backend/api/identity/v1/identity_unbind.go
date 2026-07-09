// identity_unbind.go defines the request and response DTOs for unbinding one of
// the current session user's external identities.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UnbindIdentityReq is the request for unbinding one of the current session
// user's external identities.
type UnbindIdentityReq struct {
	g.Meta   `path:"/plugins/linapro-oidc-core/identities" method:"delete" tags:"External Identity" summary:"Unbind an external identity" dc:"Remove one external identity linkage from the current session user. Unbinding only acts on the current session user's own linkages; a linkage that does not belong to the current user is reported as not found without leaking other accounts' linkage existence. Unbinding frees the (provider, subject) key for a future relink."`
	Provider string `json:"provider" v:"required|length:1,64" dc:"Stable external provider ID of the linkage to remove" eg:"google"`
	Subject  string `json:"subject" v:"required|length:1,191" dc:"Immutable provider-issued subject identifier of the linkage to remove" eg:"110169484474386276334"`
}

// UnbindIdentityRes is the response for unbinding one external identity.
type UnbindIdentityRes struct{}

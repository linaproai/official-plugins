// identity_bind.go defines the request and response DTOs for binding one
// verified external identity to the current session user.

package v1

import "github.com/gogf/gf/v2/frame/g"

// BindIdentityReq is the request for binding one verified external identity to
// the current session user.
type BindIdentityReq struct {
	g.Meta   `path:"/plugins/linapro-oidc-core/identities" method:"post" tags:"External Identity" summary:"Bind an external identity" dc:"Link one verified external identity (provider + subject) to the current session user. The caller must have already verified the identity through the owning OAuth plugin; this endpoint performs no OAuth exchange. Binding only acts on the current session user, and a (provider, subject) pair already owned by another account is rejected with a conflict error. Re-binding an identity the current user already owns succeeds idempotently."`
	Provider string `json:"provider" v:"required|length:1,64" dc:"Stable external provider ID owned by the verifying OAuth plugin" eg:"google"`
	Subject  string `json:"subject" v:"required|length:1,191" dc:"Immutable provider-issued subject identifier obtained from the verified identity" eg:"110169484474386276334"`
	Email    string `json:"email" v:"length:0,191" dc:"Optional verified email captured as an audit snapshot only; never used as a resolution key" eg:"user@example.com"`
}

// BindIdentityRes is the response for binding one external identity.
type BindIdentityRes struct{}

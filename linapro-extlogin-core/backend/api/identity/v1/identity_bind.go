// identity_bind.go defines the request and response DTOs for binding one
// verified external identity ticket to the current session user.

package v1

import "github.com/gogf/gf/v2/frame/g"

// BindIdentityReq binds one verified-identity ticket to the current session user.
// Clients MUST NOT submit bare provider/subject pairs; tickets are issued only
// after a protocol plugin completes IdP verification.
type BindIdentityReq struct {
	g.Meta   `path:"/identities/bind" method:"post" tags:"External Identity" summary:"Bind a verified external identity ticket" dc:"Consumes a single-use verified-identity ticket issued after OAuth/OIDC verification and links that identity to the current session user. Bare provider and subject values are not accepted. A (provider, subject) pair already owned by another account is rejected with a conflict error. Re-binding an identity the current user already owns succeeds idempotently."`
	TicketID string `json:"ticketId" v:"required|length:1,128" dc:"Single-use verified-identity ticket issued by linapro-extlogin-core after protocol verification" eg:"extid_ab12..."`
}

// BindIdentityRes is the response for binding one external identity.
type BindIdentityRes struct{}

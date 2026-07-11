// handoff_exchange.go defines the public SPA exchange API for one-time
// external-login handoff codes owned by linapro-extlogin-core.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ExchangeLoginHandoffReq consumes a single-use login handoff code.
type ExchangeLoginHandoffReq struct {
	g.Meta  `path:"/handoff/exchange" method:"post" tags:"Auth Login / External Identity" summary:"Exchange external login handoff" dc:"Consumes a single-use handoff code created after a third-party identity login. Returns a token pair or a pre-login token plus tenant candidates. Invalid, expired, or already-used codes are rejected. This endpoint is public (no session required) and is the SPA delivery surface for external login; JWTs must never appear in OAuth redirect URLs."`
	Handoff string `json:"handoff" v:"required|length:1,128" dc:"Single-use handoff code from the protocol plugin SPA redirect" eg:"elh_8f4f..."`
}

// HandoffTenantEntity is one tenant candidate returned during two-stage login.
type HandoffTenantEntity struct {
	Id     int    `json:"id" dc:"Tenant ID" eg:"1"`
	Code   string `json:"code" dc:"Tenant code" eg:"acme"`
	Name   string `json:"name" dc:"Tenant display name" eg:"Acme"`
	Status string `json:"status" dc:"Tenant status" eg:"active"`
}

// ExchangeLoginHandoffRes mirrors the host password-login response shape.
type ExchangeLoginHandoffRes struct {
	AccessToken  string                 `json:"accessToken" dc:"JWT access token when tenant selection is not required"`
	RefreshToken string                 `json:"refreshToken" dc:"JWT refresh token when tenant selection is not required"`
	PreToken     string                 `json:"preToken" dc:"Short-lived pre-login token when tenant selection is required"`
	Tenants      []*HandoffTenantEntity `json:"tenants" dc:"Tenant candidates when tenant selection is required"`
}

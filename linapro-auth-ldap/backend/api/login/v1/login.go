// Package v1 declares the public LDAP login API DTOs.
package v1

import "github.com/gogf/gf/v2/frame/g"

// LoginReq is the public directory login request.
// Bound on the portal route; permission is intentionally empty (public).
type LoginReq struct {
	g.Meta   `path:"/login" method:"post" tags:"LDAP Login" summary:"LDAP directory login" dc:"Verify directory credentials and return a one-time handoff code for SPA session exchange. Passwords are never stored."`
	Username string `json:"username" v:"required|length:1,128" dc:"Directory username" eg:"alice"`
	Password string `json:"password" v:"required|length:1,256" dc:"Directory password; never logged or stored" eg:"***"`
}

// LoginRes carries the handoff for SPA exchange.
type LoginRes struct {
	Handoff string `json:"handoff" dc:"One-time handoff code; exchange via linapro-extlogin-core" eg:"handoff-code"`
}

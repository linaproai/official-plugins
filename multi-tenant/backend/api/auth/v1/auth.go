// This file defines shared tenant-auth response DTOs for the multi-tenant API.
package v1

// LoginTenantItem is one tenant candidate returned during login.
type LoginTenantItem struct {
	Id     int64  `json:"id" dc:"Tenant ID" eg:"1"`
	Code   string `json:"code" dc:"Tenant code" eg:"acme"`
	Name   string `json:"name" dc:"Tenant name" eg:"Acme BU"`
	Status string `json:"status" dc:"Tenant status" eg:"active"`
}

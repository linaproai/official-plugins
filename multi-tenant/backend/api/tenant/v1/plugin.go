// This file declares tenant plugin-governance list DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantPluginListReq defines the request for listing tenant-controllable plugins.
type TenantPluginListReq struct {
	g.Meta `path:"/tenant/plugins" method:"get" tags:"Tenant Plugins" summary:"Get tenant plugin list" dc:"List tenant-scoped plugins that the current tenant administrator may enable or disable." permission:"system:tenant:plugin:list"`
}

// TenantPluginListRes defines the tenant plugin list response.
type TenantPluginListRes struct {
	List  []*TenantPluginItem `json:"list" dc:"Tenant plugin list" eg:"[]"`
	Total int                 `json:"total" dc:"Total number of tenant plugins" eg:"1"`
}

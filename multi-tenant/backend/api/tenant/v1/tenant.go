// This file defines shared tenant-scoped response DTOs for the multi-tenant API.
package v1

// TenantPluginItem is the tenant-facing plugin governance projection.
type TenantPluginItem struct {
	Id            string `json:"id" dc:"Plugin unique identifier" eg:"monitor-loginlog"`
	Name          string `json:"name" dc:"Plugin display name" eg:"Login Log"`
	Version       string `json:"version" dc:"Plugin version" eg:"v0.1.0"`
	Type          string `json:"type" dc:"Plugin type: source or dynamic" eg:"source"`
	Description   string `json:"description" dc:"Plugin description" eg:"Tenant login audit"`
	Installed     int    `json:"installed" dc:"Whether the plugin is installed: 1=yes 0=no" eg:"1"`
	Enabled       int    `json:"enabled" dc:"Whether the plugin is globally enabled: 1=yes 0=no" eg:"1"`
	ScopeNature   string `json:"scopeNature" dc:"Plugin scope nature: tenant_aware or platform_only" eg:"tenant_aware"`
	InstallMode   string `json:"installMode" dc:"Plugin install mode: global or tenant_scoped" eg:"tenant_scoped"`
	TenantEnabled int    `json:"tenantEnabled" dc:"Whether the plugin is enabled for the current tenant: 1=yes 0=no" eg:"1"`
}

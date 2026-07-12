// GET /settings for Azure Blob storage provider configuration.
package v1

import "github.com/gogf/gf/v2/frame/g"

// GetSettingsReq queries persisted settings.
type GetSettingsReq struct {
	g.Meta `path:"/settings" method:"get" tags:"Storage / Azure Blob" summary:"Query object storage settings" dc:"Return masked Azure Blob storage settings for the admin page." permission:"linapro-storage-azure:settings:view"`
}

// GetSettingsRes carries the masked settings projection.
type GetSettingsRes struct {
	Settings *SettingsItem `json:"settings" dc:"Masked settings projection"`
}

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetSettingsReq queries persisted settings.
type GetSettingsReq struct {
	g.Meta `path:"/settings" method:"get" tags:"Storage / S3" summary:"Query object storage settings" dc:"Return masked S3 object storage settings for the admin page." permission:"linapro-storage-s3:settings:view"`
}

// GetSettingsRes carries the masked settings projection.
type GetSettingsRes struct {
	Settings *SettingsItem `json:"settings" dc:"Masked settings projection"`
}

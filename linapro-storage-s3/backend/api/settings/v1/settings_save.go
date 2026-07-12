package v1

import "github.com/gogf/gf/v2/frame/g"

// SaveSettingsReq persists settings.
type SaveSettingsReq struct {
	g.Meta          `path:"/settings" method:"put" tags:"Storage / S3" summary:"Save object storage settings" dc:"Persist S3 object storage settings. Empty or masked secret keeps the previous value." permission:"linapro-storage-s3:settings:update"`
	AccessKeyID     string `json:"accessKeyID" v:"max-length:256" dc:"Access key id"`
	SecretAccessKey string `json:"secretAccessKey" v:"max-length:512" dc:"Secret key; empty keeps previous"`
	Region          string `json:"region" v:"max-length:128" dc:"Optional signing region; empty defaults to us-east-1"`
	Bucket          string `json:"bucket" v:"max-length:256" dc:"Target bucket"`
	Endpoint        string `json:"endpoint" v:"max-length:512" dc:"Required S3 API endpoint URL"`
	PathPrefix      string `json:"pathPrefix" v:"max-length:256" dc:"Optional key prefix"`
	ForcePathStyle  bool   `json:"forcePathStyle" dc:"Use path-style addressing"`
}

// SaveSettingsRes returns the fresh masked projection.
type SaveSettingsRes struct {
	Settings *SettingsItem `json:"settings" dc:"Masked settings projection after save"`
}

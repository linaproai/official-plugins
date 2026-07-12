package v1

import "github.com/gogf/gf/v2/frame/g"

// SaveSettingsReq persists AWS S3 settings.
type SaveSettingsReq struct {
	g.Meta          `path:"/settings" method:"put" tags:"Storage / AWS" summary:"Save object storage settings" dc:"Persist AWS S3 settings. Empty or masked secret keeps the previous value." permission:"linapro-storage-aws:settings:update"`
	AccessKeyID     string `json:"accessKeyID" v:"max-length:256" dc:"AWS access key id"`
	SecretAccessKey string `json:"secretAccessKey" v:"max-length:512" dc:"AWS secret access key; empty keeps previous"`
	Region          string `json:"region" v:"max-length:128" dc:"AWS region"`
	Bucket          string `json:"bucket" v:"max-length:256" dc:"Target S3 bucket"`
	PathPrefix      string `json:"pathPrefix" v:"max-length:256" dc:"Optional key prefix"`
}

// SaveSettingsRes returns the fresh masked projection.
type SaveSettingsRes struct {
	Settings *SettingsItem `json:"settings" dc:"Masked settings projection after save"`
}

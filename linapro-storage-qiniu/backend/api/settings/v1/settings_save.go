package v1

import "github.com/gogf/gf/v2/frame/g"

// SaveSettingsReq persists settings.
type SaveSettingsReq struct {
	g.Meta          `path:"/settings" method:"put" tags:"Storage / Qiniu Kodo" summary:"Save object storage settings" dc:"Persist cloud object storage settings. Empty or masked secret keeps the previous value." permission:"linapro-storage-qiniu:settings:update"`
	AccessKeyID     string `json:"accessKeyID" v:"max-length:256" dc:"Cloud access key id"`
	SecretAccessKey string `json:"secretAccessKey" v:"max-length:512" dc:"Cloud secret key; empty keeps previous"`
	Region          string `json:"region" v:"max-length:128" dc:"Optional Kodo region ID; empty auto-detects"`
	Bucket          string `json:"bucket" v:"max-length:256" dc:"Target bucket"`
	Endpoint        string `json:"endpoint" v:"max-length:512" dc:"Optional custom download domain"`
	PathPrefix      string `json:"pathPrefix" v:"max-length:256" dc:"Optional key prefix"`
}

// SaveSettingsRes returns the fresh masked projection.
type SaveSettingsRes struct {
	Settings *SettingsItem `json:"settings" dc:"Masked settings projection after save"`
}

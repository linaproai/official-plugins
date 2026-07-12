// PUT /settings for Azure Blob storage provider configuration.
package v1

import "github.com/gogf/gf/v2/frame/g"

// SaveSettingsReq persists settings.
type SaveSettingsReq struct {
	g.Meta      `path:"/settings" method:"put" tags:"Storage / Azure Blob" summary:"Save object storage settings" dc:"Persist Azure Blob storage settings. Empty or masked account key keeps the previous value." permission:"linapro-storage-azure:settings:update"`
	AccountName string `json:"accountName" v:"max-length:256" dc:"Azure storage account name"`
	AccountKey  string `json:"accountKey" v:"max-length:512" dc:"Azure storage account key; empty keeps previous"`
	Container   string `json:"container" v:"max-length:256" dc:"Target blob container"`
	Endpoint    string `json:"endpoint" v:"max-length:512" dc:"Optional custom service endpoint"`
	PathPrefix  string `json:"pathPrefix" v:"max-length:256" dc:"Optional key prefix"`
}

// SaveSettingsRes returns the fresh masked projection.
type SaveSettingsRes struct {
	Settings *SettingsItem `json:"settings" dc:"Masked settings projection after save"`
}

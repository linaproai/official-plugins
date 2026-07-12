// Package v1 declares settings API DTOs for linapro-storage-azure.
package v1

// SettingsItem is the masked settings projection returned to the admin page.
type SettingsItem struct {
	AccountName          string `json:"accountName" dc:"Azure storage account name"`
	AccountKeyMasked     string `json:"accountKeyMasked" dc:"Masked account key indicator"`
	AccountKeyConfigured bool   `json:"accountKeyConfigured" dc:"Whether an account key is stored"`
	Container            string `json:"container" dc:"Target blob container name"`
	Endpoint             string `json:"endpoint" dc:"Optional custom service endpoint URL"`
	PathPrefix           string `json:"pathPrefix" dc:"Optional object key prefix"`
}

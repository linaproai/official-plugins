// Package v1 declares settings API DTOs for linapro-storage-s3.
package v1

// SettingsItem is the masked settings projection returned to the admin page.
type SettingsItem struct {
	AccessKeyID               string `json:"accessKeyID" dc:"Cloud access key id"`
	SecretAccessKeyMasked     string `json:"secretAccessKeyMasked" dc:"Masked secret indicator"`
	SecretAccessKeyConfigured bool   `json:"secretAccessKeyConfigured" dc:"Whether a secret is stored"`
	Region                    string `json:"region" dc:"Optional signing region; empty uses default us-east-1"`
	Bucket                    string `json:"bucket" dc:"Target bucket name"`
	Endpoint                  string `json:"endpoint" dc:"Required S3 API endpoint"`
	PathPrefix                string `json:"pathPrefix" dc:"Optional object key prefix"`
	ForcePathStyle            bool   `json:"forcePathStyle" dc:"Use path-style addressing"`
}

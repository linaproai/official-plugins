// Package v1 declares settings API DTOs for linapro-storage-aws.
package v1

// SettingsItem is the masked settings projection returned to the admin page.
type SettingsItem struct {
	AccessKeyID               string `json:"accessKeyID" dc:"AWS access key id"`
	SecretAccessKeyMasked     string `json:"secretAccessKeyMasked" dc:"Masked secret indicator"`
	SecretAccessKeyConfigured bool   `json:"secretAccessKeyConfigured" dc:"Whether a secret is stored"`
	Region                    string `json:"region" dc:"AWS region"`
	Bucket                    string `json:"bucket" dc:"Target S3 bucket name"`
	PathPrefix                string `json:"pathPrefix" dc:"Optional object key prefix"`
}

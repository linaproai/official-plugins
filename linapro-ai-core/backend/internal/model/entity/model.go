// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Model is the golang structure for table model.
type Model struct {
	Id         int64      `json:"id"         orm:"id"          description:"Model ID"`
	ProviderId int64      `json:"providerId" orm:"provider_id" description:"Provider ID"`
	EndpointId int64      `json:"endpointId" orm:"endpoint_id" description:"Provider endpoint ID used by the model"`
	ModelName  string     `json:"modelName"  orm:"model_name"  description:"Provider model name"`
	Protocol   string     `json:"protocol"   orm:"protocol"    description:"Protocol: openai or anthropic"`
	Source     string     `json:"source"     orm:"source"      description:"Model source: manual or api"`
	Enabled    int        `json:"enabled"    orm:"enabled"     description:"Enabled flag: 0=disabled 1=enabled"`
	CreatedAt  *time.Time `json:"createdAt"  orm:"created_at"  description:"Creation time"`
	UpdatedAt  *time.Time `json:"updatedAt"  orm:"updated_at"  description:"Update time"`
	DeletedAt  *time.Time `json:"deletedAt"  orm:"deleted_at"  description:"Deletion time"`
}

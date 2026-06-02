// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// TierBinding is the golang structure for table tier_binding.
type TierBinding struct {
	Id         int64      `json:"id"         orm:"id"          description:"Binding ID"`
	TierId     int64      `json:"tierId"     orm:"tier_id"     description:"Tier ID"`
	ProviderId int64      `json:"providerId" orm:"provider_id" description:"Provider ID"`
	ModelId    int64      `json:"modelId"    orm:"model_id"    description:"Model ID"`
	Priority   int        `json:"priority"   orm:"priority"    description:"Binding priority, 0 is primary"`
	Enabled    int        `json:"enabled"    orm:"enabled"     description:"Enabled flag: 0=disabled 1=enabled"`
	CreatedAt  *time.Time `json:"createdAt"  orm:"created_at"  description:"Creation time"`
	UpdatedAt  *time.Time `json:"updatedAt"  orm:"updated_at"  description:"Update time"`
	DeletedAt  *time.Time `json:"deletedAt"  orm:"deleted_at"  description:"Deletion time"`
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// TierBinding is the golang structure of table plugin_linapro_ai_tier_binding for DAO operations like Where/Data.
type TierBinding struct {
	g.Meta     `orm:"table:plugin_linapro_ai_tier_binding, do:true"`
	Id         any        // Binding ID
	TierId     any        // Tier ID
	ProviderId any        // Provider ID
	ModelId    any        // Model ID
	Priority   any        // Binding priority, 0 is primary
	Enabled    any        // Enabled flag: 0=disabled 1=enabled
	CreatedAt  *time.Time // Creation time
	UpdatedAt  *time.Time // Update time
	DeletedAt  *time.Time // Deletion time
}

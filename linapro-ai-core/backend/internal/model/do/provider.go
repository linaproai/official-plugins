// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Provider is the golang structure of table plugin_linapro_ai_provider for DAO operations like Where/Data.
type Provider struct {
	g.Meta     `orm:"table:plugin_linapro_ai_provider, do:true"`
	Id         any        // Provider ID
	Name       any        // Provider display name
	WebsiteUrl any        // Provider website URL
	Remark     any        // Provider remark
	Enabled    any        // Enabled flag: 0=disabled 1=enabled
	CreatedAt  *time.Time // Creation time
	UpdatedAt  *time.Time // Update time
	DeletedAt  *time.Time // Deletion time
}

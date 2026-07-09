// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// UserExternalIdentity is the golang structure of table user_external_identity for DAO operations like Where/Data.
type UserExternalIdentity struct {
	g.Meta        `orm:"table:user_external_identity, do:true"`
	Id            any        // External identity linkage ID
	UserId        any        // Linked local user ID
	Provider      any        // Stable external provider ID owned by the calling plugin, e.g. google, discord
	Subject       any        // Immutable provider-issued subject identifier, e.g. OIDC sub
	PluginId      any        // Calling plugin ID stamped by the host when the linkage was created
	EmailSnapshot any        // Email captured at link time for audit only, never used as a resolution key
	CreatedAt     *time.Time // Creation time
	UpdatedAt     *time.Time // Update time
	DeletedAt     *time.Time // Soft delete time; live rows keep NULL
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Account is the golang structure of table plugin_linapro_mail_core_account for DAO operations like Where/Data.
type Account struct {
	g.Meta               `orm:"table:plugin_linapro_mail_core_account, do:true"`
	Id                   any        // Primary key ID
	Name                 any        // Account display name
	FromAddress          any        // Default From address
	OutboundConnectionId any        // Outbound connection ID; 0 means none
	InboundConnectionId  any        // Inbound connection ID; 0 means none
	IsDefault            any        // Default account flag: 1=default, 0=normal
	Status               any        // Status: 1=enabled, 0=disabled
	TenantId             any        // Tenant ID; 0 means platform scope
	Remark               any        // Remark
	CreatedAt            *time.Time // Creation time
	UpdatedAt            *time.Time // Update time
	DeletedAt            *time.Time // Deletion time
}

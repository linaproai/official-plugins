// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Tier is the golang structure of table plugin_linapro_ai_tier for DAO operations like Where/Data.
type Tier struct {
	g.Meta               `orm:"table:plugin_linapro_ai_tier, do:true"`
	Id                   any        // Tier ID
	CapabilityType       any        // Capability type
	CapabilityMethod     any        // Capability method
	Code                 any        // Tier code: basic, standard, advanced
	DisplayName          any        // Tier display name
	Description          any        // Tier description
	DefaultEffort        any        // Default thinking effort
	Enabled              any        // Enabled flag: 0=disabled 1=enabled
	SortOrder            any        // Stable sort order
	LastTestStatus       any        // Last tier test status
	LastTestLatencyMs    any        // Last tier test latency in milliseconds
	LastTestErrorSummary any        // Last tier test masked error summary
	LastTestAt           *time.Time // Last tier test time
	CreatedAt            *time.Time // Creation time
	UpdatedAt            *time.Time // Update time
}

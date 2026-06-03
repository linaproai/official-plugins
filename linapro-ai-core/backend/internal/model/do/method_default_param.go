// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// MethodDefaultParam is the golang structure of table plugin_linapro_ai_method_default_param for DAO operations like Where/Data.
type MethodDefaultParam struct {
	g.Meta            `orm:"table:plugin_linapro_ai_method_default_param, do:true"`
	Id                any        // Method default param ID
	CapabilityType    any        // Capability type
	CapabilityMethod  any        // Capability method
	DefaultParamsJson any        // Method-specific default params JSON
	Enabled           any        // Enabled flag: 0=disabled 1=enabled
	CreatedAt         *time.Time // Creation time
	UpdatedAt         *time.Time // Update time
}

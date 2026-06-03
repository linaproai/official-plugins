// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// MethodDefaultParam is the golang structure for table method_default_param.
type MethodDefaultParam struct {
	Id                int64      `json:"id"                orm:"id"                  description:"Method default param ID"`
	CapabilityType    string     `json:"capabilityType"    orm:"capability_type"     description:"Capability type"`
	CapabilityMethod  string     `json:"capabilityMethod"  orm:"capability_method"   description:"Capability method"`
	DefaultParamsJson string     `json:"defaultParamsJson" orm:"default_params_json" description:"Method-specific default params JSON"`
	Enabled           int        `json:"enabled"           orm:"enabled"             description:"Enabled flag: 0=disabled 1=enabled"`
	CreatedAt         *time.Time `json:"createdAt"         orm:"created_at"          description:"Creation time"`
	UpdatedAt         *time.Time `json:"updatedAt"         orm:"updated_at"          description:"Update time"`
}

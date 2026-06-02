// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Tier is the golang structure for table tier.
type Tier struct {
	Id                   int64      `json:"id"                   orm:"id"                      description:"Tier ID"`
	CapabilityType       string     `json:"capabilityType"       orm:"capability_type"         description:"Capability type: text"`
	Code                 string     `json:"code"                 orm:"code"                    description:"Tier code: basic, standard, advanced"`
	DisplayName          string     `json:"displayName"          orm:"display_name"            description:"Tier display name"`
	Description          string     `json:"description"          orm:"description"             description:"Tier description"`
	DefaultEffort        string     `json:"defaultEffort"        orm:"default_effort"          description:"Default thinking effort"`
	Enabled              int        `json:"enabled"              orm:"enabled"                 description:"Enabled flag: 0=disabled 1=enabled"`
	SortOrder            int        `json:"sortOrder"            orm:"sort_order"              description:"Stable sort order"`
	LastTestStatus       string     `json:"lastTestStatus"       orm:"last_test_status"        description:"Last tier test status"`
	LastTestLatencyMs    int        `json:"lastTestLatencyMs"    orm:"last_test_latency_ms"    description:"Last tier test latency in milliseconds"`
	LastTestErrorSummary string     `json:"lastTestErrorSummary" orm:"last_test_error_summary" description:"Last tier test masked error summary"`
	LastTestAt           *time.Time `json:"lastTestAt"           orm:"last_test_at"            description:"Last tier test time"`
	CreatedAt            *time.Time `json:"createdAt"            orm:"created_at"              description:"Creation time"`
	UpdatedAt            *time.Time `json:"updatedAt"            orm:"updated_at"              description:"Update time"`
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Account is the golang structure for table account.
type Account struct {
	Id                   int64      `json:"id"                   orm:"id"                     description:"Primary key ID"`
	Name                 string     `json:"name"                 orm:"name"                   description:"Account display name"`
	FromAddress          string     `json:"fromAddress"          orm:"from_address"           description:"Default From address"`
	OutboundConnectionId int64      `json:"outboundConnectionId" orm:"outbound_connection_id" description:"Outbound connection ID; 0 means none"`
	InboundConnectionId  int64      `json:"inboundConnectionId"  orm:"inbound_connection_id"  description:"Inbound connection ID; 0 means none"`
	IsDefault            int        `json:"isDefault"            orm:"is_default"             description:"Default account flag: 1=default, 0=normal"`
	Status               int        `json:"status"               orm:"status"                 description:"Status: 1=enabled, 0=disabled"`
	TenantId             int64      `json:"tenantId"             orm:"tenant_id"              description:"Tenant ID; 0 means platform scope"`
	Remark               string     `json:"remark"               orm:"remark"                 description:"Remark"`
	CreatedAt            *time.Time `json:"createdAt"            orm:"created_at"             description:"Creation time"`
	UpdatedAt            *time.Time `json:"updatedAt"            orm:"updated_at"             description:"Update time"`
	DeletedAt            *time.Time `json:"deletedAt"            orm:"deleted_at"             description:"Deletion time"`
}

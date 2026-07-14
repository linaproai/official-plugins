// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Connection is the golang structure for table connection.
type Connection struct {
	Id        int64      `json:"id"        orm:"id"         description:"Primary key ID"`
	Name      string     `json:"name"      orm:"name"       description:"Connection display name"`
	Kind      string     `json:"kind"      orm:"kind"       description:"Transport kind: smtp, imap, pop3"`
	Host      string     `json:"host"      orm:"host"       description:"Mail server host"`
	Port      int        `json:"port"      orm:"port"       description:"Mail server port"`
	Username  string     `json:"username"  orm:"username"   description:"Authentication username"`
	SecretRef string     `json:"secretRef" orm:"secret_ref" description:"Secret reference for password or token"`
	TlsMode   string     `json:"tlsMode"   orm:"tls_mode"   description:"TLS mode: disable, starttls, tls"`
	AuthMode  string     `json:"authMode"  orm:"auth_mode"  description:"Auth mode: password, oauth2"`
	ExtraJson string     `json:"extraJson" orm:"extra_json" description:"Protocol extension JSON without secrets"`
	Status    int        `json:"status"    orm:"status"     description:"Status: 1=enabled, 0=disabled"`
	TenantId  int64      `json:"tenantId"  orm:"tenant_id"  description:"Tenant ID; 0 means platform scope"`
	Remark    string     `json:"remark"    orm:"remark"     description:"Remark"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}

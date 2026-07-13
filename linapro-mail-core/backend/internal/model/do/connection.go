// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Connection is the golang structure of table plugin_linapro_mail_core_connection for DAO operations like Where/Data.
type Connection struct {
	g.Meta    `orm:"table:plugin_linapro_mail_core_connection, do:true"`
	Id        any        // Primary key ID
	Name      any        // Connection display name
	Kind      any        // Transport kind: smtp, imap, pop3
	Host      any        // Mail server host
	Port      any        // Mail server port
	Username  any        // Authentication username
	SecretRef any        // Secret reference for password or token
	TlsMode   any        // TLS mode: disable, starttls, tls
	AuthMode  any        // Auth mode: password, oauth2
	ExtraJson any        // Protocol extension JSON without secrets
	Status    any        // Status: 1=enabled, 0=disabled
	TenantId  any        // Tenant ID; 0 means platform scope
	Remark    any        // Remark
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
	DeletedAt *time.Time // Deletion time
}

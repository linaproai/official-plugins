// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// UserExternalIdentity is the golang structure for table plugin_linapro_extid_core_user_external_identity.
type UserExternalIdentity struct {
	Id                  int64      `json:"id"                  orm:"id"                    description:"External identity linkage ID"`
	UserId              int        `json:"userId"              orm:"user_id"               description:"Linked local user ID"`
	Provider            string     `json:"provider"            orm:"provider"              description:"Stable external provider ID owned by the calling plugin, e.g. google, discord"`
	Subject             string     `json:"subject"             orm:"subject"               description:"Immutable provider-issued subject identifier, e.g. OIDC sub"`
	SubjectKind         string     `json:"subjectKind"         orm:"subject_kind"          description:"Subject classification: oidc_sub, openid, unionid, custom"`
	AppContext          string     `json:"appContext"          orm:"app_context"           description:"Multi-app context such as WeChat appId"`
	PluginId            string     `json:"pluginId"            orm:"plugin_id"             description:"Calling plugin ID stamped by the host when the linkage was created"`
	EmailSnapshot       string     `json:"emailSnapshot"       orm:"email_snapshot"        description:"Email captured at link time for audit only, never used as a resolution key"`
	PhoneSnapshot       string     `json:"phoneSnapshot"       orm:"phone_snapshot"        description:"Phone captured at link time for audit only"`
	DisplayNameSnapshot string     `json:"displayNameSnapshot" orm:"display_name_snapshot" description:"Display name captured at link time"`
	AvatarUrlSnapshot   string     `json:"avatarUrlSnapshot"   orm:"avatar_url_snapshot"   description:"Avatar URL captured at link time"`
	CreatedAt           *time.Time `json:"createdAt"           orm:"created_at"            description:"Creation time"`
	UpdatedAt           *time.Time `json:"updatedAt"           orm:"updated_at"            description:"Update time"`
	DeletedAt           *time.Time `json:"deletedAt"           orm:"deleted_at"            description:"Soft delete time; live rows keep NULL"`
}

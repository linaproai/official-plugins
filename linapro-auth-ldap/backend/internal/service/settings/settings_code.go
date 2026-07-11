package settings

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	CodeStorageUnavailable = bizerr.MustDefine(
		"PLUGIN_AUTH_LDAP_SETTINGS_STORAGE_UNAVAILABLE",
		"Settings storage service is unavailable",
		gcode.CodeInvalidOperation,
	)
	CodeReadFailed = bizerr.MustDefine(
		"PLUGIN_AUTH_LDAP_SETTINGS_READ_FAILED",
		"Failed to read LDAP settings",
		gcode.CodeInternalError,
	)
	CodeSaveFailed = bizerr.MustDefine(
		"PLUGIN_AUTH_LDAP_SETTINGS_SAVE_FAILED",
		"Failed to save LDAP settings",
		gcode.CodeInternalError,
	)
	CodeTLSInvalid = bizerr.MustDefine(
		"PLUGIN_AUTH_LDAP_SETTINGS_TLS_INVALID",
		"Plain LDAP is only allowed for localhost; use ldaps or starttls",
		gcode.CodeInvalidParameter,
	)
	CodeConfigInvalid = bizerr.MustDefine(
		"PLUGIN_AUTH_LDAP_SETTINGS_CONFIG_INVALID",
		"LDAP settings are invalid",
		gcode.CodeInvalidParameter,
	)
)

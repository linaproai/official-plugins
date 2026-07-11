// ldapauth_code.go declares business error codes for LDAP login.

package ldapauth

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	CodeConfigMissing = bizerr.MustDefine(
		"PLUGIN_AUTH_LDAP_CONFIG_MISSING",
		"LDAP directory configuration is missing",
		gcode.CodeInvalidConfiguration,
	)
	CodeAuthFailed = bizerr.MustDefine(
		"PLUGIN_AUTH_LDAP_AUTH_FAILED",
		"Invalid directory username or password",
		gcode.CodeNotAuthorized,
	)
	CodeDirectoryUnavailable = bizerr.MustDefine(
		"PLUGIN_AUTH_LDAP_DIRECTORY_UNAVAILABLE",
		"Directory service is unavailable",
		gcode.CodeInternalError,
	)
	CodeExternalLoginUnavailable = bizerr.MustDefine(
		"PLUGIN_AUTH_LDAP_EXTERNAL_LOGIN_UNAVAILABLE",
		"External-login service is unavailable",
		gcode.CodeInvalidOperation,
	)
	CodeExternalLoginFailed = bizerr.MustDefine(
		"PLUGIN_AUTH_LDAP_EXTERNAL_LOGIN_FAILED",
		"LDAP external login failed",
		gcode.CodeInternalError,
	)
	CodeUsernameRequired = bizerr.MustDefine(
		"PLUGIN_AUTH_LDAP_USERNAME_REQUIRED",
		"Username is required",
		gcode.CodeInvalidParameter,
	)
	CodePasswordRequired = bizerr.MustDefine(
		"PLUGIN_AUTH_LDAP_PASSWORD_REQUIRED",
		"Password is required",
		gcode.CodeInvalidParameter,
	)
)

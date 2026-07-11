// settings_code.go declares linapro-oidc-generic settings business error codes.

package settings

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	CodeStorageUnavailable = bizerr.MustDefine(
		"PLUGIN_OIDC_GENERIC_SETTINGS_STORAGE_UNAVAILABLE",
		"Settings storage service is unavailable",
		gcode.CodeInvalidOperation,
	)
	CodeReadFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_GENERIC_SETTINGS_READ_FAILED",
		"Failed to read Generic OIDC settings",
		gcode.CodeInternalError,
	)
	CodeSaveFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_GENERIC_SETTINGS_SAVE_FAILED",
		"Failed to save Generic OIDC settings",
		gcode.CodeInternalError,
	)
	CodeIssuerInvalid = bizerr.MustDefine(
		"PLUGIN_OIDC_GENERIC_SETTINGS_ISSUER_INVALID",
		"Issuer must be an absolute https URL (http allowed only for localhost)",
		gcode.CodeInvalidParameter,
	)
)

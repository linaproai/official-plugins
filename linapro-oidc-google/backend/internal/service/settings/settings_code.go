// settings_code.go declares linapro-oidc-google settings business error codes.

package settings

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeStorageUnavailable reports that the host sys_config service is unavailable.
	CodeStorageUnavailable = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_SETTINGS_STORAGE_UNAVAILABLE",
		"Settings storage service is unavailable",
		gcode.CodeInvalidOperation,
	)
	// CodeReadFailed reports that reading persisted settings failed.
	CodeReadFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_SETTINGS_READ_FAILED",
		"Failed to read Google OIDC settings",
		gcode.CodeInternalError,
	)
	// CodeSaveFailed reports that persisting settings values failed.
	CodeSaveFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_SETTINGS_SAVE_FAILED",
		"Failed to save Google OIDC settings",
		gcode.CodeInternalError,
	)
	// CodeRulesInvalid reports a backend-redirect rules payload that is not a
	// JSON object of string receiver URLs.
	CodeRulesInvalid = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_SETTINGS_RULES_INVALID",
		"Backend redirect rules must be a JSON object mapping state keys to receiver URLs",
		gcode.CodeInvalidParameter,
	)
)

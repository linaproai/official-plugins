package settings

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	CodeStorageUnavailable = bizerr.MustDefine(
		"PLUGIN_QINIU_SETTINGS_STORAGE_UNAVAILABLE",
		"Settings storage service is unavailable",
		gcode.CodeInvalidOperation,
	)
	CodeReadFailed = bizerr.MustDefine(
		"PLUGIN_QINIU_SETTINGS_READ_FAILED",
		"Failed to read object storage settings",
		gcode.CodeInternalError,
	)
	CodeSaveFailed = bizerr.MustDefine(
		"PLUGIN_QINIU_SETTINGS_SAVE_FAILED",
		"Failed to save object storage settings",
		gcode.CodeInternalError,
	)
	CodeConfigInvalid = bizerr.MustDefine(
		"PLUGIN_QINIU_SETTINGS_CONFIG_INVALID",
		"Object storage settings are incomplete or invalid",
		gcode.CodeInvalidParameter,
	)
	CodeTestFailed = bizerr.MustDefine(
		"PLUGIN_QINIU_SETTINGS_TEST_FAILED",
		"Object storage connectivity test failed",
		gcode.CodeInvalidOperation,
	)
)

// This file defines linapro-oidc-generic business error codes.

package oauth

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	CodeConfigMissing = bizerr.MustDefine(
		"PLUGIN_OIDC_GENERIC_CONFIG_MISSING",
		"OIDC login is not configured yet. Enter Issuer, Client ID, and Client Secret in the plugin settings first",
		gcode.CodeInvalidConfiguration,
	)
	CodeStateGenerateFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_GENERIC_STATE_GENERATE_FAILED",
		"Failed to generate OIDC state value",
		gcode.CodeInternalError,
	)
	CodeCallbackCodeRequired = bizerr.MustDefine(
		"PLUGIN_OIDC_GENERIC_CALLBACK_CODE_REQUIRED",
		"Authorization code is required",
		gcode.CodeInvalidParameter,
	)
	CodeCallbackStateMismatch = bizerr.MustDefine(
		"PLUGIN_OIDC_GENERIC_CALLBACK_STATE_MISMATCH",
		"OIDC state value mismatch",
		gcode.CodeSecurityReason,
	)
	CodeIdentityVerifyFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_GENERIC_IDENTITY_VERIFY_FAILED",
		"Failed to verify OIDC identity",
		gcode.CodeInternalError,
	)
	CodeExternalLoginUnavailable = bizerr.MustDefine(
		"PLUGIN_OIDC_GENERIC_EXTERNAL_LOGIN_UNAVAILABLE",
		"External-login service is unavailable",
		gcode.CodeInvalidOperation,
	)
	CodeExternalLoginFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_GENERIC_EXTERNAL_LOGIN_FAILED",
		"Generic OIDC external login failed",
		gcode.CodeInternalError,
	)
	CodeDiscoveryFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_GENERIC_DISCOVERY_FAILED",
		"Could not reach the identity provider. Check that the Issuer URL is correct and reachable from the application server",
		gcode.CodeInternalError,
	)
)

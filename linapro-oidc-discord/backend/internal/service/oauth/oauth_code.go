// This file defines linapro-oidc-discord business error codes and runtime i18n
// metadata for the Discord OIDC login orchestration.

package oauth

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeConfigMissing reports that the plugin-owned Discord client configuration is missing.
	CodeConfigMissing = bizerr.MustDefine(
		"PLUGIN_OIDC_DISCORD_CONFIG_MISSING",
		"Discord OIDC client configuration is missing",
		gcode.CodeInvalidConfiguration,
	)
	// CodeStateGenerateFailed reports that generating one CSRF state value failed.
	CodeStateGenerateFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_DISCORD_STATE_GENERATE_FAILED",
		"Failed to generate OIDC state value",
		gcode.CodeInternalError,
	)
	// CodeCallbackCodeRequired reports that the Discord callback did not carry an authorization code.
	CodeCallbackCodeRequired = bizerr.MustDefine(
		"PLUGIN_OIDC_DISCORD_CALLBACK_CODE_REQUIRED",
		"Authorization code is required",
		gcode.CodeInvalidParameter,
	)
	// CodeCallbackStateMismatch reports that the callback state value did not match the expected value.
	CodeCallbackStateMismatch = bizerr.MustDefine(
		"PLUGIN_OIDC_DISCORD_CALLBACK_STATE_MISMATCH",
		"OIDC state value mismatch",
		gcode.CodeSecurityReason,
	)
	// CodeIdentityVerifyFailed reports that exchanging the code for a verified identity failed.
	CodeIdentityVerifyFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_DISCORD_IDENTITY_VERIFY_FAILED",
		"Failed to verify Discord identity",
		gcode.CodeInternalError,
	)
	// CodeEmailNotVerified reports that the Discord account email is not verified.
	CodeEmailNotVerified = bizerr.MustDefine(
		"PLUGIN_OIDC_DISCORD_EMAIL_NOT_VERIFIED",
		"Discord account email is not verified",
		gcode.CodeNotAuthorized,
	)
	// CodeExternalLoginUnavailable reports that the host external-login service is missing.
	CodeExternalLoginUnavailable = bizerr.MustDefine(
		"PLUGIN_OIDC_DISCORD_EXTERNAL_LOGIN_UNAVAILABLE",
		"External-login service is unavailable",
		gcode.CodeInvalidOperation,
	)
	// CodeExternalLoginFailed reports that the host external-login exchange failed.
	CodeExternalLoginFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_DISCORD_EXTERNAL_LOGIN_FAILED",
		"Discord external login failed",
		gcode.CodeInternalError,
	)
)

// This file defines linapro-oidc-google business error codes and runtime i18n
// metadata for the Google OIDC login orchestration.

package oauth

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeConfigMissing reports that the plugin-owned Google client configuration is missing.
	CodeConfigMissing = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_CONFIG_MISSING",
		"Google OIDC client configuration is missing",
		gcode.CodeInvalidConfiguration,
	)
	// CodeStateGenerateFailed reports that generating one CSRF state value failed.
	CodeStateGenerateFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_STATE_GENERATE_FAILED",
		"Failed to generate OIDC state value",
		gcode.CodeInternalError,
	)
	// CodeCallbackCodeRequired reports that the Google callback did not carry an authorization code.
	CodeCallbackCodeRequired = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_CALLBACK_CODE_REQUIRED",
		"Authorization code is required",
		gcode.CodeInvalidParameter,
	)
	// CodeCallbackStateMismatch reports that the callback state value did not match the expected value.
	CodeCallbackStateMismatch = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_CALLBACK_STATE_MISMATCH",
		"OIDC state value mismatch",
		gcode.CodeSecurityReason,
	)
	// CodeIdentityVerifyFailed reports that exchanging the code for a verified identity failed.
	CodeIdentityVerifyFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_IDENTITY_VERIFY_FAILED",
		"Failed to verify Google identity",
		gcode.CodeInternalError,
	)
	// CodeEmailNotVerified reports that the Google account email is not verified.
	CodeEmailNotVerified = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_EMAIL_NOT_VERIFIED",
		"Google account email is not verified",
		gcode.CodeNotAuthorized,
	)
	// CodeExternalLoginUnavailable reports that the host external-login service is missing.
	CodeExternalLoginUnavailable = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_EXTERNAL_LOGIN_UNAVAILABLE",
		"External-login service is unavailable",
		gcode.CodeInvalidOperation,
	)
	// CodeExternalLoginFailed reports that the host external-login exchange failed.
	CodeExternalLoginFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_EXTERNAL_LOGIN_FAILED",
		"Google external login failed",
		gcode.CodeInternalError,
	)
	// CodeOneTapCSRFMismatch reports that the GSI double-submit CSRF check failed.
	CodeOneTapCSRFMismatch = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_ONE_TAP_CSRF_MISMATCH",
		"One Tap CSRF validation failed",
		gcode.CodeSecurityReason,
	)
	// CodeOneTapDisabled reports that the One Tap endpoint was called while the
	// admin switch is off.
	CodeOneTapDisabled = bizerr.MustDefine(
		"PLUGIN_OIDC_GOOGLE_ONE_TAP_DISABLED",
		"Google One Tap login is disabled",
		gcode.CodeInvalidOperation,
	)
)

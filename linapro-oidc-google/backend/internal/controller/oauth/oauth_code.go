// This file defines Google OAuth controller business error codes.

package oauth

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeOAuthProviderDisabled reports that the Google provider is not
	// currently enabled by the host plugin state snapshot.
	CodeOAuthProviderDisabled = bizerr.MustDefineWithKey(
		"provider_disabled",
		"authentication.oauthHandoff.errors.providerDisabled",
		"This authentication provider is currently disabled",
		gcode.CodeNotAuthorized,
	)
	// CodeOAuthSettingsUnavailable reports that Google OAuth settings are
	// unavailable or incomplete for the current request.
	CodeOAuthSettingsUnavailable = bizerr.MustDefineWithKey(
		"settings_unavailable",
		"authentication.oauthHandoff.errors.settingsUnavailable",
		"OAuth provider settings are temporarily unavailable",
		gcode.CodeInternalError,
	)
	// CodeOAuthAuthorizeURLFailed reports that the authorization redirect URL
	// could not be built from the configured Google OAuth settings.
	CodeOAuthAuthorizeURLFailed = bizerr.MustDefineWithKey(
		"authorize_url_failed",
		"authentication.oauthHandoff.errors.authorizeUrlFailed",
		"Failed to start OAuth login",
		gcode.CodeInternalError,
	)
	// CodeOAuthMissingCodeOrState reports a callback without required OAuth
	// callback parameters.
	CodeOAuthMissingCodeOrState = bizerr.MustDefineWithKey(
		"missing_code_or_state",
		"authentication.oauthHandoff.errors.missingCodeOrState",
		"OAuth callback is missing required parameters",
		gcode.CodeInvalidParameter,
	)
	// CodeOAuthInvalidState reports an OAuth callback whose state token failed
	// validation.
	CodeOAuthInvalidState = bizerr.MustDefineWithKey(
		"invalid_state",
		"authentication.oauthHandoff.errors.invalidState",
		"OAuth state validation failed",
		gcode.CodeInvalidParameter,
	)
	// CodeOAuthCodeExchangeFailed reports a failed authorization-code exchange.
	CodeOAuthCodeExchangeFailed = bizerr.MustDefineWithKey(
		"code_exchange_failed",
		"authentication.oauthHandoff.errors.codeExchangeFailed",
		"Failed to exchange the authorization code",
		gcode.CodeInternalError,
	)
	// CodeOAuthUserinfoFailed reports a failed external user profile request.
	CodeOAuthUserinfoFailed = bizerr.MustDefineWithKey(
		"userinfo_failed",
		"authentication.oauthHandoff.errors.userinfoFailed",
		"Failed to fetch the external user profile",
		gcode.CodeInternalError,
	)
	// CodeOAuthEmailNotVerified reports that Google did not return a verified
	// email address.
	CodeOAuthEmailNotVerified = bizerr.MustDefineWithKey(
		"email_not_verified",
		"authentication.oauthHandoff.errors.emailNotVerified",
		"The provider did not return a verified email address",
		gcode.CodeNotAuthorized,
	)
	// CodeOAuthEmptyLoginResult reports that the host login handoff returned no
	// usable token or tenant-selection payload.
	CodeOAuthEmptyLoginResult = bizerr.MustDefineWithKey(
		"empty_login_result",
		"authentication.oauthHandoff.errors.emptyLoginResult",
		"OAuth login returned an empty result",
		gcode.CodeInternalError,
	)
)

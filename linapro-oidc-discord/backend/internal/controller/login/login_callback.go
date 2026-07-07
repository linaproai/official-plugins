// login_callback.go implements the browser-facing GET /callback route that
// consumes the Discord callback and hands the verified identity to the host
// external-login seam.

package login

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/logger"

	oauthsvc "lina-plugin-linapro-oidc-discord/backend/internal/service/oauth"
	settingssvc "lina-plugin-linapro-oidc-discord/backend/internal/service/settings"
)

// defaultReturnPath is used when the plugin cannot resolve one from the
// configured login return path. The SPA renders its login page at this path
// in the default LinaPro workspace layout; the outcome query must live inside
// the hash fragment so the SPA hash router exposes it through route.query.
const defaultReturnPath = "/admin/#/auth/login"

// Callback handles GET /portal/linapro-oidc-discord/callback. It forwards the
// callback values to the OAuth service, which validates the self-contained
// signed state (no cookie required), and redirects the browser back to the
// SPA login page with the outcome encoded in the query string.
func (c *ControllerV1) Callback(request *ghttp.Request) {
	ctx := request.Context()
	callback, err := c.oauthSvc.CompleteCallback(ctx, oauthsvc.CallbackInput{
		Code:  request.Get("code").String(),
		State: request.Get("state").String(),
	})
	returnPath := c.resolveReturnPath()
	if err != nil {
		logger.Warningf(ctx, "linapro-oidc-discord callback failed: %v", err)
		redirectURL := buildErrorRedirect(returnPath, err.Error())
		request.Response.RedirectTo(redirectURL, http.StatusFound)
		return
	}
	// The business state key is recovered from the signed state token.
	stateKey := strings.TrimSpace(callback.StateKey)
	settings := c.loadSettingsSnapshot(ctx)
	// SSO token delivery: when enabled and the business state key matches one
	// configured rule, the tokens are delivered straight to the third-party
	// receiver URL and the browser never visits the SPA login page.
	if receiver := resolveSSOReceiver(ctx, settings, stateKey, callback); receiver != "" {
		request.Response.RedirectTo(buildReceiverRedirect(receiver, stateKey, callback), http.StatusFound)
		return
	}
	redirectURL := buildSuccessRedirect(returnPath, callback, resolveSPALanding(settings))
	request.Response.RedirectTo(redirectURL, http.StatusFound)
}

// loadSettingsSnapshot reads the persisted settings, degrading to nil so a
// temporarily unavailable settings store never breaks the login redirect.
func (c *ControllerV1) loadSettingsSnapshot(ctx context.Context) *settingssvc.Snapshot {
	if c.settingsSvc == nil {
		return nil
	}
	snapshot, err := c.settingsSvc.Load(ctx)
	if err != nil {
		logger.Warningf(ctx, "linapro-oidc-discord callback settings load failed: %v", err)
		return nil
	}
	return snapshot
}

// resolveSPALanding returns the configured SPA landing path for normal logins
// or an empty string to keep the host default landing.
func resolveSPALanding(settings *settingssvc.Snapshot) string {
	if settings == nil {
		return ""
	}
	return strings.TrimSpace(settings.DefaultBackendRedirect)
}

// resolveSSOReceiver returns the third-party receiver URL when SSO delivery is
// enabled, the login produced a token pair, and the business state key matches
// one configured rule. Pre-token (multi-tenant) outcomes always fall back to
// the SPA flow because tenant selection requires the workspace UI.
func resolveSSOReceiver(ctx context.Context, settings *settingssvc.Snapshot, stateKey string, callback *oauthsvc.CallbackOutput) string {
	if settings == nil || !settings.EnableBackendRedirect || stateKey == "" {
		return ""
	}
	if callback == nil || callback.AccessToken == "" {
		return ""
	}
	rulesJSON := strings.TrimSpace(settings.BackendRedirects)
	if rulesJSON == "" {
		return ""
	}
	var rules map[string]string
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		logger.Warningf(ctx, "linapro-oidc-discord backend redirect rules malformed: %v", err)
		return ""
	}
	return strings.TrimSpace(rules[stateKey])
}

// buildReceiverRedirect appends the token pair and the echoed state key to the
// receiver URL so the third-party system can establish its own session.
func buildReceiverRedirect(receiver string, stateKey string, callback *oauthsvc.CallbackOutput) string {
	query := url.Values{}
	query.Set("provider", "discord")
	query.Set("state", stateKey)
	query.Set("accessToken", callback.AccessToken)
	if callback.RefreshToken != "" {
		query.Set("refreshToken", callback.RefreshToken)
	}
	separator := "?"
	if strings.Contains(receiver, "?") {
		separator = "&"
	}
	return receiver + separator + query.Encode()
}

// resolveReturnPath returns the SPA login return path with a safe default.
// Falling back to a stable SPA path guarantees the browser is never stranded
// on the plugin callback URL.
func (c *ControllerV1) resolveReturnPath() string {
	if path := c.oauthSvc.LoginReturnPath(); path != "" {
		return path
	}
	return defaultReturnPath
}

// buildErrorRedirect renders one SPA redirect URL for a failed callback.
func buildErrorRedirect(returnPath string, message string) string {
	query := url.Values{}
	query.Set("externalLogin", "1")
	query.Set("provider", "discord")
	query.Set("status", "error")
	query.Set("message", message)
	return appendQuery(returnPath, query)
}

// buildSuccessRedirect renders one SPA redirect URL for a successful callback.
// spaLanding optionally overrides the workspace landing path the SPA navigates
// to after storing the tokens.
func buildSuccessRedirect(returnPath string, callback *oauthsvc.CallbackOutput, spaLanding string) string {
	query := url.Values{}
	query.Set("externalLogin", "1")
	query.Set("provider", "discord")
	if spaLanding != "" {
		query.Set("redirect", spaLanding)
	}
	if callback == nil {
		query.Set("status", "error")
		query.Set("message", "empty external login response")
		return appendQuery(returnPath, query)
	}
	if callback.PreToken != "" {
		query.Set("status", "select-tenant")
		query.Set("preToken", callback.PreToken)
		query.Set("tenantCount", strconv.Itoa(len(callback.TenantCandidates)))
		if encoded := encodeTenantCandidates(callback); encoded != "" {
			query.Set("tenants", encoded)
		}
		return appendQuery(returnPath, query)
	}
	query.Set("status", "signed-in")
	if callback.AccessToken != "" {
		query.Set("accessToken", callback.AccessToken)
	}
	if callback.RefreshToken != "" {
		query.Set("refreshToken", callback.RefreshToken)
	}
	return appendQuery(returnPath, query)
}

// encodeTenantCandidates serializes the tenant candidates into the compact
// JSON array the SPA login page parses to render the two-stage tenant
// selection after an external login. Encoding failures degrade to an empty
// string so the redirect never breaks over an optional enrichment.
func encodeTenantCandidates(callback *oauthsvc.CallbackOutput) string {
	if callback == nil || len(callback.TenantCandidates) == 0 {
		return ""
	}
	type tenantProjection struct {
		Id     int    `json:"id"`
		Code   string `json:"code"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	items := make([]tenantProjection, 0, len(callback.TenantCandidates))
	for _, tenant := range callback.TenantCandidates {
		items = append(items, tenantProjection{
			Id:     tenant.ID,
			Code:   tenant.Code,
			Name:   tenant.Name,
			Status: tenant.Status,
		})
	}
	encoded, err := json.Marshal(items)
	if err != nil {
		return ""
	}
	return string(encoded)
}

// appendQuery attaches encoded query parameters to the given return path
// while preserving any query the path itself already carries.
func appendQuery(returnPath string, query url.Values) string {
	if returnPath == "" {
		returnPath = defaultReturnPath
	}
	separator := "?"
	if strings.Contains(returnPath, "?") {
		separator = "&"
	}
	return returnPath + separator + query.Encode()
}

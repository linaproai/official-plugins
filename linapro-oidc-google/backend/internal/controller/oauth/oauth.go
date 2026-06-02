// Package oauth wires the Google OAuth2 authorization-code endpoints exposed
// by the linapro-oidc-google source plugin. The initiation endpoint redirects
// browsers to Google's authorization URL with a signed state token, and the
// callback endpoint validates the state, exchanges the authorization code for
// a Google access token, fetches the verified userinfo, hands the identity
// off to the host auth service, and finally redirects the browser to the
// frontend OAuth handoff page with the host login outcome.
package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	netUrl "net/url"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	plugincontract "lina-core/pkg/plugin/capability/contract"

	configsvc "lina-plugin-linapro-oidc-google/backend/internal/service/config"
	oauthsvc "lina-plugin-linapro-oidc-google/backend/internal/service/oauth"
)

// callbackPath is the OAuth callback path registered by the plugin. It
// lives under the host's /api/v1/* namespace so deployments that already
// proxy the API prefix do not need additional reverse-proxy rules for the
// OAuth callback URL registered in Google Cloud Console.
const callbackPath = "/api/v1/auth/google/callback"

// Controller implements the Google OAuth2 authorization-code endpoints.
type Controller struct {
	// authSvc is the host auth contract used to convert verified Google
	// identities into host login outcomes (token pair or pre-login handoff).
	authSvc plugincontract.AuthService
	// settingsSvc reads the typed Google OIDC settings on every callback so
	// admin updates take effect without restarting the plugin.
	settingsSvc *configsvc.Service
	// pluginState reads the host-owned provider enablement state. The Google
	// plugin no longer owns a private enabled flag in settings.
	pluginState plugincontract.PluginStateService
	// oauthSvc encapsulates the Google authorization-code flow.
	oauthSvc *oauthsvc.Service
}

// New constructs a Controller bound to the host auth contract, the shared
// Google OIDC settings service, and the host-owned PluginState capability.
// Callers must pass a non-nil pluginState; the plugin registrar rejects
// construction when the host has not published the unified provider
// enablement contract so the controller never operates without it.
//
// pluginState is consulted on every StartLogin and HandleCallback request
// through isProviderEnabled to decide whether the Google provider may run.
// authSvc converts verified external identities into host login outcomes.
// settingsSvc provides per-request typed access to the masked OAuth client
// credentials and SSO delivery rules stored under sys_config.
func New(authSvc plugincontract.AuthService, settingsSvc *configsvc.Service, pluginState plugincontract.PluginStateService) *Controller {
	return &Controller{
		authSvc:     authSvc,
		settingsSvc: settingsSvc,
		pluginState: pluginState,
		oauthSvc:    oauthsvc.New(),
	}
}

// StartLogin handles GET /api/v1/auth/google. It first asks the host
// PluginState capability whether the Google provider is platform-enabled so
// disabled plugins short-circuit without touching sys_config, then loads the
// current plugin settings, builds Google's authorization URL with a signed
// state token, and 302 redirects the browser to Google.
func (c *Controller) StartLogin(req *ghttp.Request) {
	ctx := req.Context()
	if !c.isProviderEnabled(ctx) {
		c.writeError(req, CodeOAuthProviderDisabled, bizerr.NewCode(CodeOAuthProviderDisabled))
		return
	}
	settings, err := c.settingsSvc.Get(ctx)
	if err != nil {
		c.writeError(req, CodeOAuthSettingsUnavailable, bizerr.WrapCode(err, CodeOAuthSettingsUnavailable))
		return
	}
	if settings == nil {
		c.writeError(req, CodeOAuthSettingsUnavailable, bizerr.NewCode(CodeOAuthSettingsUnavailable))
		return
	}
	stateKey := strings.TrimSpace(req.Get("state").String())
	redirectURI := c.resolveRedirectURI(req, settings.RedirectURI)
	authorizeURL, _, err := c.oauthSvc.BuildAuthorizeURL(toOAuthSettings(settings), redirectURI, stateKey)
	if err != nil {
		c.writeError(req, CodeOAuthAuthorizeURLFailed, bizerr.WrapCode(err, CodeOAuthAuthorizeURLFailed))
		return
	}
	req.Response.RedirectTo(authorizeURL, 302)
}

// HandleCallback handles GET /api/v1/auth/google/callback. It first checks
// the unified provider enablement state so disabled plugins fail fast before
// touching sys_config, then validates the state, exchanges the authorization
// code for a Google access token, fetches the verified userinfo, hands the
// identity off to the host login flow, and redirects to the frontend OAuth
// handoff page with the host login outcome.
func (c *Controller) HandleCallback(req *ghttp.Request) {
	ctx := req.Context()
	if errParam := strings.TrimSpace(req.Get("error").String()); errParam != "" {
		c.redirectWithError(req, "/", errParam)
		return
	}
	code := strings.TrimSpace(req.Get("code").String())
	rawState := strings.TrimSpace(req.Get("state").String())
	if code == "" || rawState == "" {
		c.redirectWithCode(req, "/", CodeOAuthMissingCodeOrState)
		return
	}
	if !c.isProviderEnabled(ctx) {
		c.redirectWithCode(req, "/", CodeOAuthProviderDisabled)
		return
	}
	settings, err := c.settingsSvc.Get(ctx)
	if err != nil {
		c.writeError(req, CodeOAuthSettingsUnavailable, bizerr.WrapCode(err, CodeOAuthSettingsUnavailable))
		return
	}
	if settings == nil {
		c.redirectWithCode(req, "/", CodeOAuthProviderDisabled)
		return
	}
	statePayload, err := c.oauthSvc.DecodeState(rawState, settings.ClientSecret)
	if err != nil {
		logger.Warningf(ctx, "google oauth state validation failed err=%v", err)
		c.redirectWithCode(req, "/", CodeOAuthInvalidState)
		return
	}
	redirectURI := c.resolveRedirectURI(req, settings.RedirectURI)
	googleAccessToken, err := c.oauthSvc.ExchangeCode(ctx, toOAuthSettings(settings), redirectURI, code)
	if err != nil {
		logger.Warningf(ctx, "google oauth code exchange failed err=%v", err)
		c.redirectWithCode(req, "/", CodeOAuthCodeExchangeFailed)
		return
	}
	identity, err := c.oauthSvc.FetchUserIdentity(ctx, googleAccessToken)
	if err != nil {
		logger.Warningf(ctx, "google oauth userinfo failed err=%v", err)
		c.redirectWithCode(req, "/", CodeOAuthUserinfoFailed)
		return
	}
	if !identity.EmailVerified {
		c.redirectWithCode(req, "/", CodeOAuthEmailNotVerified)
		return
	}
	hostOutput, err := c.authSvc.LoginByExternal(ctx, plugincontract.ExternalLoginInput{
		ProviderID:     oauthsvc.ProviderID,
		PluginID:       oauthsvc.PluginID,
		ExternalUserID: identity.Subject,
		Email:          identity.Email,
		DisplayName:    identity.Name,
		ClientIP:       req.GetClientIp(),
	})
	if err != nil {
		hostCode := classifyHostLoginError(err)
		logger.Warningf(ctx, "google oauth host login handoff failed email=%s code=%s err=%v", identity.Email, hostCode, err)
		c.redirectWithError(req, "/", hostCode)
		return
	}
	// Two independent post-login paths:
	// - SSO token delivery: when enableBackendRedirect=true and the state
	//   key matches one of the configured backendRedirects rules, the
	//   tokens are delivered directly to the rule's external URL as query
	//   parameters and the browser never visits the SPA handoff page.
	// - SPA handoff: in every other case (rule miss, multi-tenant
	//   preToken, backend redirect disabled, host login error) the browser
	//   lands on /oauth-handoff inside the workbench so the SPA can store
	//   tokens, run tenant selection if needed, and navigate to the SPA
	//   landing URL configured by defaultBackendRedirect.
	if settings.EnableBackendRedirect && hostOutput != nil && hostOutput.AccessToken != "" {
		if ssoTarget := c.resolveSSOReceiver(ctx, statePayload.StateKey, settings.BackendRedirects); ssoTarget != "" {
			c.redirectToSSOReceiver(req, ssoTarget, statePayload.StateKey, hostOutput)
			return
		}
	}
	c.redirectToSPAHandoff(req, c.resolveSPALanding(settings.DefaultBackendRedirect), statePayload.StateKey, hostOutput)
}

// resolveRedirectURI returns the OAuth redirect URI registered with Google.
// It prefers the explicitly configured value and falls back to deriving the
// URL from the current request when no value is configured.
func (c *Controller) resolveRedirectURI(req *ghttp.Request, configured string) string {
	if value := strings.TrimSpace(configured); value != "" {
		return value
	}
	scheme := "https"
	if req.TLS == nil {
		scheme = "http"
	}
	host := req.Host
	if host == "" {
		host = req.Request.Host
	}
	return scheme + "://" + host + callbackPath
}

// resolveSPALanding returns the SPA route the workbench navigates to after
// the OAuth handoff page consumes the host login outcome. It is
// intentionally independent of the SSO token delivery rules: the operator
// uses this value to control "after a normal login, where does the user
// land inside the workbench". It falls back to /dashboard/analytics when the
// operator did not configure a custom landing page.
func (c *Controller) resolveSPALanding(defaultRedirect string) string {
	landing := strings.TrimSpace(defaultRedirect)
	if landing == "" {
		return "/dashboard/analytics"
	}
	return landing
}

// resolveSSOReceiver looks up the SSO token-delivery URL for the supplied
// state key. It returns an empty string when no rule matches so the caller
// can fall back to the SPA handoff flow without sending the token outside
// the workbench. The defaultBackendRedirect field is deliberately NOT
// consulted here because SSO delivery is opt-in per state key.
func (c *Controller) resolveSSOReceiver(ctx context.Context, stateKey string, rulesJSON string) string {
	trimmedKey := strings.TrimSpace(stateKey)
	if trimmedKey == "" {
		return ""
	}
	trimmedJSON := strings.TrimSpace(rulesJSON)
	if trimmedJSON == "" {
		return ""
	}
	var rules map[string]string
	if err := json.Unmarshal([]byte(trimmedJSON), &rules); err != nil {
		logger.Warningf(ctx, "google oauth backend redirect rules malformed err=%v", err)
		return ""
	}
	if value, ok := rules[trimmedKey]; ok {
		return strings.TrimSpace(value)
	}
	return ""
}

// redirectToSSOReceiver delivers the host login tokens directly to the
// external receiver URL configured by an SSO rule. The tokens are appended
// as URL query parameters so the receiving system can extract them
// without running a SPA handoff page; the browser is never sent to the
// workbench for this code path.
func (c *Controller) redirectToSSOReceiver(
	req *ghttp.Request,
	receiverURL string,
	stateKey string,
	output *plugincontract.ExternalLoginOutput,
) {
	values := netUrl.Values{}
	values.Set("provider", oauthsvc.ProviderID)
	if stateKey != "" {
		values.Set("state", stateKey)
	}
	if output != nil {
		if output.AccessToken != "" {
			values.Set("accessToken", output.AccessToken)
		}
		if output.RefreshToken != "" {
			values.Set("refreshToken", output.RefreshToken)
		}
	}
	separator := "?"
	if strings.Contains(receiverURL, "?") {
		separator = "&"
	}
	req.Response.RedirectTo(receiverURL+separator+values.Encode(), 302)
}

// redirectToSPAHandoff sends the browser to the workbench /oauth-handoff
// route with the host login outcome encoded as query parameters. The
// handoff page reads the outcome, stores tokens (or runs tenant selection
// for multi-tenant users), and then navigates to the SPA landing URL.
func (c *Controller) redirectToSPAHandoff(
	req *ghttp.Request,
	spaLanding string,
	stateKey string,
	output *plugincontract.ExternalLoginOutput,
) {
	values := map[string]string{
		"provider": oauthsvc.ProviderID,
	}
	if stateKey != "" {
		values["state"] = stateKey
	}
	if redirect := strings.TrimSpace(spaLanding); redirect != "" {
		values["redirect"] = redirect
	}
	switch {
	case output != nil && output.AccessToken != "":
		values["accessToken"] = output.AccessToken
		if output.RefreshToken != "" {
			values["refreshToken"] = output.RefreshToken
		}
	case output != nil && output.PreToken != "":
		values["preToken"] = output.PreToken
		if encodedTenants, err := encodeTenants(output.Tenants); err == nil && encodedTenants != "" {
			values["tenants"] = encodedTenants
		}
	default:
		c.redirectWithCode(req, spaLanding, CodeOAuthEmptyLoginResult)
		return
	}
	req.Response.RedirectTo(c.authSvc.OAuthHandoffURL(req.Context(), values), 302)
}

// redirectWithError sends the browser to the frontend OAuth handoff page with
// a structured failure reason so the UI can display an actionable error.
func (c *Controller) redirectWithError(req *ghttp.Request, target string, reason string) {
	values := map[string]string{
		"provider": oauthsvc.ProviderID,
		"error":    reason,
	}
	if redirect := strings.TrimSpace(target); redirect != "" {
		values["redirect"] = redirect
	}
	req.Response.RedirectTo(c.authSvc.OAuthHandoffURL(req.Context(), values), 302)
}

// redirectWithCode sends the browser to the frontend OAuth handoff page with
// the runtime code from one local bizerr definition. This keeps local plugin
// OAuth failures centralized in oauth_code.go while preserving the query-code
// handoff contract consumed by the SPA.
func (c *Controller) redirectWithCode(req *ghttp.Request, target string, code *bizerr.Code) {
	c.redirectWithError(req, target, code.RuntimeCode())
}

// writeError logs and returns a non-redirecting 4xx response. It is reserved
// for unrecoverable misconfigurations (missing settings, disabled provider)
// where redirecting back to the handoff page would mask the root cause.
func (c *Controller) writeError(req *ghttp.Request, code *bizerr.Code, cause error) {
	logger.Warningf(req.Context(), "google oauth request rejected code=%s err=%v", code.RuntimeCode(), cause)
	req.Response.WriteHeader(400)
	req.Response.WriteJsonExit(g.Map{
		"providerId":    oauthsvc.ProviderID,
		"error":         code.RuntimeCode(),
		"errorCode":     code.RuntimeCode(),
		"fallback":      code.Fallback(),
		"messageKey":    code.MessageKey(),
		"message":       code.Fallback(),
		"messageParams": g.Map{},
	})
}

// toOAuthSettings projects plugin settings into the OAuth service input shape.
func toOAuthSettings(settings *configsvc.Settings) oauthsvc.Settings {
	if settings == nil {
		return oauthsvc.Settings{}
	}
	return oauthsvc.Settings{
		ClientID:     settings.ClientID,
		ClientSecret: settings.ClientSecret,
		RedirectURI:  settings.RedirectURI,
	}
}

// isProviderEnabled reports whether the Google provider may currently serve
// OAuth requests by consulting the host-owned PluginState capability through
// IsProviderEnabled, which reads the platform plugin enabled snapshot.
//
// The check is fail-closed: a nil controller or a missing PluginState
// dependency returns false so disabled or misconstructed plugins cannot
// silently keep serving OAuth. Callers must use this single seam instead of
// any plugin-private "enabled" toggle so the workbench login button, OAuth
// initiation, and OAuth callback share one authoritative source of truth.
//
// The snapshot is owned by the host plugin lifecycle service; freshness and
// recovery semantics are documented in the OpenSpec tasks record. The check
// is cheap (in-memory map lookup) and is intentionally performed before
// reading sys_config so disabled plugins do not incur a settings round-trip.
func (c *Controller) isProviderEnabled(ctx context.Context) bool {
	return c != nil && c.pluginState != nil && c.pluginState.IsProviderEnabled(ctx, configsvc.PluginID)
}

// classifyHostLoginError extracts the stable runtime error code from a host
// LoginByExternal failure so the frontend handoff page can distinguish
// "no local account linked" from other login policy rejections without
// leaking internal error text or breaking i18n.
func classifyHostLoginError(err error) string {
	if err == nil {
		return ""
	}
	if structured, ok := bizerr.As(err); ok {
		if code := strings.TrimSpace(structured.RuntimeCode()); code != "" {
			return code
		}
	}
	return "AUTH_EXTERNAL_LOGIN_FAILED"
}

// encodeTenants serializes tenant candidates as a base64url-encoded JSON
// array so the frontend handoff page can parse the multi-tenant selection
// input without re-querying the backend.
func encodeTenants(tenants []plugincontract.ExternalLoginTenant) (string, error) {
	if len(tenants) == 0 {
		return "", nil
	}
	type tenantView struct {
		ID     int    `json:"id"`
		Code   string `json:"code"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	views := make([]tenantView, 0, len(tenants))
	for _, item := range tenants {
		views = append(views, tenantView{
			ID:     item.ID,
			Code:   item.Code,
			Name:   item.Name,
			Status: item.Status,
		})
	}
	raw, err := json.Marshal(views)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// Package oauth wires the Discord OAuth2 authorization-code endpoints
// exposed by the linapro-oidc-discord source plugin. The initiation endpoint
// redirects browsers to Discord's authorization URL with a signed state
// token, and the callback endpoint validates the state, exchanges the
// authorization code for a Discord access token, fetches the verified
// userinfo, hands the identity off to the host auth service, and finally
// redirects the browser to the frontend OAuth handoff page with the host
// login outcome.
package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	netUrl "net/url"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	plugincontract "lina-core/pkg/plugin/capability/contract"

	configsvc "lina-plugin-linapro-oidc-discord/backend/internal/service/config"
	oauthsvc "lina-plugin-linapro-oidc-discord/backend/internal/service/oauth"
)

// callbackPath is the OAuth callback path registered by the plugin. It
// lives under the host's /api/v1/* namespace so deployments that already
// proxy the API prefix do not need additional reverse-proxy rules for the
// OAuth callback URL registered in Discord Developer Portal.
const callbackPath = "/api/v1/auth/discord/callback"

// Controller implements the Discord OAuth2 authorization-code endpoints.
type Controller struct {
	// authSvc is the host auth contract used to convert verified Discord
	// identities into host login outcomes (token pair or pre-login handoff).
	authSvc plugincontract.AuthService
	// settingsSvc reads the typed Discord OAuth2 settings on every callback
	// so admin updates take effect without restarting the plugin.
	settingsSvc *configsvc.Service
	// oauthSvc encapsulates the Discord authorization-code flow.
	oauthSvc *oauthsvc.Service
}

// New constructs a Controller bound to the host auth contract and the
// shared Discord OAuth2 settings service.
func New(authSvc plugincontract.AuthService, settingsSvc *configsvc.Service) *Controller {
	return &Controller{
		authSvc:     authSvc,
		settingsSvc: settingsSvc,
		oauthSvc:    oauthsvc.New(),
	}
}

// StartLogin handles GET /api/v1/auth/discord. It loads the current plugin
// settings, builds Discord's authorization URL with a signed state token,
// and 302 redirects the browser to Discord.
func (c *Controller) StartLogin(req *ghttp.Request) {
	ctx := req.Context()
	settings, err := c.settingsSvc.Get(ctx)
	if err != nil {
		c.writeError(req, "discord login settings unavailable", err)
		return
	}
	if settings == nil || !settings.Enabled {
		c.writeError(req, "discord login is disabled", gerror.New("plugin disabled"))
		return
	}
	stateKey := strings.TrimSpace(req.Get("state").String())
	redirectURI := c.resolveRedirectURI(req, settings.RedirectURI)
	authorizeURL, _, err := c.oauthSvc.BuildAuthorizeURL(toOAuthSettings(settings), redirectURI, stateKey)
	if err != nil {
		c.writeError(req, "build discord authorize url failed", err)
		return
	}
	req.Response.RedirectTo(authorizeURL, 302)
}

// HandleCallback handles GET /api/v1/auth/discord/callback. It validates
// the state, exchanges the authorization code for a Discord access token,
// fetches the verified userinfo, hands the identity off to the host
// login flow, and redirects to the frontend OAuth handoff page with the
// host login outcome.
func (c *Controller) HandleCallback(req *ghttp.Request) {
	ctx := req.Context()
	if errParam := strings.TrimSpace(req.Get("error").String()); errParam != "" {
		c.redirectWithError(req, "/", errParam)
		return
	}
	code := strings.TrimSpace(req.Get("code").String())
	rawState := strings.TrimSpace(req.Get("state").String())
	if code == "" || rawState == "" {
		c.redirectWithError(req, "/", "missing_code_or_state")
		return
	}
	settings, err := c.settingsSvc.Get(ctx)
	if err != nil {
		c.writeError(req, "discord login settings unavailable", err)
		return
	}
	if settings == nil || !settings.Enabled {
		c.redirectWithError(req, "/", "provider_disabled")
		return
	}
	statePayload, err := c.oauthSvc.DecodeState(rawState, settings.ClientSecret)
	if err != nil {
		logger.Warningf(ctx, "discord oauth state validation failed err=%v", err)
		c.redirectWithError(req, "/", "invalid_state")
		return
	}
	redirectURI := c.resolveRedirectURI(req, settings.RedirectURI)
	discordAccessToken, err := c.oauthSvc.ExchangeCode(ctx, toOAuthSettings(settings), redirectURI, code)
	if err != nil {
		logger.Warningf(ctx, "discord oauth code exchange failed err=%v", err)
		c.redirectWithError(req, "/", "code_exchange_failed")
		return
	}
	identity, err := c.oauthSvc.FetchUserIdentity(ctx, discordAccessToken)
	if err != nil {
		logger.Warningf(ctx, "discord oauth userinfo failed err=%v", err)
		c.redirectWithError(req, "/", "userinfo_failed")
		return
	}
	if !identity.EmailVerified {
		c.redirectWithError(req, "/", "email_not_verified")
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
		logger.Warningf(ctx, "discord oauth host login handoff failed email=%s code=%s err=%v", identity.Email, hostCode, err)
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

// resolveRedirectURI returns the OAuth redirect URI registered with Discord.
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
// land inside the workbench". It falls back to /dashboard when the
// operator did not configure a custom landing page.
func (c *Controller) resolveSPALanding(defaultRedirect string) string {
	landing := strings.TrimSpace(defaultRedirect)
	if landing == "" {
		return "/dashboard"
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
		logger.Warningf(ctx, "discord oauth backend redirect rules malformed err=%v", err)
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
		c.redirectWithError(req, spaLanding, "empty_login_result")
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

// writeError logs and returns a non-redirecting 4xx response. It is reserved
// for unrecoverable misconfigurations (missing settings, disabled provider)
// where redirecting back to the handoff page would mask the root cause.
func (c *Controller) writeError(req *ghttp.Request, message string, cause error) {
	logger.Warningf(req.Context(), "discord oauth request rejected reason=%s err=%v", message, cause)
	req.Response.WriteHeader(400)
	req.Response.WriteJsonExit(g.Map{
		"providerId": oauthsvc.ProviderID,
		"error":      message,
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
		Enabled:      settings.Enabled,
	}
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

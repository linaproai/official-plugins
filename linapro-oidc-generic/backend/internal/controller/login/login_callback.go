// login_callback.go consumes the IdP callback and redirects to SPA with handoff.

package login

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"

	oauthsvc "lina-plugin-linapro-oidc-generic/backend/internal/service/oauth"
	settingssvc "lina-plugin-linapro-oidc-generic/backend/internal/service/settings"
)

const defaultReturnPath = "/admin/auth/login"
const providerQueryValue = oauthsvc.Provider

// Callback handles GET /portal/linapro-oidc-generic/callback.
func (c *ControllerV1) Callback(request *ghttp.Request) {
	ctx := request.Context()
	callback, err := c.oauthSvc.CompleteCallback(ctx, oauthsvc.CallbackInput{
		Code:  request.Get("code").String(),
		State: request.Get("state").String(),
	})
	returnPath := c.resolveReturnPath()
	if callback != nil && strings.TrimSpace(callback.ReturnTo) != "" {
		returnPath = callback.ReturnTo
	}
	if err != nil {
		logger.Warningf(ctx, "linapro-oidc-generic callback failed: %v", err)
		redirectURL := buildErrorRedirect(returnPath, safeExternalLoginErrorMessage(err))
		request.Response.RedirectTo(redirectURL, http.StatusFound)
		return
	}
	settings := c.loadSettingsSnapshot(ctx)
	redirectURL := buildSuccessRedirect(returnPath, callback, resolveSPALanding(settings))
	request.Response.RedirectTo(redirectURL, http.StatusFound)
}

func (c *ControllerV1) loadSettingsSnapshot(ctx context.Context) *settingssvc.Snapshot {
	if c.settingsSvc == nil {
		return nil
	}
	snapshot, err := c.settingsSvc.Load(ctx)
	if err != nil {
		logger.Warningf(ctx, "linapro-oidc-generic callback settings load failed: %v", err)
		return nil
	}
	return snapshot
}

func resolveSPALanding(settings *settingssvc.Snapshot) string {
	if settings == nil {
		return ""
	}
	return strings.TrimSpace(settings.DefaultBackendRedirect)
}

func (c *ControllerV1) resolveReturnPath() string {
	if path := c.oauthSvc.LoginReturnPath(); path != "" {
		return path
	}
	return defaultReturnPath
}

func safeExternalLoginErrorMessage(err error) string {
	if err == nil {
		return "external_login_failed"
	}
	if be, ok := bizerr.As(err); ok && be != nil {
		if code := strings.TrimSpace(be.RuntimeCode()); code != "" {
			return code
		}
	}
	return "external_login_failed"
}

func buildErrorRedirect(returnPath string, message string) string {
	query := url.Values{}
	query.Set("externalLogin", "1")
	query.Set("provider", providerQueryValue)
	query.Set("status", "error")
	query.Set("message", message)
	return appendQuery(returnPath, query)
}

func buildSuccessRedirect(returnPath string, callback *oauthsvc.CallbackOutput, spaLanding string) string {
	query := url.Values{}
	query.Set("externalLogin", "1")
	query.Set("provider", providerQueryValue)
	if spaLanding != "" {
		query.Set("redirect", spaLanding)
	}
	if callback == nil || strings.TrimSpace(callback.Handoff) == "" {
		query.Set("status", "error")
		query.Set("message", "external_login_failed")
		return appendQuery(returnPath, query)
	}
	query.Set("status", "signed-in")
	query.Set("handoff", callback.Handoff)
	return appendQuery(returnPath, query)
}

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

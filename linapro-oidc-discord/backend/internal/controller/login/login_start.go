// login_start.go implements the browser-facing GET /login route that
// initiates the Discord OIDC authorize redirect. The anti-CSRF state is a
// self-contained HMAC-signed token embedded in the authorize URL, so no
// cookie is required to survive the cross-site round trip through Discord.

package login

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/logger"
	oauthsvc "lina-plugin-linapro-oidc-discord/backend/internal/service/oauth"
)

// Start handles GET /portal/linapro-oidc-discord/login. It asks the OAuth
// service for one authorize URL carrying the signed state (with the optional
// ?state=<key> business key embedded) and issues an HTTP 302 redirect to
// Discord. On configuration errors it redirects back to the SPA login page
// with an error hint instead of stranding the browser on the plugin route.
func (c *ControllerV1) Start(request *ghttp.Request) {
	var (
		ctx      = request.Context()
		stateKey = request.Get("state").String()
		returnTo = oauthsvc.SanitizeReturnTo(request.Get("returnTo").String())
	)
	authorize, err := c.oauthSvc.BuildAuthorizeURL(ctx, stateKey, returnTo)
	if err != nil {
		logger.Warningf(ctx, "linapro-oidc-discord authorize URL build failed: %v", err)
		path := returnTo
		if path == "" {
			path = c.resolveReturnPath()
		}
		redirectURL := buildErrorRedirect(path, safeExternalLoginErrorMessage(err))
		request.Response.RedirectTo(redirectURL, http.StatusFound)
		return
	}
	request.Response.RedirectTo(authorize.URL, http.StatusFound)
}

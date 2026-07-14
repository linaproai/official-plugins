// login_start.go initiates the OIDC authorize redirect.

package login

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/logger"
	oauthsvc "lina-plugin-linapro-oidc-generic/backend/internal/service/oauth"
)

// Start handles GET /portal/linapro-oidc-generic/login.
func (c *ControllerV1) Start(request *ghttp.Request) {
	var (
		ctx      = request.Context()
		stateKey = request.Get("state").String()
		returnTo = oauthsvc.SanitizeReturnTo(request.Get("returnTo").String())
	)
	authorize, err := c.oauthSvc.BuildAuthorizeURL(ctx, stateKey, returnTo)
	if err != nil {
		logger.Warningf(ctx, "linapro-oidc-generic authorize URL build failed: %v", err)
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

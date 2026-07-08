// login_onetap.go implements the browser-facing POST /onetap route that
// consumes the Google One Tap (GSI login_uri) form POST from embeddable
// snippets on third-party pages. The GSI library submits the ID Token as a
// top-level form navigation, so the handler responds with the same redirect
// semantics as the authorize-code callback: SSO delivery to the matched
// receiver URL, or the SPA login handoff.

package login

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"

	oauthsvc "lina-plugin-linapro-oidc-google/backend/internal/service/oauth"
)

// gsiCSRFTokenName is the cookie and form field name Google's GSI library uses
// for its double-submit CSRF token on login_uri form POSTs.
const gsiCSRFTokenName = "g_csrf_token"

// OneTap handles POST /portal/linapro-oidc-google/onetap. It applies Google's
// double-submit CSRF check when the g_csrf_token cookie is present, validates
// the ID Token credential through the OAuth service, and reuses the SSO
// delivery / SPA handoff redirect logic shared with the authorize-code flow.
func (c *ControllerV1) OneTap(request *ghttp.Request) {
	ctx := request.Context()
	returnPath := c.resolveReturnPath()
	// Google's GSI form POST carries a double-submit CSRF token. Enforce the
	// comparison whenever the cookie made it across (same-site embeds); pure
	// cross-site embeds may miss the cookie, where the ID Token's signature,
	// audience, and expiry remain the integrity anchor.
	csrfCookie := strings.TrimSpace(request.Cookie.Get(gsiCSRFTokenName).String())
	csrfBody := strings.TrimSpace(request.Get(gsiCSRFTokenName).String())
	if csrfCookie != "" && subtle.ConstantTimeCompare([]byte(csrfCookie), []byte(csrfBody)) != 1 {
		logger.Warningf(ctx, "linapro-oidc-google one tap csrf token mismatch")
		request.Response.RedirectTo(
			buildErrorRedirect(returnPath, bizerr.NewCode(oauthsvc.CodeOneTapCSRFMismatch).Error()),
			http.StatusFound,
		)
		return
	}
	credential := request.Get("credential").String()
	stateKey := strings.TrimSpace(request.Get("state").String())
	callback, err := c.oauthSvc.CompleteOneTap(ctx, credential, stateKey)
	if err != nil {
		logger.Warningf(ctx, "linapro-oidc-google one tap failed: %v", err)
		request.Response.RedirectTo(buildErrorRedirect(returnPath, err.Error()), http.StatusFound)
		return
	}
	settings := c.loadSettingsSnapshot(ctx)
	// SSO token delivery: identical semantics to the authorize-code callback.
	if receiver := resolveSSOReceiver(ctx, settings, stateKey, callback); receiver != "" {
		request.Response.RedirectTo(buildReceiverRedirect(receiver, stateKey, callback), http.StatusFound)
		return
	}
	request.Response.RedirectTo(buildSuccessRedirect(returnPath, callback, resolveSPALanding(settings)), http.StatusFound)
}

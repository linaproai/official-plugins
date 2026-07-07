# linapro-oidc-google

`linapro-oidc-google` is a reference source plugin for `LinaPro` that demonstrates how to plug a third-party OIDC provider (Google) into the new host external-identity seam.

English | [简体中文](README.zh-CN.md)

## What This Sample Demonstrates

- Declares external-identity provider ownership at plugin init through `plugin.Providers().ProvideExternalIdentity("google")` so the host will only accept `LoginByVerifiedIdentity` calls that name a provider owned by this plugin.
- Registers two browser-facing routes on a plugin-owned portal path group:
  - `GET /portal/linapro-oidc-google/login` builds the Google authorize URL, persists an anti-CSRF state cookie, and redirects the browser to Google.
  - `GET /portal/linapro-oidc-google/callback` validates the state, exchanges the OAuth code for a verified identity, calls `Services.Auth().ExternalLogin().LoginByVerifiedIdentity(...)`, and redirects the browser back to the SPA login page with the host outcome encoded in the query string.
- Ships one Vue frontend slot component under `frontend/slots/auth.login.after/google-login-entry.vue` that the workspace slot registry auto-mounts below the SPA login form when the plugin is installed and enabled.

The OAuth code exchange and userinfo fetch are intentionally minimal and clearly stubbed. The reference verifier returns a deterministic verified identity from the incoming authorization code so the surrounding orchestration can be exercised end-to-end without a live Google project. A real deployment must swap `oauthsvc.NewStubIdentityVerifier()` for an HTTP-backed implementation before this plugin can go to production.

## Directory Layout

```text
linapro-oidc-google/
  plugin.yaml
  plugin_embed.go
  go.mod
  Makefile
  backend/
    plugin.go
    internal/
      controller/login/           login-start + callback handlers
      service/oauth/              authorize URL construction, state, verifier, callback
  frontend/
    slots/auth.login.after/       "Continue with Google" button component
  manifest/
    i18n/en-US/                   English i18n resources
    i18n/zh-CN/                   Simplified Chinese i18n resources
  README.md
  README.zh-CN.md
```

## New External-Identity Seam Integration

The host owns provisioning, tenant candidate resolution, and token minting. The plugin only submits the verified identity DTO and forwards the outcome to the SPA login page.

| Step | Actor | Action |
| --- | --- | --- |
| 1. User clicks "Continue with Google" | Browser | Full-page navigation to `/portal/linapro-oidc-google/login`. |
| 2. Plugin builds authorize URL | Plugin | `oauthsvc.BuildAuthorizeURL(ctx, ...)` returns a Google authorize URL plus a fresh CSRF state; the plugin sets the state cookie and issues a 302 to Google. |
| 3. Google callback | Browser | Google redirects back to `/portal/linapro-oidc-google/callback?code=...&state=...`. |
| 4. Plugin verifies identity | Plugin | `oauthsvc.CompleteCallback(ctx, ...)` matches the cookie state, exchanges the code for `{subject, email, displayName}`. |
| 5. Host session issuance | Host | `Services.Auth().ExternalLogin().LoginByVerifiedIdentity(ctx, externallogin.LoginInput{...})`. |
| 6. SPA redirect | Plugin | The plugin redirects back to the SPA login page with `accessToken`/`refreshToken` or `preToken`+tenant candidates encoded in the query. |

Unlinked identities are rejected by the host with `AUTH_EXTERNAL_USER_NOT_PROVISIONED`. The plugin surfaces this to the user through the callback error redirect.

## Configuration

`oauth.DefaultConfig()` returns reference values with placeholder client credentials. Real deployments override the config at wiring time so a production Google Cloud project can drive the flow:

```go
cfg := oauthsvc.DefaultConfig()
cfg.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
cfg.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
cfg.RedirectURL = "https://your-host/portal/linapro-oidc-google/callback"
```

For a production reference wiring, replace the stub verifier with an HTTP-backed one that:

1. Exchanges the code for an `id_token` at `TokenURL`.
2. Validates the `id_token` signature against Google's JWKS.
3. Fetches the userinfo endpoint if needed for `displayName`.
4. Returns `{Subject, Email, DisplayName}` as `oauthsvc.VerifiedIdentity`.

## Frontend Slot

The Vue slot component lives at `frontend/slots/auth.login.after/google-login-entry.vue`. The workspace slot registry auto-discovers it at build time; the runtime registry hides the button when the plugin is not installed or is disabled. The button triggers a full-page navigation to the login-start route so the browser cookie set by the plugin controller is respected during the OIDC handshake.

## Review Checklist

- provider ownership is declared through `ProvideExternalIdentity`
- login routes stay on the plugin-owned portal path group, outside the `/x` API namespace
- provisioning, tenant resolution, and token minting stay host-owned
- OAuth stub is documented and easy to replace
- Vue slot component is only rendered when the plugin is installed and enabled

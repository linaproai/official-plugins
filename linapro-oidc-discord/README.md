# linapro-oidc-discord

`linapro-oidc-discord` is the official source plugin that adds Discord sign-in to the `LinaPro` login page. It ships OAuth settings, optional auto-provisioning, and depends on `linapro-extlogin-core` for identity linking and account provisioning.

English | [简体中文](README.zh-CN.md)

## What This Sample Demonstrates

- Declares external-identity provider ownership at plugin init through `plugin.Providers().ProvideExternalIdentity("discord")` so the host will only accept `LoginByVerifiedIdentity` calls that name a provider owned by this plugin.
- Registers two browser-facing routes on a plugin-owned portal path group:
  - `GET /portal/linapro-oidc-discord/login` builds the Discord authorize URL, persists an anti-CSRF state cookie, and redirects the browser to Discord.
  - `GET /portal/linapro-oidc-discord/callback` validates the state, exchanges the OAuth code for a verified identity, calls `Services.Auth().ExternalLogin().LoginByVerifiedIdentity(...)`, and redirects the browser back to the SPA login page with the host outcome encoded in the query string.
- Ships one Vue frontend slot component under `frontend/slots/auth.login.social/discord-login-entry.vue` that the workspace slot registry auto-mounts as a platform social icon under “其他登录方式” when the plugin is installed and enabled.

The OAuth code exchange and userinfo fetch are intentionally minimal and clearly stubbed. The reference verifier returns a deterministic verified identity from the incoming authorization code so the surrounding orchestration can be exercised end-to-end without a live Discord application. A real deployment must swap `oauthsvc.NewStubIdentityVerifier()` for an HTTP-backed implementation before this plugin can go to production.

## Directory Layout

```text
linapro-oidc-discord/
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
    slots/auth.login.social/      Discord social icon entry (tooltip button)
  manifest/
    i18n/en-US/                   English i18n resources
    i18n/zh-CN/                   Simplified Chinese i18n resources
  README.md
  README.zh-CN.md
```

## New External-Identity Seam Integration

The host owns tenant candidate resolution and token minting, and delegates `(provider, subject)` linkage resolution plus provisioning policy to `linapro-extlogin-core`, which this plugin declares as a dependency in `plugin.yaml`. When `linapro-extlogin-core` is not installed or enabled, external login is fail-closed. This plugin only submits the verified identity DTO and forwards the outcome to the SPA login page.

| Step | Actor | Action |
| --- | --- | --- |
| 1. User clicks "Continue with Discord" | Browser | Full-page navigation to `/portal/linapro-oidc-discord/login`. |
| 2. Plugin builds authorize URL | Plugin | `oauthsvc.BuildAuthorizeURL(ctx, ...)` returns a Discord authorize URL plus a fresh CSRF state; the plugin sets the state cookie and issues a 302 to Discord. |
| 3. Discord callback | Browser | Discord redirects back to `/portal/linapro-oidc-discord/callback?code=...&state=...`. |
| 4. Plugin verifies identity | Plugin | `oauthsvc.CompleteCallback(ctx, ...)` matches the cookie state, exchanges the code for `{subject, email, displayName}`. |
| 5. Host session issuance | Host | `Services.Auth().ExternalLogin().LoginByVerifiedIdentity(ctx, extlogin.LoginInput{...})`. |
| 6. SPA redirect | Plugin | The plugin redirects back to the SPA login page with `accessToken`/`refreshToken` or `preToken`+tenant candidates encoded in the query. |

Unlinked identities are rejected by the host with `AUTH_EXTERNAL_USER_NOT_PROVISIONED`. The plugin surfaces this to the user through the callback error redirect.

## Configuration

`oauth.DefaultConfig()` returns reference values with placeholder client credentials. Real deployments override the config at wiring time so a production Discord application can drive the flow:

```go
cfg := oauthsvc.DefaultConfig()
cfg.ClientID = os.Getenv("DISCORD_CLIENT_ID")
cfg.ClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
cfg.RedirectURL = "https://your-host/portal/linapro-oidc-discord/callback"
```

For a production reference wiring, replace the stub verifier with an HTTP-backed one that:

1. Exchanges the code for an access token at `TokenURL` (`https://discord.com/api/oauth2/token`).
2. Calls the userinfo endpoint (`https://discord.com/api/users/@me`) with the access token.
3. Projects the returned `id`, `email`, and `global_name`/`username` into `oauthsvc.VerifiedIdentity`.

## Frontend Slot

The Vue slot component lives at `frontend/slots/auth.login.social/discord-login-entry.vue`. The workspace slot registry auto-discovers it at build time; the runtime registry hides the icon when the plugin is not installed or is disabled. Clicking the icon triggers a full-page navigation to the login-start route so the browser cookie set by the plugin controller is respected during the OIDC handshake. Protocol / directory logins (generic OIDC, LDAP) use the separate `auth.login.after` full-width button stack instead.

Its `pluginSlotMeta.order` is `20`, so when both `linapro-oidc-google` (order `10`) and this plugin are enabled, Google renders first and Discord renders below it.

## Review Checklist

- provider ownership is declared through `ProvideExternalIdentity`
- login routes stay on the plugin-owned portal path group, outside the `/x` API namespace
- tenant resolution and token minting stay host-owned; linkage storage and provisioning policy stay in `linapro-extlogin-core`
- OAuth stub is documented and easy to replace
- Vue slot component is only rendered when the plugin is installed and enabled

## Dependency and handoff

- Depends on managed `linapro-extlogin-core` (install and enable first).
- Successful login redirects carry only a one-time `handoff`; the SPA exchanges it via managed plugin `linapro-extlogin-core` handoff API.

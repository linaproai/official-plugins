# linapro-oidc-generic

`linapro-oidc-generic` is an official LinaPro **managed** source plugin that adds
configurable enterprise OIDC sign-in to the login page. It depends on
`linapro-extid-core` for identity linkage, provisioning policy, and SPA handoff.

[English](README.md) | [简体中文](README.zh-CN.md)

## Capability boundary

| Owns | Does not own |
| --- | --- |
| OIDC Authorization Code + PKCE, Discovery, JWKS id_token verify | Linkage table / provision engine (`linapro-extid-core`) |
| Admin settings (issuer, client, scopes, JIT flag) | Host JWT / session minting |
| Login portal routes + `auth.login.after` entry | Multi-connection UI (v1.1), LDAP, tenant-scoped IdP |

**Provider encoding (v1):** `oidc:default` with `subject = id_token.sub`.

## Install order

1. Install and enable `linapro-extid-core`
2. Install and enable `linapro-oidc-generic`
3. Configure issuer / client credentials on the settings page
4. Register the callback URL at the IdP

## Login flow

1. User clicks OIDC entry → `GET /portal/linapro-oidc-generic/login`
2. Plugin runs Discovery, builds authorize URL with PKCE + state + nonce
3. IdP callback → code exchange → verify `id_token`
4. `LoginByVerifiedIdentity(provider=oidc:default, ...)`
5. SPA receives **handoff only** (no JWT in URL) via `linapro-extid-core` exchange

## Settings

Persisted through host `sys_config` (created on first save; no mandatory SQL seed):

- Issuer (https; http only for localhost)
- Client ID / Secret (secret masked on read)
- Optional redirect override (default: derived callback URL)
- Scopes (always includes `openid`)
- Display name, workspace landing path
- **Allow auto-provision: default off**

## Security checklist

- [x] PKCE S256
- [x] HMAC state + nonce
- [x] JWKS signature + iss/aud/exp/sub
- [x] Fail-closed when credentials/issuer missing
- [x] Handoff-only SPA redirect
- [x] Secret never projected in plaintext

## Review checklist

- Depends on `linapro-extid-core` in `plugin.yaml`
- `ProvideExternalIdentity("oidc:default")`
- No linkage table in this plugin
- Managed distribution; not builtin

## Directory

```text
linapro-oidc-generic/
  plugin.yaml
  backend/plugin.go
  backend/internal/service/oauth/
  backend/internal/service/settings/
  frontend/pages/settings.vue
  frontend/slots/auth.login.after/
  manifest/i18n/
```

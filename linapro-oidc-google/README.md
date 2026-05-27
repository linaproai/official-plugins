# linapro-oidc-google

`linapro-oidc-google` is the official LinaPro source plugin for Google login entry and OIDC provider metadata.

## Scope

This plugin owns:

- Google login entry presentation on the default workbench
- OIDC provider metadata for the Google / Gmail login route
- plugin-owned configuration and front-end login entry rendering

## Host Boundary

The host keeps auth-provider discovery, login-page aggregation, session issuance, and token handoff. This plugin adds Google-specific entry presentation and provider metadata only.

## Directory Layout

```text
linapro-oidc-google/
  plugin.yaml
  plugin_embed.go
  backend/
  frontend/
  manifest/
```

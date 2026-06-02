# linapro-oidc-discord

`linapro-oidc-discord` is the official LinaPro source plugin for Discord login entry and OAuth2 provider metadata.

## Scope

This plugin owns:

- Discord login entry presentation on the default workbench
- OAuth2 provider metadata for the Discord login route
- plugin-owned configuration and front-end login entry rendering

## Host Boundary

The host keeps auth-provider discovery, login-page aggregation, session issuance, and token handoff. This plugin adds Discord-specific entry presentation and provider metadata only.

## Directory Layout

```text
linapro-oidc-discord/
  plugin.yaml
  plugin_embed.go
  backend/
  frontend/
  manifest/
```

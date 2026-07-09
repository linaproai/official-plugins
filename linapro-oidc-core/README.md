# linapro-oidc-core

`linapro-oidc-core` is the official LinaPro source plugin that owns external identity storage and the provider engine behind the host external-login seam.

English | [简体中文](README.zh-CN.md)

## Scope

This plugin owns:

- the `user_external_identity` linkage table mapping a verified `(provider, subject)` pair to a local user account
- the `externalidentityspi.Provider` engine: resolution, auto-provisioning policy (same-email conflict rejection, email-less deterministic username anchors, idempotent `(provider, subject)` de-duplication), and least-privilege account creation delegated back to the host user owner through `usercap.ProvisionExternal`
- the current-user identity binding API: list, bind, and unbind under `/x/linapro-oidc-core/api/v1/plugins/linapro-oidc-core/identities`

OAuth protocol plugins such as `linapro-oidc-google` and `linapro-oidc-discord` depend on this plugin. They keep verifying identities and calling the host `externallogin` seam unchanged; this plugin never calls `LoginByVerifiedIdentity` and declares no provider-ID ownership.

## Host Boundary

The host keeps token minting, session persistence, tenant resolution, pre-login-token handoff, login-IP policy, and auth hooks. This plugin registers only an engine factory through `plugin.Providers().ProvideExternalIdentityProvider(...)`; the host manager lazily constructs the provider gated by plugin enablement and binds the manager-backed seam into the auth service.

When this plugin is not installed or is disabled, external login is fail-closed: no linkage resolves, no account is provisioned, and no session is issued. Re-enabling the plugin restores external login immediately. Disabling keeps the linkage table data; uninstalling with storage purge drops the table but never cascade-deletes provisioned user accounts, which stay governable through host user management.

## Data Permission Boundary

External identity linkages are self-isolated user resources:

- login resolution uses only the authoritative `(provider, subject)` partial unique index and reports a uniform not-provisioned outcome, never leaking whether an email exists on another account
- bind, unbind, and list act exclusively on the current session user's own linkages; cross-user targets are rejected as a whole
- a `(provider, subject)` pair already owned by another account is rejected with a conflict error; re-binding an identity the current user already owns succeeds idempotently

## Provisioning Contract

Host account provisioning and the plugin linkage write cannot share one database transaction across module boundaries. Correctness converges on the `(provider, subject)` partial unique index: the account is provisioned first, then the linkage is inserted; a unique-index conflict from a concurrent provision is absorbed by reusing the winning linkage instead of surfacing a 500. Email-less identities derive a deterministic, collision-resistant username anchor from the identity digest, so retries reuse the same host account.

## Directory Layout

```text
linapro-oidc-core/
  plugin.yaml
  plugin_embed.go
  go.mod
  Makefile
  backend/
    plugin.go
    api/identity/v1/              list/bind/unbind DTOs
    internal/
      controller/identity/        current-user identity binding handlers
      service/identity/           provider engine and linkage policy
      dao/, model/do/, model/entity/  generated table access objects
  manifest/
    sql/                          install SQL for user_external_identity
    sql/uninstall/                uninstall SQL (drop table on purge)
    i18n/en-US/, i18n/zh-CN/      error and API documentation resources
  hack/config.yaml                plugin-local DAO generation config
```

## Review Checklist

- linkage storage, provisioning policy, and binding stay inside this plugin
- token, session, and tenant minting stay host-owned
- the engine factory is declared through `ProvideExternalIdentityProvider`; no provider-ID ownership is declared
- disable keeps linkage data and fails external login closed; uninstall purge drops only the plugin table

# linapro-extid-core

`linapro-extid-core` is the official **managed** source plugin that owns the external-identity domain for third-party login: linkage storage, provisioning policy, verified tickets, provider catalog, and bind APIs. It is **not** a host-critical builtin; when not installed, the domain fails closed and dependent UI must hide.

English | [简体中文](README.zh-CN.md)

## Capability boundary

This plugin owns:

- the `user_external_identity` linkage table (`(provider, subject)` key, with subject_kind / app_context / phone snapshots)
- the host `extidspi.Provider` engine: Resolve / Provision / Bind / Unbind / List
- the plugin-owned `backend/cap/extidcap` contract: wide `Service` entry with `Ticket` / `Login` / `Linkage` / `Providers` sub surfaces, plus process-bound catalog and handoff facades
- current-user bind APIs that accept **verified tickets only** (no bare client-supplied subject)

Protocol plugins such as `linapro-oidc-google` and `linapro-oidc-discord` **must** declare a dependency on this plugin. They verify IdPs and call host `extlogin` plus this domain; this plugin never calls `LoginByVerifiedIdentity` and declares no provider-ID ownership.

## Distribution and degradation

| State | Behavior |
| --- | --- |
| Not installed / disabled | External login fail-closed; protocol plugins blocked by dependency checks; third-party login entries hidden |
| Enabled | Protocol plugins can complete login; SPA exchanges a host handoff for the session |
| Disable | Keep linkage table; login fail-closed; re-enable restores |
| Uninstall (purge) | Drop plugin table; **do not** cascade-delete provisioned `sys_user` rows |

Install order: `linapro-extid-core` → protocol plugin → credentials → enable.

## Host boundary

The host keeps token minting, sessions, tenants, pre-tokens, login IP policy, auth hooks, and **one-time login handoffs** (protocol plugins must not put JWTs in SPA redirect URLs). This plugin registers only the engine factory via `ProvideExternalIdentityProvider`.

## Review checklist

- `distribution: managed` (host does not auto-install)
- protocol plugins depend on `linapro-extid-core`
- bind is ticket-only; SPA uses handoff only
- token/session stay host-owned
- password login unaffected when core is absent

## SPA handoff

Protocol plugins call `extidcap.CreateLoginHandoffFromHost` after host minting (owner plugin binds a process-local store at startup); the SPA redeems codes via this plugin's public `POST .../handoff/exchange`. The host exposes no handoff HTTP surface.

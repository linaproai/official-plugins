# Third-party Login - OIDC

This document is bundled under the plugin manifest and is intended for plugin-market display. Runtime behavior is unchanged by the documentation resource.

## Summary

Provides configurable enterprise OIDC third-party login for providers such as Keycloak, Okta, and Azure AD—including SSO, identity linking, and optional auto-provisioning for login entries and other product flows. Requires Third-party Login - Core.

## Documents

- [Configuration](configuration.md)
- [Changelog](changelog.md)

## Feature Highlights

- Provides configurable enterprise OIDC third-party login for providers such as Keycloak, Okta, and Azure AD—including SSO, identity linking, and optional auto-provisioning for login entries and other product flows. Requires Third-party Login - Core.
- Provides workbench entry points for OIDC.
- Depends on `linapro-extlogin-core`.

## Where It Fits

Use it to add governed third-party login entries, identity linking, and optional account provisioning.

## Entry Points

| Name | Path |
| --- | --- |
| OIDC | `linapro-oidc-generic-settings` |

## Metadata

| Field | Description |
| --- | --- |
| Plugin ID | `linapro-oidc-generic` |
| Version | `v0.1.0` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| Multi-tenant | No |

# Third-party Login - Google

This document is bundled under the plugin manifest and is intended for plugin-market display. Runtime behavior is unchanged by the documentation resource.

## Summary

Provides general-purpose Google third-party login (OAuth settings, optional auto-provisioning, identity linking, and One Tap embed) for login entries, account binding, and other product flows. Requires Third-party Login - Core.

## Documents

- [Configuration](configuration.md)
- [Changelog](changelog.md)

## Feature Highlights

- Provides general-purpose Google third-party login (OAuth settings, optional auto-provisioning, identity linking, and One Tap embed) for login entries, account binding, and other product flows. Requires Third-party Login - Core.
- Provides workbench entry points for Google.
- Depends on `linapro-extlogin-core`.

## Where It Fits

Use it to add governed third-party login entries, identity linking, and optional account provisioning.

## Entry Points

| Name | Path |
| --- | --- |
| Google | `linapro-oidc-google-settings` |

## Metadata

| Field | Description |
| --- | --- |
| Plugin ID | `linapro-oidc-google` |
| Version | `v0.1.0` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| Multi-tenant | No |

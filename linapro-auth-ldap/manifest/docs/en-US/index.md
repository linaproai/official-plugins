# Third-party Login - LDAP

This document is bundled under the plugin manifest and is intended for plugin-market display. Runtime behavior is unchanged by the documentation resource.

## Summary

Provides general-purpose enterprise LDAP / Active Directory directory login (directory bind authentication and optional auto-provisioning) for login entries and other product flows. Requires Third-party Login - Core.

## Documents

- [Configuration](configuration.md)
- [Changelog](changelog.md)

## Feature Highlights

- Provides general-purpose enterprise LDAP / Active Directory directory login (directory bind authentication and optional auto-provisioning) for login entries and other product flows. Requires Third-party Login - Core.
- Provides workbench entry points for LDAP.
- Depends on `linapro-extlogin-core`.

## Where It Fits

Use it to add governed third-party login entries, identity linking, and optional account provisioning.

## Entry Points

| Name | Path |
| --- | --- |
| LDAP | `linapro-auth-ldap-settings` |

## Metadata

| Field | Description |
| --- | --- |
| Plugin ID | `linapro-auth-ldap` |
| Version | `v0.1.0` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| Multi-tenant | No |

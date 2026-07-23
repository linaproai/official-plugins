# Mail - Core

This document is bundled under the plugin manifest and is intended for plugin-market display. Runtime behavior is unchanged by the documentation resource.

## Summary

Manages mail server connections and accounts, and provides the shared mail capability used by SMTP, IMAP, and POP3 plugins. Platform email notifications also go through this plugin.

## Documents

- [Configuration](configuration.md)
- [Changelog](changelog.md)

## Feature Highlights

- Manages mail server connections and accounts, and provides the shared mail capability used by SMTP, IMAP, and POP3 plugins. Platform email notifications also go through this plugin.
- Provides workbench entry points for Mail.

## Where It Fits

Use it to manage mail connections and transport capabilities for platform notification flows.

## Entry Points

| Name | Path |
| --- | --- |
| Mail | `linapro-mail-core-settings` |

## Metadata

| Field | Description |
| --- | --- |
| Plugin ID | `linapro-mail-core` |
| Version | `v0.1.2` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| Multi-tenant | No |

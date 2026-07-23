# Object Storage - Azure Blob

This document is bundled under the plugin manifest and is intended for plugin-market display. Runtime behavior is unchanged by the documentation resource.

## Summary

Provides an Azure Blob Storage backend for the host Storage domain capability.

## Documents

- [Configuration](configuration.md)
- [Changelog](changelog.md)

## Feature Highlights

- Provides an Azure Blob Storage backend for the host Storage domain capability.
- Provides workbench entry points for Azure Blob.

## Where It Fits

Use it as the active object-storage provider behind the host `Storage()` domain capability.

## Entry Points

| Name | Path |
| --- | --- |
| Azure Blob | `linapro-storage-azure-settings` |

## Metadata

| Field | Description |
| --- | --- |
| Plugin ID | `linapro-storage-azure` |
| Version | `v0.1.0` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| Multi-tenant | No |

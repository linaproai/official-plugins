# Object Storage - S3

This document is bundled under the plugin manifest and is intended for plugin-market display. Runtime behavior is unchanged by the documentation resource.

## Summary

Provides an S3 protocol object storage backend (MinIO, R2, Ceph RGW, and similar) for the host Storage domain capability.

## Documents

- [Configuration](configuration.md)
- [Changelog](changelog.md)

## Feature Highlights

- Provides an S3 protocol object storage backend (MinIO, R2, Ceph RGW, and similar) for the host Storage domain capability.
- Provides workbench entry points for Storage Management - S3.

## Where It Fits

Use it as the active object-storage provider behind the host `Storage()` domain capability.

## Entry Points

| Name | Path |
| --- | --- |
| Storage Management - S3 | `linapro-storage-s3-settings` |

## Metadata

| Field | Description |
| --- | --- |
| Plugin ID | `linapro-storage-s3` |
| Version | `v0.1.0` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| Multi-tenant | No |

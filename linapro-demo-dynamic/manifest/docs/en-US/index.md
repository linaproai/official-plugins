# Sample Plugin - Dynamic

This document is bundled under the plugin manifest and is intended for plugin-market display. Runtime behavior is unchanged by the documentation resource.

## Summary

Dynamic wasm sample that demonstrates a host-embedded menu page, plugin-owned SQL CRUD, and a hosted standalone page.

## Documents

- [Configuration](configuration.md)
- [Changelog](changelog.md)

## Feature Highlights

- Dynamic wasm sample that demonstrates a host-embedded menu page, plugin-owned SQL CRUD, and a hosted standalone page.
- Provides workbench entry points for Sample Plugin - Dynamic.
- Depends on `LinaPro framework`, `linapro-ai-core`, `linapro-demo-source`.
- Demonstrates dynamic `WASM` plugin packaging, public assets, and host-service authorization.
- Supports tenant-aware usage scenarios.

## Where It Fits

Use it as a reference implementation for LinaPro plugin development and lifecycle governance.

## Entry Points

| Name | Path |
| --- | --- |
| Sample Plugin - Dynamic | `/x-assets/linapro-demo-dynamic/v0.1.0/mount.js` |

## Metadata

| Field | Description |
| --- | --- |
| Plugin ID | `linapro-demo-dynamic` |
| Version | `v0.1.0` |
| Type | `dynamic` |
| Distribution | `managed` |
| Scope | `tenant_aware` |
| Multi-tenant | Yes |

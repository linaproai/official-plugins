# AI Hub

This document is bundled under the plugin manifest and is intended for plugin-market display. Runtime behavior is unchanged by the documentation resource.

## Summary

Official source plugin for AI providers, models, capability tiers, and request logs

## Documents

- [Configuration](configuration.md)
- [Changelog](changelog.md)

## Feature Highlights

- Official source plugin for AI providers, models, capability tiers, and request logs
- Provides workbench entry points for Providers, Models, AI Tiers, Request Logs.

## Where It Fits

Use it when LinaPro needs governed `AI` providers, provider models, capability tiers, and invocation audit records.

## Entry Points

| Name | Path |
| --- | --- |
| Providers | `/ai/providers` |
| Models | `/ai/models` |
| AI Tiers | `/ai/tiers` |
| Request Logs | `/ai/invocations` |

## Metadata

| Field | Description |
| --- | --- |
| Plugin ID | `linapro-ai-core` |
| Version | `v0.1.0` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| Multi-tenant | No |

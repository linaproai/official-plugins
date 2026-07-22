# linapro-monitor-operlog

`linapro-monitor-operlog` is the official LinaPro source plugin for operation-log persistence and governance.

English | [简体中文](README.zh-CN.md)

## Scope

This plugin owns:

- operation-log capture and persistence through the plugin-owned global middleware
- operation-log query, export, cleanup, and detail pages
- operation-log dictionaries and storage schema

## Host Boundary

The host keeps the HTTP middleware chain, authentication, and menu governance. This plugin registers a global middleware to capture operations and provides query, export, and cleanup APIs under the `monitor` catalog.

## Plugin Metadata

| Field | Value |
| --- | --- |
| Plugin ID | `linapro-monitor-operlog` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `tenant_aware` |
| Install mode | `tenant_scoped` |

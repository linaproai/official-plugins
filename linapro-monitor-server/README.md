# linapro-monitor-server

`linapro-monitor-server` is the official LinaPro source plugin for server monitoring.

English | [简体中文](README.zh-CN.md)

## Scope

This plugin owns:

- server-monitor data collection
- stale snapshot cleanup
- server-monitor query APIs and workspace entry

The host keeps the cron and plugin lifecycle substrate; this plugin supplies the monitoring capability itself.

## Host Boundary

The host keeps the scheduling infrastructure and plugin lifecycle. This plugin owns the monitoring data schema, collection logic, and query APIs under the `monitor` catalog.

## Plugin Metadata

| Field | Value |
| --- | --- |
| Plugin ID | `linapro-monitor-server` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| Install mode | `global` |

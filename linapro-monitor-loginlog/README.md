# linapro-monitor-loginlog

`linapro-monitor-loginlog` is the official LinaPro source plugin for login-log persistence and governance.

English | [简体中文](README.zh-CN.md)

## Scope

This plugin owns:

- login-log persistence subscribed from host auth lifecycle events
- login-log query, export, cleanup, and detail pages
- the login-log storage schema

## Host Boundary

The host keeps the authentication lifecycle, session infrastructure, and menu governance. This plugin subscribes to auth events and provides query, export, and cleanup APIs under the `monitor` catalog.

## Plugin Metadata

| Field | Value |
| --- | --- |
| Plugin ID | `linapro-monitor-loginlog` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `tenant_aware` |
| Install mode | `tenant_scoped` |

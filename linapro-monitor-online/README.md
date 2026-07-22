# linapro-monitor-online

`linapro-monitor-online` is the official LinaPro source plugin for online-user governance.

English | [简体中文](README.zh-CN.md)

## Scope

This plugin owns:

- online session projection queries
- force-logout governance APIs and workspace entry

It consumes the host-published online-session domain capability and does not generate or query host `sys_*` tables.

## Host Boundary

The host keeps the session infrastructure, authentication, and the online-session domain capability. This plugin consumes that capability to provide query and governance UI under the `monitor` catalog.

## Plugin Metadata

| Field | Value |
| --- | --- |
| Plugin ID | `linapro-monitor-online` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `tenant_aware` |
| Install mode | `tenant_scoped` |

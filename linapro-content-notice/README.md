# linapro-content-notice

`linapro-content-notice` is the official LinaPro source plugin for notice and announcement management.

English | [简体中文](README.zh-CN.md)

## Scope

This plugin owns:

- notice and announcement CRUD
- notice-specific dictionaries and storage schema
- the default workspace entry under the `content` host catalog

## Host Boundary

The host keeps the `content` catalog skeleton, authentication, and menu governance. This plugin adds notice-specific APIs, menus, button permissions, and storage tables under `manifest/sql/`.

## Plugin Metadata

| Field | Value |
| --- | --- |
| Plugin ID | `linapro-content-notice` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `tenant_aware` |
| Install mode | `tenant_scoped` |

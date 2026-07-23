# linapro-ops-demo-guard

`linapro-ops-demo-guard` is the official LinaPro source plugin for demo-environment read-only protection.

English | [简体中文](README.zh-CN.md)

Read-only demo mode becomes active whenever this plugin is installed and enabled. If you want the host to auto-enable it at startup, add `linapro-ops-demo-guard` to the host `plugin.autoEnable` list.

## Scope

This plugin owns:

- environment-level demo request guarding based on `HTTP Method`
- whole-system request interception on the host `/*` scope
- write-operation interception for host and plugin write endpoints across the system
- the minimal session whitelist required for login, token refresh, tenant selection, tenant switching, and logout in demo mode

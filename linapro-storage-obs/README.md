# linapro-storage-obs

Managed source plugin that provides a **Huawei Cloud OBS** backend for the host `Storage()` domain capability (`storagecap.Provider`).

English | [简体中文](README.zh-CN.md)

## Behavior

- Registers via `storagecap.Provide("linapro-storage-obs", factory)`
- Supports client direct access via `storagecap.DirectAccessProvider` (`presigned_url` for put/get)
- Host `ResolveProvider` selects the unique enabled storage provider plugin; zero → local; multiple → conflict
- Admin settings under **System Settings → Huawei Cloud OBS**
- Required: access key, secret, region, bucket; optional endpoint (default `https://obs.{region}.myhuaweicloud.com`) and path prefix
- When this plugin is the sole active provider but configuration is incomplete, operations fail closed (no silent local fallback)

## Non-goals

- Host file center (`Files()` / `sys_file`) cloud migration
- Client temporary credentials / STS
- Cross-provider migration

## Install

1. Install and enable this plugin (ensure no other storage provider plugin is enabled)
2. Open **System Settings → Huawei Cloud OBS**
3. Save credentials and bucket, then run **Test connection**

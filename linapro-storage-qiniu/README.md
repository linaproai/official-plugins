# linapro-storage-qiniu

Managed source plugin that provides a **Qiniu Kodo** backend for the host `Storage()` domain capability (`storagecap.Provider`).

English | [简体中文](README.zh-CN.md)

## Behavior

- Registers via `storagecap.Provide("linapro-storage-qiniu", factory)`
- Supports client direct access via `storagecap.DirectAccessProvider` (`form_post` for put, private URL for get)
- Host `ResolveProvider` selects the unique enabled storage provider plugin; zero → local; multiple → conflict
- Admin settings under **System Settings → Qiniu Kodo**
- Required: AccessKey, SecretKey, bucket; optional region (`z0`/`z1`/`z2`/`cn-east-2`/`na0`/`as0`, auto-detect when empty); optional download domain and path prefix
- When this plugin is the sole active provider but configuration is incomplete, operations fail closed (no silent local fallback)

## Non-goals

- Host file center (`Files()` / `sys_file`) cloud migration
- Public CDN productization beyond private object Get (private download URLs and form-post upload tokens are supported)
- Cross-provider migration

## Install

1. Install and enable this plugin (ensure no other storage provider plugin is enabled)
2. Open **System Settings → Qiniu Kodo**
3. Save credentials and bucket, then run **Test connection**

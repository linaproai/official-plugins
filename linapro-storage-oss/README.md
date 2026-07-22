# linapro-storage-oss

Managed source plugin that provides a **Alibaba Cloud OSS** backend for the host `Storage()` domain capability (`storagecap.Provider`).

English | [简体中文](README.zh-CN.md)

## Behavior

- Registers via `storagecap.Provide("linapro-storage-oss", factory)`
- Supports client direct access via `storagecap.DirectAccessProvider` (`presigned_url` for put/get)
- Host `ResolveProvider` selects the unique enabled storage provider plugin; zero → local; multiple → conflict
- Configure credentials under **System Settings** in the admin workbench
- Incomplete config fails closed (does not fall back to local disk while this plugin is the unique active provider)

## Out of scope

- Host file-center (`Files()` / `sys_file`) cloud backend
- Client temporary credentials / STS
- Cross-provider migration

## Install

1. Install and enable this plugin (ensure no other storage provider plugin is enabled)
2. Open **System Settings → Alibaba Cloud OSS**
3. Save region, bucket, and credentials; run **Test connection**

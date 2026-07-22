# linapro-storage-aws

Managed source plugin that provides an **official AWS S3** backend for the host `Storage()` domain capability (`storagecap.Provider`).

English | [简体中文](README.zh-CN.md)

## Behavior

- Registers via `storagecap.Provide("linapro-storage-aws", factory)`
- Supports client direct access via `storagecap.DirectAccessProvider` (`presigned_url` for put/get)
- Host `ResolveProvider` selects the unique enabled storage provider plugin; zero → local; multiple → conflict
- Admin settings under **System Settings → AWS S3** (region required; SDK resolves regional endpoints)
- Fail-closed when this plugin is the only active provider but configuration is incomplete

## Non-goals

- Host file center (`Files()` / `sys_file`) cloud offload
- Client temporary credentials / STS
- Cross-provider migration
- Generic S3 protocol endpoints (MinIO, R2, Ceph) — use `linapro-storage-s3`

## Install

1. Install and enable this plugin (ensure no other storage provider plugin is enabled)
2. Open **System Settings → AWS S3**
3. Save region, bucket, and credentials, then **Test connection**

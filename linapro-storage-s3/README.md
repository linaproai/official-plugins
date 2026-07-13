# linapro-storage-s3

Managed source plugin that provides an **S3 protocol** backend for the host `Storage()` domain capability (`storagecap.Provider`).

Use this plugin for MinIO, Cloudflare R2, Ceph RGW, and other S3 API endpoints. For **official AWS S3**, install `linapro-storage-aws` instead.

## Behavior

- Registers via `storagecap.Provide("linapro-storage-s3", factory)`
- Host `ResolveProvider` selects the unique enabled storage provider plugin; zero → local; multiple → conflict
- Admin settings under **System Settings → Storage Management - S3**
- Required: access key, secret, **endpoint**, bucket; optional region (defaults to `us-east-1` for signing); path-style switch
- Fail-closed when this plugin is the only active provider but configuration is incomplete

## Non-goals

- Host file center (`Files()` / `sys_file`) cloud offload
- Presigned URLs
- Cross-provider migration
- AWS-only console UX — use `linapro-storage-aws`

## Install

1. Install and enable this plugin (ensure no other storage provider plugin is enabled)
2. Open **System Settings → Storage Management - S3**
3. Save endpoint, bucket, and credentials, then **Test connection**

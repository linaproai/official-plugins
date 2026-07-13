# linapro-storage-cos

Managed source plugin that provides a **Tencent COS** backend for the host `Storage()` domain capability (`storagecap.Provider`).

## Behavior

- Registers via `storagecap.Provide("linapro-storage-cos", factory)`
- Host `ResolveProvider` selects the unique enabled storage provider plugin; zero → local; multiple → conflict
- Configure credentials under **System Settings** in the admin workbench
- Incomplete config fails closed (does not fall back to local disk while this plugin is the unique active provider)

## Out of scope

- Host file-center (`Files()` / `sys_file`) cloud backend
- Presigned URLs
- Cross-provider migration

## Install

1. Install and enable this plugin (ensure no other storage provider plugin is enabled)
2. Open **System Settings → Tencent COS**
3. Save region, bucket, and credentials; run **Test connection**

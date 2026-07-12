# linapro-storage-azure

Managed source plugin that provides an **Azure Blob Storage** backend for the host `Storage()` domain capability (`storagecap.Provider`).

## Behavior

- Registers via `storagecap.Provide("linapro-storage-azure", factory)`
- Host `ResolveProvider` selects the unique enabled storage provider plugin; zero → local; multiple → conflict
- Admin settings under **Storage → Azure Blob**
- Required: account name, account key, container; optional endpoint (default `https://{account}.blob.core.windows.net/`) and path prefix
- When this plugin is the sole active provider but configuration is incomplete, operations fail closed (no silent local fallback)

## Non-goals

- Host file center (`Files()` / `sys_file`) cloud migration
- Presigned URLs
- Cross-provider migration
- Azure AD / SAS token auth in phase 1 (shared key only)

## Install

1. Install and enable this plugin (ensure no other storage provider plugin is enabled)
2. Open **Storage → Azure Blob**
3. Save account, container, and key, then run **Test connection**

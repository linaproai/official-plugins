# Configuration

## Settings Fields

| Field | Name | Description |
| --- | --- | --- |
| `account_key` | Azure Blob - Account Key | Azure storage account key. Leave blank or masked to keep the existing value. |
| `account_name` | Azure Blob - Account Name | Azure storage account name. |
| `container` | Azure Blob - Container | Azure blob container name. |
| `endpoint` | Azure Blob - Endpoint | Custom service endpoint URL when not using the provider default. |
| `path_prefix` | Azure Blob - Path Prefix | Optional object key prefix applied to all uploaded files. |

## Notes

- Only one object-storage provider should be active at a time.
- When this plugin is the sole active provider but settings are incomplete, storage operations fail closed instead of silently falling back to local disk.

## Entry Points

| Name | Path |
| --- | --- |
| Azure Blob | `linapro-storage-azure-settings` |

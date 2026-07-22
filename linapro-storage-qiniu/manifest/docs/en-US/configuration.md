# Configuration

## Settings Fields

| Field | Name | Description |
| --- | --- | --- |
| `access_key_id` | Qiniu Kodo - Access Key ID | Cloud storage access key identifier used for authenticated requests. |
| `bucket` | Qiniu Kodo - Bucket | Target object storage bucket name. |
| `endpoint` | Qiniu Kodo - Endpoint | Custom service endpoint URL when not using the provider default. |
| `path_prefix` | Qiniu Kodo - Path Prefix | Optional object key prefix applied to all uploaded files. |
| `region` | Qiniu Kodo - Region | Cloud storage region code for the target bucket or container. |
| `secret_access_key` | Qiniu Kodo - Secret Access Key | Cloud storage secret key. Leave blank or masked to keep the existing value. |

## Notes

- Only one object-storage provider should be active at a time.
- When this plugin is the sole active provider but settings are incomplete, storage operations fail closed instead of silently falling back to local disk.

## Entry Points

| Name | Path |
| --- | --- |
| Qiniu Kodo | `linapro-storage-qiniu-settings` |

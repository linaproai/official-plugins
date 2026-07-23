# Configuration

## Settings Fields

| Field | Name | Description |
| --- | --- | --- |
| `access_key_id` | Tencent COS - Access Key ID | Cloud storage access key identifier used for authenticated requests. |
| `bucket` | Tencent COS - Bucket | Target object storage bucket name. |
| `endpoint` | Tencent COS - Endpoint | Custom service endpoint URL when not using the provider default. |
| `path_prefix` | Tencent COS - Path Prefix | Optional object key prefix applied to all uploaded files. |
| `region` | Tencent COS - Region | Cloud storage region code for the target bucket or container. |
| `secret_access_key` | Tencent COS - Secret Access Key | Cloud storage secret key. Leave blank or masked to keep the existing value. |

## Notes

- Only one object-storage provider should be active at a time.
- When this plugin is the sole active provider but settings are incomplete, storage operations fail closed instead of silently falling back to local disk.

## Entry Points

| Name | Path |
| --- | --- |
| Tencent COS | `linapro-storage-cos-settings` |

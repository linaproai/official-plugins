# Configuration

## Settings Fields

| Field | Name | Description |
| --- | --- | --- |
| `access_key_id` | AWS S3 Storage - Access Key ID | Cloud storage access key identifier used for authenticated requests. |
| `bucket` | AWS S3 Storage - Bucket | Target object storage bucket name. |
| `path_prefix` | AWS S3 Storage - Path Prefix | Optional object key prefix applied to all uploaded files. |
| `region` | AWS S3 Storage - Region | Cloud storage region code for the target bucket or container. |
| `secret_access_key` | AWS S3 Storage - Secret Access Key | Cloud storage secret key. Leave blank or masked to keep the existing value. |

## Notes

- Only one object-storage provider should be active at a time.
- When this plugin is the sole active provider but settings are incomplete, storage operations fail closed instead of silently falling back to local disk.

## Entry Points

| Name | Path |
| --- | --- |
| AWS S3 | `linapro-storage-aws-settings` |

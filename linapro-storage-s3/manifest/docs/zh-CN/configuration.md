# 配置说明

## 设置字段

| 字段 | 名称 | 说明 |
| --- | --- | --- |
| `access_key_id` | S3 对象存储 - Access Key ID | 对象存储访问密钥 ID，用于鉴权请求。 |
| `bucket` | S3 对象存储 - 存储桶 | 目标对象存储桶名称。 |
| `endpoint` | S3 对象存储 - Endpoint | 自定义服务接入点；未配置时使用云厂商默认地址。 |
| `force_path_style` | S3 对象存储 - 强制路径风格 | 是否强制 path-style 寻址。可选空/false 或 1/true。 |
| `path_prefix` | S3 对象存储 - 路径前缀 | 上传对象统一附加的可选 key 前缀。 |
| `region` | S3 对象存储 - 区域 | 目标存储桶或容器所在区域编码。 |
| `secret_access_key` | S3 对象存储 - Secret Access Key | 对象存储访问密钥。留空或保持掩码可保留原值。 |

## 说明

- 同一时间只应启用一个对象存储 provider。
- 当本插件是唯一启用 provider 但配置不完整时，存储操作会 fail-closed，不会静默回退本地磁盘。

## 入口

| 名称 | 路径 |
| --- | --- |
| 存储管理-S3 | `linapro-storage-s3-settings` |

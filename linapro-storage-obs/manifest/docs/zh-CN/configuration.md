# 配置说明

## 设置字段

| 字段 | 名称 | 说明 |
| --- | --- | --- |
| `access_key_id` | 华为云 OBS - Access Key ID | 对象存储访问密钥 ID，用于鉴权请求。 |
| `bucket` | 华为云 OBS - 存储桶 | 目标对象存储桶名称。 |
| `endpoint` | 华为云 OBS - Endpoint | 自定义服务接入点；未配置时使用云厂商默认地址。 |
| `path_prefix` | 华为云 OBS - 路径前缀 | 上传对象统一附加的可选 key 前缀。 |
| `region` | 华为云 OBS - 区域 | 目标存储桶或容器所在区域编码。 |
| `secret_access_key` | 华为云 OBS - Secret Access Key | 对象存储访问密钥。留空或保持掩码可保留原值。 |

## 说明

- 同一时间只应启用一个对象存储 provider。
- 当本插件是唯一启用 provider 但配置不完整时，存储操作会 fail-closed，不会静默回退本地磁盘。

## 入口

| 名称 | 路径 |
| --- | --- |
| 华为云 OBS | `linapro-storage-obs-settings` |

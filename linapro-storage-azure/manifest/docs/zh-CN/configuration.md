# 配置说明

## 设置字段

| 字段 | 名称 | 说明 |
| --- | --- | --- |
| `account_key` | Azure Blob - 账户密钥 | Azure 存储账户密钥。留空或保持掩码可保留原值。 |
| `account_name` | Azure Blob - 账户名 | Azure 存储账户名称。 |
| `container` | Azure Blob - 容器 | Azure Blob 容器名称。 |
| `endpoint` | Azure Blob - Endpoint | 自定义服务接入点；未配置时使用云厂商默认地址。 |
| `path_prefix` | Azure Blob - 路径前缀 | 上传对象统一附加的可选 key 前缀。 |

## 说明

- 同一时间只应启用一个对象存储 provider。
- 当本插件是唯一启用 provider 但配置不完整时，存储操作会 fail-closed，不会静默回退本地磁盘。

## 入口

| 名称 | 路径 |
| --- | --- |
| Azure Blob | `linapro-storage-azure-settings` |

# 配置说明

## 设置字段

| 字段 | 名称 | 说明 |
| --- | --- | --- |
| `allow_auto_provision` | 通用 OIDC - 允许自动开户 | 是否允许未知目录用户自动开户。可选 true/false。 |
| `client_id` | 通用 OIDC - Client ID | 身份提供方颁发的 OAuth/OIDC 客户端 ID。 |
| `client_secret` | 通用 OIDC - Client Secret | OAuth/OIDC 客户端密钥。留空或保持掩码可保留原值。 |
| `connection_key` | 通用 OIDC - 连接标识 | 宿主外部登录缝使用的稳定连接标识。 |
| `default_backend_redirect` | 通用 OIDC - 默认后端回跳 | 启用后端回跳时的默认回跳目标。 |
| `display_name` | 通用 OIDC - 显示名称 | 展示给终端用户的登录入口名称。 |
| `issuer` | 通用 OIDC - Issuer | 用于发现与令牌校验的 OIDC Issuer URL。 |
| `redirect_url` | 通用 OIDC - 回调地址 | 在身份提供方注册的 OAuth/OIDC 回调 URL。 |
| `scopes` | 通用 OIDC - Scopes | 授权时请求的 OAuth/OIDC scope，以空格分隔。 |

## 说明

- 身份提供方密钥属于敏感信息，只应通过插件设置页维护。
- 需要保留已有脱敏密钥时，保持对应密钥字段为空。

## 入口

| 名称 | 路径 |
| --- | --- |
| OIDC | `linapro-oidc-generic-settings` |

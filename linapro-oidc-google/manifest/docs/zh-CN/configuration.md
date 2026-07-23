# 配置说明

## 设置字段

| 字段 | 名称 | 说明 |
| --- | --- | --- |
| `allow_auto_provision` | Google OIDC - 允许自动开户 | 是否允许未知目录用户自动开户。可选 true/false。 |
| `backend_redirects` | Google OIDC - 后端回跳列表 | 允许的后端回跳目标列表。 |
| `client_id` | Google OIDC - Client ID | 身份提供方颁发的 OAuth/OIDC 客户端 ID。 |
| `client_secret` | Google OIDC - Client Secret | OAuth/OIDC 客户端密钥。留空或保持掩码可保留原值。 |
| `default_backend_redirect` | Google OIDC - 默认后端回跳 | 启用后端回跳时的默认回跳目标。 |
| `enable_backend_redirect` | Google OIDC - 启用后端回跳 | 是否启用后端回跳白名单。可选 true/false。 |
| `enable_one_tap` | Google OIDC - 启用 One Tap | 是否在登录页启用 Google One Tap。可选 true/false。 |
| `redirect_url` | Google OIDC - 回调地址 | 在身份提供方注册的 OAuth/OIDC 回调 URL。 |

## 说明

- 身份提供方密钥属于敏感信息，只应通过插件设置页维护。
- 需要保留已有脱敏密钥时，保持对应密钥字段为空。

## 入口

| 名称 | 路径 |
| --- | --- |
| Google | `linapro-oidc-google-settings` |

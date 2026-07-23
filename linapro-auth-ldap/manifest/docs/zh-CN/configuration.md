# 配置说明

## 设置字段

| 字段 | 名称 | 说明 |
| --- | --- | --- |
| `allow_auto_provision` | LDAP 登录 - 允许自动开户 | 是否允许未知目录用户自动开户。可选 true/false。 |
| `base_dn` | LDAP 登录 - Base DN | 搜索用户时使用的基础 DN。 |
| `bind_dn` | LDAP 登录 - 绑定 DN | 用于绑定目录的服务账号 DN。 |
| `bind_password` | LDAP 登录 - 绑定密码 | 服务账号密码。留空或保持掩码可保留原值。 |
| `connection_key` | LDAP 登录 - 连接标识 | 宿主外部登录缝使用的稳定连接标识。 |
| `display_name` | LDAP 登录 - 显示名称 | 展示给终端用户的登录入口名称。 |
| `display_name_attr` | LDAP 登录 - 显示名属性 | 读取用户显示名的目录属性。 |
| `email_attr` | LDAP 登录 - 邮箱属性 | 读取用户邮箱的目录属性。 |
| `host` | LDAP 登录 - 主机 | 目录服务器主机名或 IP。 |
| `port` | LDAP 登录 - 端口 | 目录服务器 TCP 端口。 |
| `subject_attr` | LDAP 登录 - 主体属性 | 作为外部主体标识的目录属性。 |
| `tls_mode` | LDAP 登录 - TLS 模式 | LDAP 连接使用的 TLS 模式。 |
| `user_dn_template` | LDAP 登录 - 用户 DN 模板 | 可选的用户 DN 直接构造模板。 |
| `user_filter` | LDAP 登录 - 用户过滤器 | 定位用户条目的 LDAP 过滤模板。 |

## 说明

- 身份提供方密钥属于敏感信息，只应通过插件设置页维护。
- 需要保留已有脱敏密钥时，保持对应密钥字段为空。

## 入口

| 名称 | 路径 |
| --- | --- |
| LDAP | `linapro-auth-ldap-settings` |

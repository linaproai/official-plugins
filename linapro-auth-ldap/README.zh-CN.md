# linapro-auth-ldap

managed 源码插件：通过登录页表单完成 LDAP/AD 目录登录。依赖 `linapro-extid-core`。

**Provider：** `ldap:default` · **自动开户：** 默认关 · **TLS：** LDAPS/StartTLS（明文仅 localhost）

## 安装

1. 启用 `linapro-extid-core`
2. 启用 `linapro-auth-ldap`
3. 配置 host/TLS/搜索或 DN 模板
4. 登录页使用「使用 LDAP 登录」

## 安全

- 密码仅用于 bind，不落库、不进日志
- 失败统一错误语义
- SPA 仅 handoff 兑换会话

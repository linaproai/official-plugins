# linapro-auth-ldap

managed 源码插件：提供**企业 LDAP / Active Directory 目录账号的通用第三方登录能力**。

[English](README.md) | 简体中文

依赖 `linapro-extlogin-core`；同一能力可用于登录入口与其他业务接入场景。

**Provider：** `ldap:default` · **自动开户：** 默认关 · **TLS：** LDAPS/StartTLS（明文仅 localhost）

## 安装

1. 启用 `linapro-extlogin-core`
2. 启用 `linapro-auth-ldap`
3. 配置 host/TLS/搜索或 DN 模板
4. 使用 LDAP 登录入口（例如宿主登录页的「使用 LDAP 登录」）

## 安全

- 密码仅用于 bind，不落库、不进日志
- 失败统一错误语义
- SPA 仅 handoff 兑换会话

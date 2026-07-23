# 第三方登录-LDAP

本文档随插件清单打包，面向插件市场展示。新增文档资源不会改变插件运行时行为。

## 概述

提供企业 LDAP / Active Directory 目录账号的通用第三方登录能力，支持目录绑定认证与可选自动注册，可用于登录入口与其他业务接入场景。依赖「第三方登录-基础框架」。

## 文档导航

- [配置说明](configuration.md)
- [更新日志](changelog.md)

## 功能亮点

- 提供企业 LDAP / Active Directory 目录账号的通用第三方登录能力，支持目录绑定认证与可选自动注册，可用于登录入口与其他业务接入场景。依赖「第三方登录-基础框架」。
- 提供LDAP等工作台入口。
- 依赖`linapro-extlogin-core`。

## 适用场景

用于接入受治理的第三方登录入口、身份绑定和可选自动开户。

## 入口

| 名称 | 路径 |
| --- | --- |
| LDAP | `linapro-auth-ldap-settings` |

## 元数据

| 字段 | 说明 |
| --- | --- |
| 插件 ID | `linapro-auth-ldap` |
| 版本 | `v0.1.0` |
| 类型 | `source` |
| 分发方式 | `managed` |
| 作用域 | `platform_only` |
| 多租户 | 否 |

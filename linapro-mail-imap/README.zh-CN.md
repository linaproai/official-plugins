# 邮件管理-IMAP

`linapro-mail-imap` 从 IMAP 邮箱收取邮件，配合 `linapro-mail-core` 使用。

[English](README.md) | 简体中文

## 能力范围

该插件负责：

- 注册 `kind=imap` 入站传输 SPI
- IMAP 邮箱收取逻辑

Connection/Account 表由 `linapro-mail-core` 拥有，本插件不包含数据表。

## 宿主边界

`linapro-mail-core` 拥有 Connection/Account 持久化、公开 `mailcap` 契约和传输 SPI 解析能力；本插件仅实现 IMAP 传输，并依赖 `linapro-mail-core`。

## 插件元数据

| 字段 | 取值 |
| --- | --- |
| 插件 ID | `linapro-mail-imap` |
| 类型 | `source` |
| 分发治理 | `managed` |
| 作用域 | `platform_only` |
| 安装模式 | `global` |

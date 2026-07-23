# 邮件管理-SMTP

`linapro-mail-smtp` 通过标准 SMTP 服务器发送邮件，配合 `linapro-mail-core` 使用。

[English](README.md) | 简体中文

## 能力范围

该插件负责：

- 注册 `kind=smtp` 出站传输 SPI
- SMTP 连接探测与邮件发送逻辑

Connection/Account 表由 `linapro-mail-core` 拥有，本插件不包含数据表。

## 宿主边界

`linapro-mail-core` 拥有 Connection/Account 持久化、公开 `mailcap` 契约和传输 SPI 解析能力；本插件仅实现 SMTP 传输，并依赖 `linapro-mail-core`。

## 插件元数据

| 字段 | 取值 |
| --- | --- |
| 插件 ID | `linapro-mail-smtp` |
| 类型 | `source` |
| 分发治理 | `managed` |
| 作用域 | `platform_only` |
| 安装模式 | `global` |

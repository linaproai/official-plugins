# linapro-content-notice

`linapro-content-notice` 是 LinaPro 官方提供的通知公告源码插件。

[English](README.md) | 简体中文

## 能力范围

该插件负责：

- 通知与公告的增删改查
- 通知公告相关字典与表结构
- 挂载到宿主 `content` 目录下的默认菜单入口

## 宿主边界

宿主保留 `content` 目录骨架、认证和菜单治理能力；本插件补充通知公告相关的 API、菜单、按钮权限和 `manifest/sql/` 下的插件自有表结构。

## 插件元数据

| 字段 | 取值 |
| --- | --- |
| 插件 ID | `linapro-content-notice` |
| 类型 | `source` |
| 分发治理 | `managed` |
| 作用域 | `tenant_aware` |
| 安装模式 | `tenant_scoped` |

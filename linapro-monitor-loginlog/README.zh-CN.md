# linapro-monitor-loginlog

`linapro-monitor-loginlog` 是 LinaPro 官方提供的登录日志源码插件。

[English](README.md) | 简体中文

## 能力范围

该插件负责：

- 订阅宿主认证生命周期事件并持久化登录日志
- 登录日志查询、导出、清理与详情页面
- 登录日志表结构

## 宿主边界

宿主保留认证生命周期、会话基础设施和菜单治理能力；本插件订阅认证事件，并在 `monitor` 目录下提供查询、导出和清理 API。

## 插件元数据

| 字段 | 取值 |
| --- | --- |
| 插件 ID | `linapro-monitor-loginlog` |
| 类型 | `source` |
| 分发治理 | `managed` |
| 作用域 | `tenant_aware` |
| 安装模式 | `tenant_scoped` |

# linapro-monitor-operlog

`linapro-monitor-operlog` 是 LinaPro 官方提供的操作日志源码插件。

[English](README.md) | 简体中文

## 能力范围

该插件负责：

- 通过插件自有全局中间件采集并持久化操作日志
- 操作日志查询、导出、清理与详情页面
- 操作日志相关字典与表结构

## 宿主边界

宿主保留 HTTP 中间件链、认证和菜单治理能力；本插件注册全局中间件采集操作日志，并在 `monitor` 目录下提供查询、导出和清理 API。

## 插件元数据

| 字段 | 取值 |
| --- | --- |
| 插件 ID | `linapro-monitor-operlog` |
| 类型 | `source` |
| 分发治理 | `managed` |
| 作用域 | `tenant_aware` |
| 安装模式 | `tenant_scoped` |

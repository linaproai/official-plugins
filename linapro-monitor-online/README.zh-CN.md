# linapro-monitor-online

`linapro-monitor-online` 是 LinaPro 官方提供的在线用户治理源码插件。

[English](README.md) | 简体中文

## 能力范围

该插件负责：

- 在线会话投影查询
- 强制下线治理 API 与工作台入口

该插件通过宿主发布的在线会话领域能力访问会话投影，不生成或查询宿主 `sys_*` 表。

## 宿主边界

宿主保留会话基础设施、认证和在线会话领域能力；本插件消费该能力，在 `monitor` 目录下提供查询和治理页面。

## 插件元数据

| 字段 | 取值 |
| --- | --- |
| 插件 ID | `linapro-monitor-online` |
| 类型 | `source` |
| 分发治理 | `managed` |
| 作用域 | `tenant_aware` |
| 安装模式 | `tenant_scoped` |

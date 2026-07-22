# linapro-monitor-server

`linapro-monitor-server` 是 LinaPro 官方提供的服务监控源码插件。

[English](README.md) | 简体中文

## 能力范围

该插件负责：

- 服务监控数据采集
- 过期快照清理
- 服务监控查询 API 与工作台入口

宿主保留 Cron 与插件生命周期底座，监控能力本身由该插件提供。

## 宿主边界

宿主保留调度基础设施和插件生命周期管理；本插件拥有监控数据表结构、采集逻辑和 `monitor` 目录下的查询 API。

## 插件元数据

| 字段 | 取值 |
| --- | --- |
| 插件 ID | `linapro-monitor-server` |
| 类型 | `source` |
| 分发治理 | `managed` |
| 作用域 | `platform_only` |
| 安装模式 | `global` |

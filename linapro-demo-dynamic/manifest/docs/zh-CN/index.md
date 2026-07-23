# 示例插件-动态插件

本文档随插件清单打包，面向插件市场展示。新增文档资源不会改变插件运行时行为。

## 概述

提供独立的 dynamic wasm 插件样例，演示左侧菜单内嵌页面、插件自有 SQL 表 CRUD 与独立静态页面跳转

## 文档导航

- [配置说明](configuration.md)
- [更新日志](changelog.md)

## 功能亮点

- 提供独立的 dynamic wasm 插件样例，演示左侧菜单内嵌页面、插件自有 SQL 表 CRUD 与独立静态页面跳转
- 提供示例插件-动态插件等工作台入口。
- 依赖`LinaPro framework`, `linapro-ai-core`, `linapro-demo-source`。
- 演示动态`WASM`插件、公开静态资源和宿主服务授权。
- 支持租户感知使用场景。

## 适用场景

作为 LinaPro 插件开发和生命周期治理的参考实现。

## 入口

| 名称 | 路径 |
| --- | --- |
| 示例插件-动态插件 | `/x-assets/linapro-demo-dynamic/v0.1.0/mount.js` |

## 元数据

| 字段 | 说明 |
| --- | --- |
| 插件 ID | `linapro-demo-dynamic` |
| 版本 | `v0.1.0` |
| 类型 | `dynamic` |
| 分发方式 | `managed` |
| 作用域 | `tenant_aware` |
| 多租户 | 是 |

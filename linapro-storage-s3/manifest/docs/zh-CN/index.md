# 对象存储-S3

本文档随插件清单打包，面向插件市场展示。新增文档资源不会改变插件运行时行为。

## 概述

为宿主 Storage 领域能力提供 S3 协议对象存储后端（MinIO、R2、Ceph RGW 等）。

## 文档导航

- [配置说明](configuration.md)
- [更新日志](changelog.md)

## 功能亮点

- 为宿主 Storage 领域能力提供 S3 协议对象存储后端（MinIO、R2、Ceph RGW 等）。
- 提供存储管理-S3等工作台入口。

## 适用场景

将本插件作为宿主`Storage()`领域能力背后的对象存储提供方。

## 入口

| 名称 | 路径 |
| --- | --- |
| 存储管理-S3 | `linapro-storage-s3-settings` |

## 元数据

| 字段 | 说明 |
| --- | --- |
| 插件 ID | `linapro-storage-s3` |
| 版本 | `v0.1.0` |
| 类型 | `source` |
| 分发方式 | `managed` |
| 作用域 | `platform_only` |
| 多租户 | 否 |

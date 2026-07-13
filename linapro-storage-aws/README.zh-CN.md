# linapro-storage-aws

托管源码插件，为宿主 `Storage()` 领域能力（`storagecap.Provider`）提供 **官方 AWS S3** 后端。

## 行为

- 通过 `storagecap.Provide("linapro-storage-aws", factory)` 注册
- 宿主 `ResolveProvider` 选择唯一可服务的 storage provider 插件；0 个回退 local；多个冲突拒绝
- 在管理后台 **系统设置 → AWS S3** 配置凭证（region 必填；由 SDK 解析官方地域 endpoint）
- 本插件为唯一 active 但配置不完整时 fail-closed，不会静默回退本地磁盘

## 非目标

- 宿主文件中心（`Files()` / `sys_file`）上云
- 预签名 URL
- 跨 provider 迁移
- 通用 S3 协议端点（MinIO、R2、Ceph 等）— 请使用 `linapro-storage-s3`

## 安装

1. 安装并启用本插件（确保未同时启用其他 storage provider 插件）
2. 打开 **系统设置 → AWS S3**
3. 保存地域、桶与凭证，并执行 **测试连接**

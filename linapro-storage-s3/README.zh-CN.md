# linapro-storage-s3

托管源码插件，为宿主 `Storage()` 领域能力（`storagecap.Provider`）提供 **S3 协议** 对象存储后端。

适用于 MinIO、Cloudflare R2、Ceph RGW 等 S3 API 端点。**官方 AWS S3** 请使用 `linapro-storage-aws`。

## 行为

- 通过 `storagecap.Provide("linapro-storage-s3", factory)` 注册
- 宿主 `ResolveProvider` 选择唯一可服务的 storage provider 插件；0 个回退 local；多个冲突拒绝
- 在管理后台 **系统设置 → 存储管理-S3** 配置
- 必填：访问密钥、**endpoint**、桶；可选 region（签名默认 `us-east-1`）；path-style 开关
- 本插件为唯一 active 但配置不完整时 fail-closed，不会静默回退本地磁盘

## 非目标

- 宿主文件中心（`Files()` / `sys_file`）上云
- 预签名 URL
- 跨 provider 迁移
- 纯 AWS 控制台心智 — 请用 `linapro-storage-aws`

## 安装

1. 安装并启用本插件（确保未同时启用其他 storage provider 插件）
2. 打开 **系统设置 → 存储管理-S3**
3. 保存 endpoint、桶与凭证，并执行 **测试连接**

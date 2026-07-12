# linapro-storage-qiniu

托管源码插件，为宿主 `Storage()` 领域能力（`storagecap.Provider`）提供 **七牛云 Kodo** 对象存储后端。

## 行为

- 通过 `storagecap.Provide("linapro-storage-qiniu", factory)` 注册
- 宿主 `ResolveProvider` 选择唯一可服务的 storage provider 插件；0 个回退 local；多个冲突拒绝
- 在管理后台 **存储管理 → 七牛云 Kodo** 配置
- 必填：AccessKey、SecretKey、Bucket；可选 region（`z0`/`z1`/`z2`/`cn-east-2`/`na0`/`as0`，留空自动探测）；可选下载域名与路径前缀
- 本插件为唯一 active 但配置不完整时 fail-closed，不会静默回退本地磁盘

## 非目标

- 宿主文件中心（`Files()` / `sys_file`）上云
- 面向业务的公开 CDN 预签名产品化（对象 Get 使用私有下载语义）
- 跨 provider 迁移

## 安装

1. 安装并启用本插件（确保未同时启用其他 storage provider 插件）
2. 打开 **存储管理 → 七牛云 Kodo**
3. 保存凭据与存储空间，并执行 **测试连接**

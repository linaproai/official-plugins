# linapro-storage-obs

托管源码插件，为宿主 `Storage()` 领域能力（`storagecap.Provider`）提供 **华为云 OBS** 对象存储后端。

## 行为

- 通过 `storagecap.Provide("linapro-storage-obs", factory)` 注册
- 宿主 `ResolveProvider` 选择唯一可服务的 storage provider 插件；0 个回退 local；多个冲突拒绝
- 在管理后台 **存储管理 → 华为云 OBS** 配置
- 必填：访问密钥、地域、桶；可选 endpoint（默认 `https://obs.{region}.myhuaweicloud.com`）与路径前缀
- 本插件为唯一 active 但配置不完整时 fail-closed，不会静默回退本地磁盘

## 非目标

- 宿主文件中心（`Files()` / `sys_file`）上云
- 预签名 URL
- 跨 provider 迁移

## 安装

1. 安装并启用本插件（确保未同时启用其他 storage provider 插件）
2. 打开 **存储管理 → 华为云 OBS**
3. 保存凭据与桶，并执行 **测试连接**

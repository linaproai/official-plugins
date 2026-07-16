# linapro-storage-cos

托管源码插件，为宿主 `Storage()` 领域能力（`storagecap.Provider`）提供 **腾讯云 COS** 后端。

## 行为

- 通过 `storagecap.Provide("linapro-storage-cos", factory)` 注册
- 通过 `storagecap.DirectAccessProvider` 支持客户端直连访问（put/get 使用 `presigned_url`）
- 宿主 `ResolveProvider` 选择唯一可服务的 storage provider 插件；0 个回退 local；多个冲突拒绝
- 在管理后台 **系统设置** 下配置凭证
- 本插件为唯一 active 但配置不完整时 fail-closed，不会静默回退本地磁盘

## 非目标

- 宿主文件中心（`Files()` / `sys_file`）上云
- 客户端临时凭证 / STS
- 跨 provider 迁移

## 安装

1. 安装并启用本插件（确保未同时启用其他 storage provider 插件）
2. 打开 **系统设置 → 腾讯云 COS**
3. 保存地域、桶与凭证，并执行 **测试连接**

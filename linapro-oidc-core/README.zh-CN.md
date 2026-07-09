# linapro-oidc-core

`linapro-oidc-core`是 LinaPro 官方源码插件，拥有外部身份存储和宿主外部登录 seam 背后的 provider 引擎。

[English](README.md) | 简体中文

## 能力边界

本插件拥有：

- `user_external_identity`链接表，把已验证的`(provider, subject)`键映射到本地用户账号
- `externalidentityspi.Provider`引擎：身份解析、自动开户策略（同邮箱冲突拒绝、无邮箱确定性用户名 anchor、幂等`(provider, subject)`去重），以及经`usercap.ProvisionExternal`委托回宿主用户域的最小权限建号
- 当前用户身份绑定 API：`/x/linapro-oidc-core/api/v1/plugins/linapro-oidc-core/identities`下的列举、绑定与解绑

`linapro-oidc-google`和`linapro-oidc-discord`等 OAuth 协议插件依赖本插件。它们继续负责身份验签并调用宿主`externallogin` seam，路径不变；本插件从不调用`LoginByVerifiedIdentity`，也不声明任何 provider ID 归属。

## 宿主边界

宿主保留 token 铸造、会话持久化、租户解析、预登录令牌交接、登录 IP 策略与认证 hook。本插件只通过`plugin.Providers().ProvideExternalIdentityProvider(...)`注册引擎工厂；宿主 manager 按插件启用状态惰性构造 provider，并把 manager-backed seam 绑定进 auth 服务。

本插件未安装或被禁用时，外部登录 fail-closed：不解析链接、不开户、不签发会话。重新启用后外部登录立即恢复。禁用保留链接表数据；卸载并确认清除数据时删除该表，但绝不级联删除已开户的用户账号，这些账号仍可通过宿主用户管理治理。

## 数据权限边界

外部身份链接是用户自隔离资源：

- 登录解析仅使用权威的`(provider, subject)`部分唯一索引，并返回统一的未开户结果，绝不泄露某邮箱是否存在于其他账号
- 绑定、解绑与列举仅作用于当前会话用户自己的链接；越权目标整体拒绝
- `(provider, subject)`已被其他账号占用时返回冲突错误；重复绑定当前用户已拥有的身份幂等成功

## 开户契约

宿主建号与插件链接写入跨模块边界，无法共享单个数据库事务。正确性收敛在`(provider, subject)`部分唯一索引上：先建号、再写链接；并发开户触发唯一索引冲突时复用胜出链接，而不是冒泡 500。无邮箱身份从身份摘要派生确定性、碰撞抗性的用户名 anchor，重试会复用同一宿主账号。

## 目录结构

```text
linapro-oidc-core/
  plugin.yaml
  plugin_embed.go
  go.mod
  Makefile
  backend/
    plugin.go
    api/identity/v1/              list/bind/unbind DTO
    internal/
      controller/identity/        current-user identity binding handlers
      service/identity/           provider engine and linkage policy
      dao/, model/do/, model/entity/  generated table access objects
  manifest/
    sql/                          install SQL for user_external_identity
    sql/uninstall/                uninstall SQL (drop table on purge)
    i18n/en-US/, i18n/zh-CN/      error and API documentation resources
  hack/config.yaml                plugin-local DAO generation config
```

## 审查清单

- 链接存储、开户策略与绑定逻辑保持在本插件内
- token、会话与租户铸造保持宿主独占
- 引擎工厂通过`ProvideExternalIdentityProvider`声明；不声明 provider ID 归属
- 禁用保留链接数据并使外部登录 fail-closed；卸载清除仅删除插件自有表

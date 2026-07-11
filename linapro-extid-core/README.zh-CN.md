# linapro-extid-core

`linapro-extid-core`是 LinaPro 官方**可装卸**源码插件（`distribution: managed`），作为第三方登录的**外部身份领域** owner：链接存储、开户策略、验证票据、Provider 目录与绑定 API。它不是主框架关键领域；未安装时相关能力 fail-closed，依赖 UI 应隐藏。

[English](README.md) | 简体中文

## 能力边界

本插件拥有：

- `user_external_identity` 链接表（`(provider, subject)` 权威键，支持 subject_kind / app_context / 手机号等扩展快照）
- 宿主 `extidspi.Provider` 引擎：Resolve / Provision / Bind / Unbind / List
- plugin-owned 领域契约 `backend/cap/extidcap`：宽入口 `Service`，聚合 `Ticket` / `Login` / `Linkage` / `Providers` 子面，以及进程内 catalog / handoff 门面
- 当前用户绑定 API：仅接受**已验证 ticket**（禁止客户端自报裸 subject）

`linapro-oidc-google`、`linapro-oidc-discord` 及未来微信/QQ/抖音等协议插件**必须**在 `dependencies.plugins` 中声明依赖本插件。协议插件只做 IdP 验签，调用宿主 `extlogin` + 本插件 ticket/目录；本插件从不调用 `LoginByVerifiedIdentity`，也不声明任何 provider ID 归属。

## 分发与降级

| 状态 | 行为 |
| --- | --- |
| 未安装 / 未启用 | 外部登录 fail-closed；协议插件依赖不满足；登录页第三方入口不展示 |
| 已启用 | 协议插件可完成登录；SPA 经宿主 handoff 交换会话 |
| 禁用 | 保留链接表；登录 fail-closed；再启用后恢复 |
| 卸载（清数据） | 删除插件表；**不**级联删除已开户 `sys_user` |

安装顺序建议：`linapro-extid-core` → 协议插件 → 配置凭证 → 启用。

## 宿主边界

宿主保留 token 铸造、会话、租户、pre-token、登录 IP、认证 hook，以及**一次性 handoff**（协议插件回跳不得把 JWT 放进 URL）。本插件通过 `ProvideExternalIdentityProvider` 注册引擎工厂。

## 数据权限边界

外部身份链接是用户自隔离资源：

- 登录解析仅用 `(provider, subject)`，统一未开户结果，不泄露邮箱是否存在
- 绑定必须消费协议插件签发路径产生的 ticket；解绑/列举仅当前用户

## 目录结构

```text
linapro-extid-core/
  plugin.yaml                 # distribution: managed
  backend/
    cap/extidcap/             # 领域公开契约 + 绑定门面（实现在 internal/service）
    api/identity/v1/          # list / bind-by-ticket / unbind
    internal/service/identity/
  manifest/sql/
```

## 审查清单

- `distribution` 为 `managed`，宿主不强制安装
- 协议插件声明依赖 `linapro-extid-core`
- 绑定仅 ticket；SPA 仅 handoff
- token/session 仍在宿主
- 未安装时密码登录不受影响

## SPA handoff

协议插件登录成功后调用 `extidcap.CreateLoginHandoffFromHost`（owner 插件启动时绑定进程内 store）；SPA 通过本插件公开接口 `POST .../handoff/exchange` 兑换会话。宿主不提供 handoff HTTP。

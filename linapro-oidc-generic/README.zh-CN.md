# linapro-oidc-generic

`linapro-oidc-generic` 是 LinaPro 官方 **managed** 源码插件，提供**可配置的企业 OIDC 通用第三方登录能力**（如 Keycloak、Okta、Azure AD）。它依赖 `linapro-extlogin-core` 完成身份链接、开户策略与 SPA handoff；同一能力可用于登录入口、SSO 投递与其他业务接入场景。

[English](README.md) | 简体中文

## 能力边界

| 本插件拥有 | 不拥有 |
| --- | --- |
| OIDC 授权码 + PKCE、Discovery、JWKS 验 id_token | 链接表 / 开户引擎（`linapro-extlogin-core`） |
| 管理设置（issuer、client、scopes、JIT） | 宿主 JWT / 会话铸造 |
| portal 登录路由 + `auth.login.after` 入口 | 多连接 UI（v1.1）、LDAP、租户级 IdP |

**Provider 编码（v1）：** `oidc:default`，`subject = id_token.sub`。

## 安装顺序

1. 安装并启用 `linapro-extlogin-core`
2. 安装并启用 `linapro-oidc-generic`
3. 在设置页配置 issuer / 客户端凭证
4. 在 IdP 登记回调 URL

## 登录流

1. 点击 OIDC 入口 → `GET /portal/linapro-oidc-generic/login`
2. Discovery → 构造 authorize（PKCE + state + nonce）
3. 回调 → 换票 → 验 `id_token`
4. `LoginByVerifiedIdentity(provider=oidc:default, ...)`
5. SPA 仅带 **handoff**（JWT 不进 URL），由 `linapro-extlogin-core` 兑换会话

## 配置

经宿主 `sys_config` 持久化（首次保存创建；无强制 SQL 种子）：

- Issuer（https；localhost 允许 http）
- Client ID / Secret（读取脱敏）
- 可选回调覆盖（默认按请求 host 推导）
- Scopes（始终包含 `openid`）
- 显示名、工作台落地页
- **自动开户：默认关闭**

## 安全清单

- [x] PKCE S256
- [x] HMAC state + nonce
- [x] JWKS 签名与 iss/aud/exp/sub
- [x] 凭证/issuer 缺失 fail-closed
- [x] SPA 仅 handoff
- [x] Secret 不明文投影

## 审查清单

- `plugin.yaml` 依赖 `linapro-extlogin-core`
- `ProvideExternalIdentity("oidc:default")`
- 本插件无链接表
- `distribution: managed`，非 builtin

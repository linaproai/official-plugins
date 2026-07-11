# linapro-oidc-google

`linapro-oidc-google` 是 `LinaPro` 官方源码插件，提供**基于 Google 账号的通用第三方登录能力**。它提供 OAuth 配置、可选自动注册与 One Tap 嵌入，并依赖 `linapro-extlogin-core` 完成身份绑定与账号开通；同一能力可用于登录入口、账号绑定与其他业务接入场景。

[English](README.md) | 简体中文

## 示例演示内容

- 通过 `plugin.Providers().ProvideExternalIdentity("google")` 在插件初始化阶段声明外部身份 provider 归属，宿主据此拒绝任何声明其他 provider 的 `LoginByVerifiedIdentity` 请求。
- 在插件私有 portal 路径组下注册两个面向浏览器的公开路由：
  - `GET /portal/linapro-oidc-google/login`：构造 Google authorize URL，写入 anti-CSRF 状态 Cookie，并 302 跳转到 Google。
  - `GET /portal/linapro-oidc-google/callback`：校验 state、用授权码换取一个验证过的身份、调用 `Services.Auth().ExternalLogin().LoginByVerifiedIdentity(...)`，最后把宿主返回的 token 或 pre-token 通过 query 参数带回 SPA 登录页。
- 在 `frontend/slots/auth.login.social/google-login-entry.vue` 提供一个 Vue 插槽组件，前端 slot registry 会在插件安装并启用时自动以「其他登录方式」下的平台图标挂载。

OAuth 的 code 兑换和 userinfo 获取是刻意保留的极简 stub 实现：默认的 verifier 直接根据传入的授权码生成一个稳定的身份 DTO，方便在没有真实 Google 项目的情况下把整个流程串起来演练。真实部署必须将 `oauthsvc.NewStubIdentityVerifier()` 替换为基于 HTTP 的 Google OIDC 实现。

## 目录结构

```text
linapro-oidc-google/
  plugin.yaml
  plugin_embed.go
  go.mod
  Makefile
  backend/
    plugin.go
    internal/
      controller/login/           login-start + callback handlers
      service/oauth/              authorize URL 构造、state、verifier、callback
  frontend/
    slots/auth.login.social/      Google 平台图标入口（Tooltip 圆形按钮）
  manifest/
    i18n/en-US/                   英文语言包
    i18n/zh-CN/                   简体中文语言包
  README.md
  README.zh-CN.md
```

## 与新版外部身份接缝的集成

宿主负责租户候选与 Token 发放，并把`(provider, subject)`链接解析与开户策略委托给`linapro-extlogin-core`（本插件已在`plugin.yaml`中声明对它的依赖）。`linapro-extlogin-core`未安装或未启用时，外部登录 fail-closed。本插件只提交“已验证”的身份 DTO，并把宿主返回结果带回 SPA。

| 步骤 | 参与方 | 动作 |
| --- | --- | --- |
| 1. 用户点击“使用 Google 账号登录” | 浏览器 | 整页跳转到 `/portal/linapro-oidc-google/login`。 |
| 2. 插件构造 authorize URL | 插件 | `oauthsvc.BuildAuthorizeURL(ctx, ...)` 返回 Google authorize URL 和新 state；插件写入 state Cookie 后 302 到 Google。 |
| 3. Google 回调 | 浏览器 | Google 重定向回 `/portal/linapro-oidc-google/callback?code=...&state=...`。 |
| 4. 插件校验身份 | 插件 | `oauthsvc.CompleteCallback(ctx, ...)` 校对 state、用 code 换 `{subject, email, displayName}`。 |
| 5. 宿主发放会话 | 宿主 | `Services.Auth().ExternalLogin().LoginByVerifiedIdentity(ctx, extlogin.LoginInput{...})`。 |
| 6. 回到 SPA | 插件 | 插件把 `accessToken`/`refreshToken` 或 `preToken` + 租户候选拼进 query，再重定向回 SPA 登录页。 |

宿主对未绑定本地用户的外部身份返回 `AUTH_EXTERNAL_USER_NOT_PROVISIONED`，插件通过回调 error 重定向把该情况透传给前端。

## 配置

`oauth.DefaultConfig()` 返回带占位客户端凭据的参考配置。真实部署需要在装配阶段覆盖：

```go
cfg := oauthsvc.DefaultConfig()
cfg.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
cfg.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
cfg.RedirectURL = "https://your-host/portal/linapro-oidc-google/callback"
```

生产实现需要把 stub verifier 换成真正走 HTTP 的实现：

1. 调用 `TokenURL` 用 code 换 `id_token`。
2. 用 Google JWKS 校验 `id_token` 签名。
3. 需要时再调用 userinfo 端点补齐 `displayName`。
4. 返回 `oauthsvc.VerifiedIdentity{Subject, Email, DisplayName}`。

## 前端插槽

Vue 插槽组件位于 `frontend/slots/auth.login.social/google-login-entry.vue`。构建期由 slot registry 自动发现；运行期会在插件未安装或已禁用时隐藏图标。点击图标时执行整页跳转，以便浏览器保留插件写入的 Cookie。通用 OIDC / LDAP 等协议登录使用独立的 `auth.login.after` 全宽按钮区域。

## 审查清单

- 已通过 `ProvideExternalIdentity` 声明 provider 归属
- 登录路由挂在插件私有 portal 路径组，未占用 `/x` API 命名空间
- 租户解析与 Token 发放保持宿主拥有；链接存储与开户策略归`linapro-extlogin-core`
- OAuth stub 有清晰的替换指引
- Vue 插槽组件仅在插件安装并启用时才被挂载

## 依赖与回跳

- 依赖 `linapro-extlogin-core`（managed）：请先安装并启用领域插件。
- 登录成功后 SPA 回跳仅携带一次性 `handoff`，由 `linapro-extlogin-core` 公开 API `/x/linapro-extlogin-core/api/v1/handoff/exchange` 兑换会话。

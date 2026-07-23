# linapro-oidc-discord

`linapro-oidc-discord` 是 `LinaPro` 官方源码插件，提供**基于 Discord 账号的通用第三方登录能力**。它提供 OAuth 配置与可选自动注册，并依赖 `linapro-extlogin-core` 完成身份绑定与账号开通；同一能力可用于登录入口、账号绑定与其他业务接入场景。

[English](README.md) | 简体中文

## 工作原理

- 插件启动时通过 `plugin.Providers().ProvideExternalIdentity("discord")` 注册 Discord 身份 provider，宿主会拒绝未注册 provider 的登录请求。
- 在插件私有 portal 路径组下注册两个面向浏览器的公开路由：
  - `GET /portal/linapro-oidc-discord/login`：构造 Discord authorize URL，写入 anti-CSRF 状态 Cookie，并 302 跳转到 Discord。
  - `GET /portal/linapro-oidc-discord/callback`：校验 state、用授权码换取一个验证过的身份、调用 `Services.Auth().ExternalLogin().LoginByVerifiedIdentity(...)`，最后把宿主返回的 token 或 pre-token 通过 query 参数带回 SPA 登录页。
- 在 `frontend/slots/auth.login.social/discord-login-entry.vue` 提供一个 Vue 插槽组件，前端 slot registry 会在插件安装并启用时自动以「其他登录方式」下的平台图标挂载。

当前的 OAuth code 兑换和用户信息获取使用简化的 stub 实现——verifier 直接根据授权码生成身份 DTO，无需真实 Discord 应用即可端到端演练。生产部署需将 `oauthsvc.NewStubIdentityVerifier()` 替换为基于 HTTP 的 Discord OIDC 实现。

## 目录结构

```text
linapro-oidc-discord/
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
    slots/auth.login.social/      Discord 平台图标入口（Tooltip 圆形按钮）
  manifest/
    i18n/en-US/                   英文语言包
    i18n/zh-CN/                   简体中文语言包
  README.md
  README.zh-CN.md
```

## 与新版外部身份接缝的集成

宿主负责租户候选解析与 Token 发放，`(provider, subject)` 链接解析和开户策略委托给 `linapro-extlogin-core`（本插件已在 `plugin.yaml` 中声明对其依赖）。`linapro-extlogin-core` 未安装或未启用时，外部登录功能关闭。本插件只提交已验证的身份 DTO，并将宿主返回结果带回 SPA。

| 步骤 | 参与方 | 动作 |
| --- | --- | --- |
| 1. 用户点击“使用 Discord 账号登录” | 浏览器 | 整页跳转到 `/portal/linapro-oidc-discord/login`。 |
| 2. 插件构造 authorize URL | 插件 | `oauthsvc.BuildAuthorizeURL(ctx, ...)` 返回 Discord authorize URL 和新 state；插件写入 state Cookie 后 302 到 Discord。 |
| 3. Discord 回调 | 浏览器 | Discord 重定向回 `/portal/linapro-oidc-discord/callback?code=...&state=...`。 |
| 4. 插件校验身份 | 插件 | `oauthsvc.CompleteCallback(ctx, ...)` 验证 state、用 code 换取 `{subject, email, displayName}`。 |
| 5. 宿主发放会话 | 宿主 | `Services.Auth().ExternalLogin().LoginByVerifiedIdentity(ctx, extlogin.LoginInput{...})`。 |
| 6. 回到 SPA | 插件 | 插件将 `accessToken`/`refreshToken` 或 `preToken` + 租户候选拼入 query，重定向回 SPA 登录页。 |

宿主对未绑定本地用户的外部身份返回 `AUTH_EXTERNAL_USER_NOT_PROVISIONED`，插件通过回调错误重定向将该信息传递给前端。

## 配置

`oauth.DefaultConfig()` 返回带占位客户端凭据的参考配置。真实部署需要在装配阶段覆盖：

```go
cfg := oauthsvc.DefaultConfig()
cfg.ClientID = os.Getenv("DISCORD_CLIENT_ID")
cfg.ClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
cfg.RedirectURL = "https://your-host/portal/linapro-oidc-discord/callback"
```

生产实现需要把 stub verifier 换成真正走 HTTP 的实现：

1. 调用 `TokenURL`（`https://discord.com/api/oauth2/token`）用 code 换 access token。
2. 用 access token 调用 userinfo 端点（`https://discord.com/api/users/@me`）。
3. 把返回的 `id`、`email`、`global_name`/`username` 投影为 `oauthsvc.VerifiedIdentity`。

## 前端插槽

Vue 插槽组件位于 `frontend/slots/auth.login.social/discord-login-entry.vue`。构建期由 slot registry 自动发现；运行期会在插件未安装或已禁用时隐藏图标。点击图标时执行整页跳转，以便浏览器保留插件写入的 Cookie。通用 OIDC / LDAP 等协议登录使用独立的 `auth.login.after` 全宽按钮区域。

`pluginSlotMeta.order` 设为 `20`，与 `linapro-oidc-google`（order `10`）同时启用时，Google 按钮先渲染，Discord 按钮渲染在其下方。

## 审查清单

- 已通过 `ProvideExternalIdentity` 声明 provider 归属
- 登录路由挂在插件私有 portal 路径组，未占用 `/x` API 命名空间
- 租户解析与 Token 发放保持宿主拥有；链接存储与开户策略归`linapro-extlogin-core`
- OAuth stub 有清晰的替换指引
- Vue 插槽组件仅在插件安装并启用时才被挂载

## 依赖与回跳

- 依赖 `linapro-extlogin-core`（managed）：请先安装并启用领域插件。
- 登录成功后 SPA 回跳仅携带一次性 `handoff`，由 `linapro-extlogin-core` 公开 handoff 兑换 API 兑换会话。

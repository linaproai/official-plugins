# linapro-oidc-discord

`linapro-oidc-discord` 是 `LinaPro` 官方源码插件，用于在登录页接入 Discord 账号登录，提供 OAuth 配置与可选自动注册，并依赖 `linapro-extid-core` 完成身份绑定与账号开通。

[English](README.md) | 简体中文

## 示例演示内容

- 通过 `plugin.Providers().ProvideExternalIdentity("discord")` 在插件初始化阶段声明外部身份 provider 归属，宿主据此拒绝任何声明其他 provider 的 `LoginByVerifiedIdentity` 请求。
- 在插件私有 portal 路径组下注册两个面向浏览器的公开路由：
  - `GET /portal/linapro-oidc-discord/login`：构造 Discord authorize URL，写入 anti-CSRF 状态 Cookie，并 302 跳转到 Discord。
  - `GET /portal/linapro-oidc-discord/callback`：校验 state、用授权码换取一个验证过的身份、调用 `Services.Auth().ExternalLogin().LoginByVerifiedIdentity(...)`，最后把宿主返回的 token 或 pre-token 通过 query 参数带回 SPA 登录页。
- 在 `frontend/slots/auth.login.social/discord-login-entry.vue` 提供一个 Vue 插槽组件，前端 slot registry 会在插件安装并启用时自动以「其他登录方式」下的平台图标挂载。

OAuth 的 code 兑换和 userinfo 获取是刻意保留的极简 stub 实现：默认的 verifier 直接根据传入的授权码生成一个稳定的身份 DTO，方便在没有真实 Discord 应用的情况下把整个流程串起来演练。真实部署必须将 `oauthsvc.NewStubIdentityVerifier()` 替换为基于 HTTP 的 Discord OIDC 实现。

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

宿主负责租户候选与 Token 发放，并把`(provider, subject)`链接解析与开户策略委托给`linapro-extid-core`（本插件已在`plugin.yaml`中声明对它的依赖）。`linapro-extid-core`未安装或未启用时，外部登录 fail-closed。本插件只提交“已验证”的身份 DTO，并把宿主返回结果带回 SPA。

| 步骤 | 参与方 | 动作 |
| --- | --- | --- |
| 1. 用户点击“使用 Discord 账号登录” | 浏览器 | 整页跳转到 `/portal/linapro-oidc-discord/login`。 |
| 2. 插件构造 authorize URL | 插件 | `oauthsvc.BuildAuthorizeURL(ctx, ...)` 返回 Discord authorize URL 和新 state；插件写入 state Cookie 后 302 到 Discord。 |
| 3. Discord 回调 | 浏览器 | Discord 重定向回 `/portal/linapro-oidc-discord/callback?code=...&state=...`。 |
| 4. 插件校验身份 | 插件 | `oauthsvc.CompleteCallback(ctx, ...)` 校对 state、用 code 换 `{subject, email, displayName}`。 |
| 5. 宿主发放会话 | 宿主 | `Services.Auth().ExternalLogin().LoginByVerifiedIdentity(ctx, extlogin.LoginInput{...})`。 |
| 6. 回到 SPA | 插件 | 插件把 `accessToken`/`refreshToken` 或 `preToken` + 租户候选拼进 query，再重定向回 SPA 登录页。 |

宿主对未绑定本地用户的外部身份返回 `AUTH_EXTERNAL_USER_NOT_PROVISIONED`，插件通过回调 error 重定向把该情况透传给前端。

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
- 租户解析与 Token 发放保持宿主拥有；链接存储与开户策略归`linapro-extid-core`
- OAuth stub 有清晰的替换指引
- Vue 插槽组件仅在插件安装并启用时才被挂载

## 依赖与回跳

- 依赖 `linapro-extid-core`（managed）：请先安装并启用领域插件。
- 登录成功后 SPA 回跳仅携带一次性 `handoff`，由 `linapro-extid-core` 公开 handoff 兑换 API 兑换会话。

# linapro-oidc-discord

`linapro-oidc-discord` 是 LinaPro 官方源码插件，用于提供 Discord 登录入口和 OAuth2 provider 元数据。

## Scope

本插件负责：

- 在默认工作台上展示 Discord 登录入口
- 提供 Discord 登录路由所需的 OAuth2 provider 元数据
- 插件自有配置与前端登录入口渲染

## Host Boundary

宿主负责 auth provider 发现、登录页聚合、会话签发和 token 交接。本插件只补充 Discord 相关入口展示与 provider 元数据，不承担宿主会话签发职责。

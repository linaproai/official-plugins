# linapro-ops-demo-guard

`linapro-ops-demo-guard` 是 LinaPro 官方提供的演示环境只读保护源码插件。

[English](README.md) | 简体中文

插件安装并启用后，演示环境自动进入只读模式。如需宿主启动时自动启用，将 `linapro-ops-demo-guard` 加入 `plugin.autoEnable` 列表即可。

## 能力范围

该插件负责：

- 基于`HTTP Method`的环境级演示请求治理
- 在宿主`/*`作用域下拦截整个系统请求链路
- 对宿主与插件写请求进行统一拦截
- 演示模式下登录、token 刷新、租户选择、租户切换与登出最小会话白名单放行

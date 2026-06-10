# linapro-demo-dynamic

`linapro-demo-dynamic` 是 LinaPro 的动态 WASM 插件样例，用来演示一个受治理运行时插件的最小闭环。

## 样例覆盖内容

- 一个在默认管理工作台中渲染的菜单入口
- 一个不依赖宿主 UI 框架的独立静态页面
- 通过动态插件桥执行的后端演示路由
- 通过`pkg/plugin/pluginbridge`受治理访问`runtime`、`storage`、`network`、`data`、`plugins`、`jobs`、`manifest`、`hostConfig`、`org`、`tenant`宿主服务
- 后端 controller 方法会被自动发现为与源码插件一致命名的 `Before*` 前置处理器和 `After*` 通知处理器，并通过运行时日志展示生命周期流程

## 目录结构

```text
linapro-demo-dynamic/
  main.go
  plugin_embed.go
  plugin.yaml
  backend/
  frontend/
  manifest/
```

## 构建方式

构建全部动态插件产物：

```bash
make wasm
```

只构建当前样例：

```bash
make wasm p=linapro-demo-dynamic
```

运行时产物会输出到 `temp/output/linapro-demo-dynamic.wasm`。

## 后端契约

该样例通过`/x/linapro-demo-dynamic/api/v1`暴露受治理路由，并将业务逻辑保留在`backend/internal/service/`中。宿主只强制`/x/{pluginId}`前缀；本示例在`backend/plugin.go`中通过`RegisterPlugin`声明`/api/v1/...`作为自身路由分组。

`backend/api/`下的 API DTO 文件只保留资源本地路径，不负责维护路由分组前缀。后续如需增加新的分组，可新增独立 API 包，例如`backend/api/dynamic/v2`或`backend/api/dynamic/interface/m1`，DTO 仍只写包内资源路径，然后在`RegisterPlugin`中新增绑定，例如`plugin.Routes().Group("/api/v2", "dynamic/v2")`或`plugin.Routes().Group("/interface/m1", "dynamic/interface/m1")`。宿主最终会发布为`/x/linapro-demo-dynamic/api/v2/...`或`/x/linapro-demo-dynamic/interface/m1/...`。

## 公开资源

该样例在`plugin.yaml`中声明了`public_assets`：

```yaml
public_assets:
  - source: frontend/pages
    mount: /
    index: index.html
```

只有匹配该声明的文件会通过`/x-assets/linapro-demo-dynamic/v0.1.0/...`提供访问。访问挂载目录本身时，`index`指定默认文件；未配置时默认使用`index.html`。管理工作台菜单仍使用`system/plugin/dynamic-page`，并把`/x-assets/.../mount.js`地址作为托管资源传入；它不会直接把`/x-assets/...`作为工作台路由本身。

## 宿主服务

该样例在 `plugin.yaml` 中申请了以下宿主服务：

- `runtime`
- `storage`
- `network`
- `data`
- `plugins`
- `jobs`
- `manifest`
- `hostConfig`
- `org`
- `tenant`

这些声明会在插件生命周期流程中由宿主进行审查和授权。

`guest`业务宿主服务 client 从`lina-core/pkg/plugin/pluginbridge`导入。同一包也用于样例桥接文件中的协议 envelope、路由分发、生命周期契约、`Jobs`声明契约和响应 helper。

资源类宿主服务授权仍在`plugin.yaml`中声明，但业务代码使用领域能力接口。例如，样例将`pluginbridge.Storage()`作为`storagecap.Service`使用，并通过`storagecap.PutInput`、`GetInput`、`ListInput`、`DeleteInput`和`StatInput`完成存储操作；storage 协议 DTO 保留在桥接传输内部。

`manifest`宿主服务示例仅授权`config/profile.yaml`和`config/config.yaml`。`/api/v1/manifest-demo`路由会通过`manifest.get`读取这两个打包文件，并在内嵌页面展示返回的 profile 与配置预览，从而完整演示从声明到使用的流程。运行期实际生效插件配置通过`Plugins().Config()`读取，并授权为`plugins.config.get`；SQL 和 i18n 生命周期资源不放入本次`manifest`宿主服务授权示例。

该样例还通过`service: jobs`和`method: jobs.register`声明一个内置定时任务。同一个`RegisterPlugin`函数会通过`plugin.Jobs().Register(...)`声明内置任务；宿主 Jobs 发现期使用`pluginbridge.NewDeclarations()`执行该声明，将结果投影到`Jobs`管理中，并通过声明的`JobHeartbeat`路由执行心跳任务。

## 生命周期日志

动态样例实现了`BeforeInstall`、`AfterInstall`、`BeforeUpgrade`、`AfterUpgrade`、`BeforeDisable`、`AfterDisable`、`BeforeUninstall`、`AfterUninstall`、`BeforeTenantDisable`、`AfterTenantDisable`、`BeforeTenantDelete`、`AfterTenantDelete`、`BeforeInstallModeChange`和`AfterInstallModeChange` controller 方法。`linactl wasm`会自动发现这些方法，并将生命周期契约写入`WASM`产物。每个处理器都会返回`ok=true`，并写入包含操作名称和可用迁移字段的运行时日志。

## 审查要点

- `plugin.yaml` 中的元数据和宿主服务声明清晰可读。
- 前端资源与声明的访问模式一致。
- 构建得到的 WASM 产物可以由源码树稳定复现。
- 后端复杂逻辑保留在 service 组件中，而不是堆在 controller 中。

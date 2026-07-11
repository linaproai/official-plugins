# LinaPro 插件目录

`apps/lina-plugins/`是 LinaPro 的一方插件工作区。

LinaPro 将`apps/lina-core`定位为稳定的全栈框架宿主。宿主保留通用框架能力、治理能力和插件扩展接缝；业务模块、运维页面、演示能力和可选领域能力放在本目录以插件形式交付。这样可以保持宿主可复用，并避免核心契约绑定到某个具体管理工作台页面。

当前工作区同时包含随宿主编译的源码插件，以及一个用于运行时交付参考的动态`WASM`插件示例。

## 插件清单

| 插件 | 类型 | 作用域 | 安装模式 | 能力边界 |
|------|------|--------|----------|----------|
| `linapro-tenant-core` | `source` | `platform_only` | `global` | 租户主体、成员关系、租户解析与租户生命周期治理 |
| `linapro-org-core` | `source` | `tenant_aware` | `global` | 部门管理与岗位管理 |
| `linapro-content-notice` | `source` | `tenant_aware` | `tenant_scoped` | 通知公告管理 |
| `linapro-monitor-online` | `source` | `tenant_aware` | `tenant_scoped` | 在线用户查询与强制下线治理 |
| `linapro-monitor-server` | `source` | `platform_only` | `global` | 服务监控采集、清理与查询 |
| `linapro-monitor-operlog` | `source` | `tenant_aware` | `tenant_scoped` | 操作日志持久化与治理页面 |
| `linapro-monitor-loginlog` | `source` | `tenant_aware` | `tenant_scoped` | 登录日志持久化与治理页面 |
| `linapro-ops-demo-guard` | `source` | `tenant_aware` | `global` | 演示环境只读保护与全局写操作拦截 |
| `linapro-extid-core` | `source` | `platform_only` | `global` | 外部身份链接存储、宿主外部登录 seam 背后的解析/开户引擎，以及当前用户身份绑定/解绑/列举 |
| `linapro-oidc-google` | `source` | `platform_only` | `global` | 登录页 Google 账号登录（OAuth 配置、可选自动注册、One Tap）；依赖 `linapro-extid-core` |
| `linapro-oidc-discord` | `source` | `platform_only` | `global` | 登录页 Discord 账号登录（OAuth 配置、可选自动注册）；依赖 `linapro-extid-core` |
| `linapro-oidc-generic` | `source` | `platform_only` | `global` | 登录页可配置企业 OIDC 登录（Discovery/PKCE、可选自动开户默认关）；依赖 `linapro-extid-core` |
| `linapro-auth-ldap` | `source` | `platform_only` | `global` | 登录页 LDAP/AD 目录登录（弹层 bind、自动开户默认关）；依赖 `linapro-extid-core` |
| `linapro-demo-source` | `source` | `tenant_aware` | `tenant_scoped` | 源码插件菜单页面、公开路由和受保护路由示例 |
| `linapro-demo-dynamic` | `dynamic` | `tenant_aware` | `tenant_scoped` | 动态`WASM`插件示例，演示菜单内嵌页面、插件自有`SQL`表`CRUD`和独立静态页面 |

`linactl`会在插件完整构建时根据插件清单和插件本地`Go`模块自动生成已忽略的`temp/official-plugins`聚合模块和`temp/go.work.plugins`工作区，用于接线随宿主编译的源码插件。`linapro-demo-dynamic`不会进入源码插件聚合，它是`WASM`构建与运行时生命周期流程的参考插件。

## 工作区文件

| 路径 | 用途 |
|------|------|
| `<plugin-id>/hack/config.yaml` | 插件本地工具配置入口，包含代码生成、自定义构建和其他插件自有工具配置 |
| `<plugin-id>/plugin.yaml` | 插件清单，包含元数据、分发治理、菜单、安装模式、`i18n`、资产、依赖和宿主服务声明 |
| `<plugin-id>/Makefile` | 插件本地代码生成包装入口，会引入根目录共享的`hack/makefiles/plugin.codegen.mk`目标片段 |
| `<plugin-id>/README.md` | 插件级英文说明 |
| `<plugin-id>/README.zh-CN.md` | 插件级中文说明 |

## 仓库挂载

主仓库`linapro`会把本工作区作为`Git submodule`挂载到`apps/lina-plugins`：

```bash
git submodule update --init --recursive
```

当前配置的 SSH 远端地址为：

```text
git@github.com:linaproai/official-plugins.git
```

## 插件目录契约

每个插件目录都由插件自己拥有。生命周期资源、前端页面、后端代码、`SQL`资产、`i18n`资源和测试都应保留在`apps/lina-plugins/<plugin-id>/`内。

```text
apps/lina-plugins/<plugin-id>/
  backend/
    api/                  API DTO、路由契约与元数据
    internal/
      controller/         插件请求处理与响应投影
      service/            插件业务编排与领域逻辑
      dao/                插件存在数据库访问时生成的本地 DAO 工件
      model/do/           插件存在数据库访问时生成的本地 DO 工件
      model/entity/       插件存在数据库访问时生成的本地实体工件
    plugin.go             后端注册、路由注册、生命周期入口或动态桥接入口
  frontend/pages/         插件自有页面或公开静态资产
  manifest/
    sql/                  安装 SQL 资产
    sql/mock-data/        可选 mock 或演示 SQL 资产
    sql/uninstall/        可选卸载 SQL 资产
    i18n/<locale>/        插件 i18n 资源
  hack/
    config.yaml           插件本地工具配置入口，包含代码生成和自定义构建配置
    tests/                可选的插件自有 E2E 用例、页面对象和 helper
  go.mod                  插件本地 Go 模块
  Makefile                插件本地代码生成包装入口
  plugin.yaml             插件清单
  plugin_embed.go         嵌入资产注册入口
  README.md               英文说明
  README.zh-CN.md         中文说明
```

插件根目录`Makefile`是根目录共享`hack/makefiles/plugin.codegen.mk`片段的薄包装。共享片段会根据引入它的插件目录推导目标后端目录，因此插件`Makefile`不得硬编码`apps/lina-plugins/<plugin-id>/backend`。在插件目录内执行`make ctrl`或`make dao`会使用该插件根目录的`hack/config.yaml`；如果需要从仓库根目录显式指定插件后端，可执行`make ctrl dir=apps/lina-plugins/<plugin-id>/backend`或`make dao dir=apps/lina-plugins/<plugin-id>/backend`。直接调用`linactl ctrl`和`linactl dao`时，目标选择器同样只支持`dir=<backend-dir>`。

需要自定义构建步骤的插件必须在插件根`hack/config.yaml`的`build.commands`下声明。根目录`make build`会扫描`apps/lina-plugins`下包含`plugin.yaml`的直接插件目录，并在宿主后端编译前执行已配置的构建指令。传入`dir=apps/lina-plugins/<plugin-id>`时只构建该插件。缺少`build.commands`是合法状态，表示插件没有自定义构建步骤。

```yaml
build:
  commands:
    - pnpm --dir "$(PLUGIN_ROOT)/frontend" run build
```

`$(PLUGIN_ROOT)`会展开为插件目录，`$(REPO_ROOT)`会展开为仓库根目录。构建指令从插件根目录执行。

`backend/internal/service/`是插件业务服务的唯一合法目录，禁止创建`backend/service/`。动态插件保持同样的`backend/api/`、`backend/plugin.go`、`backend/internal/controller/`和`backend/internal/service/`结构；桥接文件只负责适配`WASM`与`pluginbridge`协议。`guest`业务能力 client 必须来自`lina-core/pkg/plugin/pluginbridge`，不得从`pluginbridge`根包获取。

## 分发治理

`plugin.yaml`可以声明`distribution`，用于描述宿主如何治理插件生命周期。

| 取值 | 语义 | 生命周期 |
|------|------|----------|
| `managed` | 普通可管理插件。省略`distribution`时默认使用该值。 | 在插件管理中可见，可安装、启用、禁用、升级、卸载，也可通过`plugin.autoEnable`托管启用。 |
| `builtin` | 随宿主编译交付的项目内建源码插件。 | 宿主启动时自动安装、启用和安全升级；普通插件管理写操作会被拒绝。 |

`distribution: builtin`只允许`type: source`插件使用，并且插件必须以相同 ID 注册到源码插件注册表。动态插件不得声明`distribution: builtin`。

## 源码插件

源码插件通过显式注册随`apps/lina-core`编译。它们适合需要随宿主构建交付、但仍不应进入宿主核心领域的一方框架能力。

源码插件开发规则：

1. 创建或更新`apps/lina-plugins/<plugin-id>/`。
2. 在`plugin.yaml`和`manifest/`中维护插件元数据、菜单、页面挂载、生命周期资源、`SQL`资产和`i18n`资产。
3. 后端实现保留在`backend/`下，业务逻辑放在`backend/internal/service/`中。
4. 前端页面放在`frontend/pages/`下，或通过`plugin.yaml`的`public_assets`声明公开资产目录。
5. 保持插件自身`go.mod`和`backend/plugin.go`完整；`linactl`会在插件完整构建时自动发现并聚合源码插件后端包。

## 动态插件

动态插件以运行时托管的`WASM`产物交付。`linapro-demo-dynamic/`是上传、安装、启用、停用、卸载、`hostServices`、公开静态资产和通过受治理宿主服务访问插件自有数据的参考实现。

在仓库根目录构建全部动态插件，或通过`p=<plugin-id>`构建单个插件：

```bash
make wasm
make wasm p=linapro-demo-dynamic
```

动态插件必须在`plugin.yaml`中声明`type: dynamic`，使用`main.go`和`go.mod`作为`guest`构建入口，通过`hostServices`描述运行时能力和资源边界，并从`lina-core/pkg/plugin/pluginbridge`导入 runtime、storage、data、cache、users、notifications、plugins 等业务宿主服务 client。插件本地配置通过`Plugins().Config()`消费，并授权为`plugins.config.get`；通知发送使用`Notifications().Send()`和`notifications.messages.send`；定时任务由`jobs`领域管理，不再通过独立动态`cron`host service 声明。

## 宿主与插件边界

宿主拥有稳定的框架表面和一级目录骨架，例如`dashboard`、`platform`、`org`、`content`、`monitor`、`setting`、`scheduler`、`extension`和`developer`。插件通过`plugin.yaml`的`parent_key`自主选择挂载点；宿主只在同步时解析声明的父级，并拒绝缺失父级以避免孤儿菜单树。

插件自有的数据表、菜单、页面、Hook、定时任务、公开资产和生命周期资源都保留在插件目录内。宿主代码应依赖稳定的插件服务接缝和公开包，而不是硬编码插件专属页面结构或菜单装配细节。

## 路由与公开资产

源码插件的`HTTP`路由由插件后端代码注册。不要在`plugin.yaml`中声明公开路由、门户路由、工作台`API`路由或路由分组。

源码插件`API`应使用`registrar.Routes().APIPrefix()`。该方法返回`/x/{plugin-id}`，其后的路径段由插件自行定义，例如`/api/v1`、`/api/v2`或`/interface/m1`。

`plugin.yaml`的`menus`是管理工作台导航与权限的事实来源。注册`HTTP`路由不会自动创建菜单、权限节点、`OpenAPI`条目或工作台路由元数据。

插件可以通过`plugin.yaml`的`public_assets`声明公开静态资产。宿主从`/x-assets/{plugin-id}/{version}/...`提供已声明资源，并把每个声明视为插件作者的发布边界。

## 测试

插件自有 Playwright 覆盖应放在：

```text
apps/lina-plugins/<plugin-id>/hack/tests/e2e/
apps/lina-plugins/<plugin-id>/hack/tests/pages/
apps/lina-plugins/<plugin-id>/hack/tests/support/
```

通过宿主测试运行器运行单个插件测试范围：

```bash
pnpm -C hack/tests test:module -- plugin:<plugin-id>
```

适用时使用插件本地`api_contract_test.go`和`Go package`测试完成后端契约检查。

## 版本升级

当已安装源码插件提升`plugin.yaml`的`version`后，发现流程不会静默替换宿主当前生效版本。

- 当前生效版本仍固定在`sys_plugin.version`和`release_id`。
- 新发现的更高源码版本会写入一个`prepared`状态的`sys_plugin_release`。
- 宿主继续启动前，必须通过受支持的插件工作区更新流程处理源码插件。
- 如果跳过更新，启动会快速失败并报告仍需处理的插件。

## 参考入口

- `apps/lina-plugins/linapro-demo-source/README.md`
- `apps/lina-plugins/linapro-demo-dynamic/README.md`

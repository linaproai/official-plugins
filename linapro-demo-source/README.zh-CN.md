# linapro-demo-source

`linapro-demo-source` 是 `LinaPro` 的源码插件样例，用来展示一个在仓库内开发、由宿主发现、并通过插件管理页显式安装后挂载到默认工作台的完整插件闭环。

## 目录结构

```text
linapro-demo-source/
  plugin.yaml
  plugin_embed.go
  backend/
    api/
    internal/
      controller/
      service/
      dao/
      model/do/
      model/entity/
    hack/config.yaml
    plugin.go
  frontend/
    pages/
  manifest/
    sql/
    sql/uninstall/
```

## 样例覆盖点

- `manifest/sql/` 下的安装 SQL 会创建插件自有表`plugin_linapro_demo_source_record`
- `manifest/sql/mock-data/`下的`mock`SQL会提供本地样例数据使用的可选演示记录
- `manifest/sql/uninstall/` 下的卸载 SQL 会在用户确认清理存储数据时删除该插件自有表
- `frontend/pages/sidebar-entry.vue` 中的示例页面可以对插件自有表执行增删查改，并支持附件上传与下载
- 插件自有附件对象通过`pluginhost.Services.Storage()`写入，使用`demo-record-files/...`这类插件 logical path
- 禁用插件时仅隐藏菜单和路由，不清理数据表数据和已存储文件
- 卸载插件时宿主会弹窗，让用户选择是否同时清理插件自有数据表数据和存储文件
- 生命周期回调会打印 `BeforeInstall`、`AfterInstall`、`BeforeUpgrade`、`Upgrade`、`AfterUpgrade`、`BeforeDisable`、`AfterDisable`、`BeforeUninstall`、`AfterUninstall`、租户生命周期回调和安装模式回调，便于开发者观察源码插件生命周期流程

## 清单范围

`plugin.yaml` 负责保存插件元数据、菜单声明和按钮权限。页面与 `SQL` 资源仍然通过目录约定发现，而不是在元数据中重复维护。

本示例声明`distribution: managed`，因此仍是通过插件管理显式安装、启用、升级、禁用和卸载的普通插件。只有已注册且属于项目组成部分的源码插件才应使用`distribution: builtin`，由宿主启动流程自动安装、启用和安全升级。

`plugin.yaml`不声明源码插件`HTTP`路由。工作台导航仍来自`menus`，后端路由由插件代码注册。

## 后端接入

- 在 `backend/` 中实现插件后端入口
- 将业务逻辑保留在 `backend/internal/service/` 下
- 插件访问数据库时，将本地 ORM 生成工件维护在 `backend/internal/dao` 与 `backend/internal/model/{do,entity}` 下
- 将插件`API`注册到`registrar.Routes().APIPrefix()`下，该前缀会解析为`/x/linapro-demo-source`；示例插件自行追加`/api/v1`作为自身路由约定
- `g.Meta path`只声明插件内业务相对路径（例如`/ping`、`/summary`、`/records`），不要再嵌套`/plugins/linapro-demo-source/`，宿主前缀已经表达插件边界
- 公开页面、门户、静态资源路由或插件自管 fallback handler 应使用非保留路径，不要放在`/x`下
- 通过宿主构建使用的源码插件注册入口显式接线安装、升级、禁用、卸载、租户和安装模式生命周期回调
- 从`registrar.Services()`显式注入`Storage()`，并用它完成附件保存、下载、替换、删除和可选卸载清理
- 将插件自有清理逻辑保留在插件服务中，便于在卸载`SQL`删除表之前按需清理`Storage()`对象

## 前端接入

- 前端页面会按照仓库约定从插件的 `frontend/` 目录中自动发现
- 示例页保留原有摘要卡片，并新增 `VXE` 表格与弹窗表单来维护示例记录
- 卸载时是否清理数据的选择由宿主插件管理页提供，而不是插件页面自行实现

## 公开资源

源码插件可以在`plugin.yaml`的`public_assets`中声明公开静态资源目录。声明后的文件会通过`/x-assets/{plugin-id}/{version}/...`提供访问，但本样例的工作台页面仍走常规`frontend/pages/`发现流程，不需要宿主托管公开资源。

不要使用`/plugin-assets`，该旧路径不再支持。

## SQL 约定

- 安装 SQL 位于 `manifest/sql/`
- 卸载 SQL 位于 `manifest/sql/uninstall/`
- 安装 SQL 需要具备幂等性，以便在“卸载但保留数据”后重新安装时继续复用原有数据
- 当插件存在自有对象存储时，卸载 SQL 应与插件清理钩子协同工作，确保表数据和`Storage()`对象可一起清理

## 附件存储边界

| 场景 | 样例行为 |
| --- | --- |
| 保存或替换附件 | 后端通过`storagecap.Service.Put`写入对象，业务表只保存 logical path、原始文件名和记录元数据。 |
| 下载附件 | 后端通过`storagecap.Service.Get`读取 logical path 并流式返回 reader，不暴露也不打开宿主文件系统路径。 |
| 删除记录或移除附件 | 后端通过`storagecap.Service.Delete`删除已存储对象。 |
| 卸载清理 | 当宿主卸载策略要求清理存储数据时，生命周期回调接收插件作用域 service，按记录中的租户作用域删除附件路径，并对剩余`demo-record-files/`对象执行有界前缀清理。 |

## 审查要点

- 元数据保持精简且准确
- 宿主接线关系保持显式
- 页面遵循目录约定
- 插件自有 SQL 与宿主 SQL 分离维护
- 禁用仅隐藏能力，不清理插件自有数据
- 卸载同时覆盖“保留数据”和“清理数据”两条生命周期路径

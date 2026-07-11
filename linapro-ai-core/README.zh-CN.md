# LinaPro AI Core 插件

`apps/lina-plugins/linapro-ai-core`是`AI`供应商、供应商模型、能力档位、调用日志和 plugin-owned `AI`领域能力契约的一方源码插件。

该插件刻意保留在`apps/lina-core`之外：`lina-core`拥有插件内核治理、descriptor 注册、依赖检查、owner-aware 动态路由、授权快照、审计和缓存失效；`linapro-ai-core`拥有`AI`公开契约、provider `SPI`、动态 guest SDK、descriptor helper、实现和版本策略。

## 插件元数据

| 字段 | 取值 |
| --- | --- |
| 插件 ID | `linapro-ai-core` |
| 类型 | `source` |
| 分发治理 | `managed` |
| 作用域 | `platform_only` |
| 安装模式 | `global` |
| `i18n` | 已启用`en-US`和`zh-CN` |

## 权属边界

| 边界 | 路径 | 职责 |
| --- | --- | --- |
| `AI`公开消费契约 | `backend/cap/aicap` | 类型化`AI`service、子能力 DTO、命名方法常量、状态投影、公开错误、descriptor 元数据和版本策略 |
| Provider `SPI` | `backend/cap/aicap/spi` | Provider factory 接口、typed factory env、provider 注册 helper 和 descriptor helper |
| 动态 Guest SDK | `backend/cap/aicap/bridge` | owner-aware `hostServices`声明 helper、payload codec、host-call client 和动态错误映射 |
| 实现层 | `backend/internal/service/ai` | provider 路由、模型与档位解析、调用日志、缓存使用和外部 provider adapter |
| 管理 API | `backend/api`和`backend/internal/controller` | 供应商、模型、档位和调用日志管理接口 |
| 资源 | `manifest/sql`和`manifest/i18n` | 插件自有 schema、卸载 SQL、菜单标签、插件标签和本地化错误 |

`backend/cap/aicap`不得 import `backend/internal/**`，不得暴露 DAO、DO、Entity、provider 密钥、私有路由表、私有缓存快照或调用日志内部结构。

## 源码插件消费

消费`AI`能力的源码插件必须从`lina-plugin-linapro-ai-core/backend/cap/aicap`或`aicap/aitext`等子包 import owner 契约。生产代码不得继续把`lina-core/pkg/plugin/capability/aicap`作为 owner 契约，也不得 import 本插件的`backend/internal/**`。

消费方必须在`plugin.yaml`中声明 owner 依赖，并保持`go.mod`依赖一致：

```yaml
dependencies:
  plugins:
    - id: linapro-ai-core
      version: ">=0.1.0"
```

源码调用方应消费`AI().Text()`或注入的`aitext.Service`等类型化 service。普通 Go 契约不得使用弱类型`Invoke(method, payload)`网关。

## 动态插件消费

动态插件通过本插件的 bridge SDK 和 owner-aware `hostServices`声明调用`AI`。SDK 只负责声明、编码和调用通用 host-call 路径；运行期授权、依赖检查、owner 启用状态、审计和分发仍由`lina-core`完成。

```yaml
dependencies:
  plugins:
    - id: linapro-ai-core
      version: ">=0.1.0"

hostServices:
  - service: ai
    owner: linapro-ai-core
    version: v1
    methods:
      - text.generate
      - text.method_status.get
```

动态插件不得把 owner 身份编码进`service`，也不得使用`dependencies.capabilities`或可选依赖申请`AI`能力。

## Provider 与版本策略

`linapro-ai-core`发布`ai.v1`descriptor，仅覆盖当前具备真实运行时路径的方法：`text.generate`、`text.method_status.get`和`ai.methods.status.batch_get`。多模态 DTO 仍保留在`backend/cap/aicap/*`供源码契约和后续 provider 绑定使用，但在真实 provider 路径落地前不会发布到授权 catalog。Provider 注册使用`aicap.ProviderDescriptor`，把 typed 文本 factory 包装为带运行时 invoker 的通用 descriptor，并由 invoker 通过`aicap.Service`分发。`pluginhost`只接收 descriptor，不得提供`ProvideAIText`等`AI`领域专属 facade。

能力族 ID 使用 plugin-owned 形式`plugin.linapro-ai-core.ai.<family>.v1`（例如`plugin.linapro-ai-core.ai.text.v1`），用于状态、审计和契约族标识；它与动态 host service 身份（`owner` + `service=ai` + `version=v1` + method）职责分离。

公开契约发生破坏性变更时，必须提升插件版本并变更 owner descriptor 版本。插件版本和 descriptor 版本共同构成源码依赖检查、动态 manifest 校验、授权快照和升级预览使用的兼容边界。

## 治理说明

每次 owner 能力调用都必须保留调用插件 ID、actor、tenant、执行来源、method 授权、存在时的资源声明和审计上下文。公开响应不得暴露 provider 密钥、内部路由配置、完整 provider 请求或响应体，也不得暴露其他插件的调用日志。

本插件配置了`i18n.enabled: true`，用户可见标签、菜单文案、API 文档源文本、owner 授权展示名称和本地化错误维护在`manifest/i18n/<locale>`中；新增 API 文档资源时维护在`manifest/i18n/<locale>/apidoc`。

`AI`供应商、模型、档位和调用日志表归属本插件的`manifest/sql`。该能力继续演进时，`lina-core`不得新增`AI`provider 业务表。

## 测试

后端包测试应靠近被修改的`backend/cap/aicap`和`backend/internal/service/ai`包。插件自有 E2E 测试放在：

```text
apps/lina-plugins/linapro-ai-core/hack/tests/e2e/
apps/lina-plugins/linapro-ai-core/hack/tests/pages/
apps/lina-plugins/linapro-ai-core/hack/tests/support/
```

通过宿主测试运行器执行插件范围 E2E：

```bash
pnpm -C hack/tests test:module -- plugin:linapro-ai-core
```

# 邮件管理-基础框架

`linapro-mail-core` 是 LinaPro 邮件能力的基础插件：

- Connection / Account 持久化
- 公开 `mailcap` 契约
- Transport SPI 注册与 kind 单例解析
- 协议插件启用冲突的全局 lifecycle 检测
- Connection/Account 管理 API 与页面

协议插件（`linapro-mail-smtp`、`linapro-mail-imap`、`linapro-mail-pop3`）仅实现 SPI 并依赖本插件。宿主 `notify` 的邮件通道必须委托本 owner，不得直连协议插件。

## 元数据

| 字段 | 取值 |
| --- | --- |
| 插件 ID | `linapro-mail-core` |
| 类型 | `source` |
| 分发 | `managed` |
| 作用域 | `platform_only` |
| i18n | `en-US`、`zh-CN` |

## 能力边界

| 边界 | 路径 | 职责 |
| --- | --- | --- |
| 公开契约 | `backend/cap/mailcap` | Send/Probe/Fetch、DTO、错误码 |
| Transport SPI | `backend/cap/mailcap/spi` | 按 kind 注册出站/入站，`Resolve` 单例 |
| 实现 | `backend/internal/service/mail` | Connection/Account、mailcap 实现 |
| 管理 API | `backend/api` + controller | Connection/Account CRUD 与 Probe |
| 资源 | `manifest/sql`、`manifest/i18n` | 表结构、卸载、错误与插件文案 |
| 管理页 | `frontend/pages/index.vue` | 路由 `linapro-mail-core-settings`，挂在宿主「系统设置」 |

## 依赖与 notify 桥接

- 协议插件（`linapro-mail-smtp` / `imap` / `pop3`）硬依赖本插件。
- 宿主 `notify` 邮件通道通过 `notifycap.ProvideEmailDelivery` 进程内桥接本插件，**不**直连 SMTP 插件。
- 同 transport kind 仅允许一个可服务协议插件：本插件 `GlobalBeforeEnable` 否决 + `Resolve` 运行时兜底。

## 卸载与数据清理

Connection / Account 配置保存在插件自有表（`plugin_linapro_mail_core_*`）。

- 卸载时**勾选**「同时清理插件自有存储数据」：执行 `manifest/sql/uninstall`（`DROP TABLE`），重装后为空配置。
- 卸载时**不勾选**：仅停用插件并移除菜单/权限等治理资源，**表数据保留**；重装后管理页仍会回显旧账号配置。

## 影响结论（本变更）

| 域 | 结论 |
| --- | --- |
| i18n | 有：插件 error/plugin/menu/pages 语言包；生命周期 veto reason 键 |
| 数据权限 | 管理 API 走 Auth/Tenancy/Permission；一期 platform_only |
| 缓存 | 无独立业务缓存；无 cache-consistency 新语义 |
| 测试 | SPI/全局冲突/helpers 单测；E2E TC001 API、TC002 壳页 |

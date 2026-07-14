# Mail - Core

`linapro-mail-core` is the foundation plugin for LinaPro email:

- Connection and Account persistence
- Public `mailcap` contracts
- Transport SPI registration and kind singleton resolution
- Global lifecycle conflict detection for transport plugins
- Management APIs and pages for connections/accounts

Protocol plugins (`linapro-mail-smtp`, `linapro-mail-imap`, `linapro-mail-pop3`) implement SPI only and depend on this plugin. Host `notify` email delivery must call this owner, not concrete transport plugins.

## Metadata

| Field | Value |
| --- | --- |
| Plugin ID | `linapro-mail-core` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| i18n | `en-US`, `zh-CN` |

## Capability boundary

| Boundary | Path | Responsibility |
| --- | --- | --- |
| Public contract | `backend/cap/mailcap` | Send/Probe/Fetch, DTOs, error codes |
| Transport SPI | `backend/cap/mailcap/spi` | Kind registration, singleton `Resolve` |
| Implementation | `backend/internal/service/mail` | Connection/Account, mailcap service |
| Management API | `backend/api` + controllers | Connection/Account CRUD and Probe |
| Resources | `manifest/sql`, `manifest/i18n` | Schema, uninstall, labels/errors |
| Management page | `frontend/pages/index.vue` | Route `linapro-mail-core-settings` under host **System Settings** |

## Dependencies and notify bridge

- Protocol plugins (`linapro-mail-smtp` / `imap` / `pop3`) hard-depend on this owner.
- Host `notify` email channel uses `notifycap.ProvideEmailDelivery` process-local bridge; it must not call SMTP plugins directly.
- One serviceable protocol plugin per transport kind: `GlobalBeforeEnable` veto plus runtime `Resolve` safety net.

## Uninstall and data cleanup

Connection/Account settings live in plugin-owned tables (`plugin_linapro_mail_core_*`).

- Uninstall **with** “Also clear plugin-owned storage data”: runs `manifest/sql/uninstall` (`DROP TABLE`); reinstall starts empty.
- Uninstall **without** that option: only deactivates the plugin and removes menus/permissions; **table data is kept**, so reinstall still shows the previous account configuration.

## Impact notes (this change)

| Domain | Assessment |
| --- | --- |
| i18n | Yes: plugin error/plugin/menu/pages packs; lifecycle veto reason keys |
| Data permission | Management APIs use Auth/Tenancy/Permission; platform_only in v1 |
| Cache | No dedicated business cache; no new cache-consistency semantics |
| Testing | SPI/global conflict/helpers unit tests; E2E TC001 API, TC002 shell page |

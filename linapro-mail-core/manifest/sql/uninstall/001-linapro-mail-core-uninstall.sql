-- Uninstall SQL for linapro-mail-core plugin-owned tables.

DROP TABLE IF EXISTS plugin_linapro_mail_core_account;
DROP TABLE IF EXISTS plugin_linapro_mail_core_connection;

-- Drop legacy short-prefix tables if present from early drafts
-- (plugin_linapro_mail_* without the full plugin-id segment "core").
DROP TABLE IF EXISTS plugin_linapro_mail_account;
DROP TABLE IF EXISTS plugin_linapro_mail_connection;

-- Uninstall script for linapro-oidc-discord.
--
-- The plugin stores its configuration in the host's shared sys_config table
-- under the "linapro-oidc-discord." namespace, not in a private SQL table.
-- Uninstalling removes every sys_config row that belongs to this plugin so
-- re-installing produces a clean state without leftover OAuth credentials.

DELETE FROM sys_config
WHERE tenant_id = 0
  AND key LIKE 'linapro-oidc-discord.%';

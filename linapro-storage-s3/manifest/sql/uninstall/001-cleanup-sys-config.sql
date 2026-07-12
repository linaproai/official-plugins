-- Remove plugin-owned sys_config keys on uninstall.
-- 卸载时清理插件拥有的 sys_config 键。
DELETE FROM sys_config
WHERE "key" LIKE 'plugin.linapro-storage-s3.%'
  AND "deleted_at" IS NULL;

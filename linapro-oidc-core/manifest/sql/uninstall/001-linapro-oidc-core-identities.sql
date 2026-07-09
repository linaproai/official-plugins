-- ------------------------------------------------------------
-- 001 linapro-oidc-core external identity linkage uninstall SQL file
-- 001 linapro-oidc-core 外部身份链接卸载 SQL 文件
-- Purpose: Drops the plugin-owned external identity linkage table when the user
-- confirms storage purge during uninstall. Provisioned local user accounts are
-- deliberately NOT cascaded: account lifecycle stays host-owned, and orphaned
-- least-privilege accounts remain manageable through host user management.
-- 用途：在卸载并确认清除数据时删除插件私有外部身份链接表。已开户的本地用户账号
-- 刻意不级联删除：账号生命周期归宿主所有，遗留的最小权限账号仍可通过宿主用户管理治理。
-- ------------------------------------------------------------

DROP TABLE IF EXISTS user_external_identity;

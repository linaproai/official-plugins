-- 001: linapro-oidc-discord plugin settings uninstall
-- 001：linapro-oidc-discord 插件设置卸载
-- Removes the platform-level sys_config rows seeded during install.
-- 删除安装阶段创建的平台级 sys_config 行。

DELETE FROM sys_config
WHERE "tenant_id" = 0
  AND "key" IN (
    'plugin.linapro-oidc-discord.client_id',
    'plugin.linapro-oidc-discord.client_secret',
    'plugin.linapro-oidc-discord.redirect_url',
    'plugin.linapro-oidc-discord.enable_backend_redirect',
    'plugin.linapro-oidc-discord.default_backend_redirect',
    'plugin.linapro-oidc-discord.backend_redirects'
  );

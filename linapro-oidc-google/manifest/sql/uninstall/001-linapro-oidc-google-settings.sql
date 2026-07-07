-- 001: linapro-oidc-google plugin settings uninstall
-- 001：linapro-oidc-google 插件设置卸载
-- Removes the platform-level sys_config rows seeded during install.
-- 删除安装阶段创建的平台级 sys_config 行。

DELETE FROM sys_config
WHERE "tenant_id" = 0
  AND "key" IN (
    'plugin.linapro-oidc-google.client_id',
    'plugin.linapro-oidc-google.client_secret',
    'plugin.linapro-oidc-google.redirect_url',
    'plugin.linapro-oidc-google.enable_backend_redirect',
    'plugin.linapro-oidc-google.default_backend_redirect',
    'plugin.linapro-oidc-google.backend_redirects'
  );

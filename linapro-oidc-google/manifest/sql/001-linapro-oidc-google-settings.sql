-- 001: linapro-oidc-google plugin settings
-- 001：linapro-oidc-google 插件设置
-- Seeds the platform-level sys_config rows the plugin reads at request time.
-- 为插件在请求期读取的平台级 sys_config 行提供初始化数据。

-- ============================================================
-- Plugin settings seed data: Google OIDC client credentials, redirect URL,
-- and SSO token-delivery configuration
-- 插件设置初始化数据：Google OIDC 客户端凭据、回调地址与 SSO 令牌投递配置
-- ============================================================
INSERT INTO sys_config ("tenant_id", "name", "key", "value", "is_builtin", "remark", "created_at", "updated_at") VALUES
(0, 'Google OIDC-Client ID', 'plugin.linapro-oidc-google.client_id', '', 1, 'Google OAuth 2.0 client ID issued by Google Cloud; edit through the Google OIDC settings page.', NOW(), NOW()),
(0, 'Google OIDC-Client Secret', 'plugin.linapro-oidc-google.client_secret', '', 1, 'Google OAuth 2.0 client secret paired with the client ID; the settings page always returns a masked value.', NOW(), NOW()),
(0, 'Google OIDC-Redirect URL', 'plugin.linapro-oidc-google.redirect_url', '', 1, 'Fully-qualified callback URL registered with Google; must resolve to /portal/linapro-oidc-google/callback.', NOW(), NOW()),
(0, 'Google OIDC-Enable SSO Delivery', 'plugin.linapro-oidc-google.enable_backend_redirect', '', 1, 'SSO token-delivery enablement flag; "1" enables delivery to third-party receiver URLs matched by state key.', NOW(), NOW()),
(0, 'Google OIDC-Workspace Landing Path', 'plugin.linapro-oidc-google.default_backend_redirect', '', 1, 'Workspace landing path after a normal external login; empty keeps the host default landing.', NOW(), NOW()),
(0, 'Google OIDC-SSO Delivery Rules', 'plugin.linapro-oidc-google.backend_redirects', '', 1, 'JSON object mapping business state keys to third-party SSO receiver URLs.', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ------------------------------------------------------------
-- 001 linapro-extlogin-core external identity linkage SQL file
-- 001 linapro-extlogin-core 外部身份链接 SQL 文件
-- Purpose: Plugin-owned linkage between a verified third-party identity
-- (provider + immutable subject) and a local user account. Platform-scoped
-- (no tenant_id): binding belongs to the user account; tenant selection happens
-- after login. Soft-deleted rows keep audit history; the partial unique index
-- only covers live rows so unbinding frees (provider, subject) for relink.
-- 用途：插件私有「已验证第三方身份（provider + 不可变 subject）」与本地用户账号的链接表。
-- 平台级（无 tenant_id）：绑定属于用户账号本身，租户在登录后选择。软删除保留审计；
-- 部分唯一索引仅覆盖存活行，解绑后可重绑同一 (provider, subject)。
-- Table naming: plugin_<plugin-id-with-underscores>_<entity>
-- 表名约定：plugin_<plugin-id 下划线化>_<实体>
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS plugin_linapro_extlogin_core_user_external_identity (
    "id"                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id"               INT NOT NULL,
    "provider"              VARCHAR(64) NOT NULL,
    "subject"               VARCHAR(191) NOT NULL,
    "subject_kind"          VARCHAR(32) NOT NULL DEFAULT 'oidc_sub',
    "app_context"           VARCHAR(128) NOT NULL DEFAULT '',
    "plugin_id"             VARCHAR(128) NOT NULL DEFAULT '',
    "email_snapshot"        VARCHAR(191) NOT NULL DEFAULT '',
    "phone_snapshot"        VARCHAR(64) NOT NULL DEFAULT '',
    "display_name_snapshot" VARCHAR(191) NOT NULL DEFAULT '',
    "avatar_url_snapshot"   VARCHAR(512) NOT NULL DEFAULT '',
    "created_at"            TIMESTAMPTZ,
    "updated_at"            TIMESTAMPTZ,
    "deleted_at"            TIMESTAMPTZ
);

COMMENT ON TABLE plugin_linapro_extlogin_core_user_external_identity IS 'Verified external identity to local user linkage owned by linapro-extlogin-core';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."id" IS 'External identity linkage ID';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."user_id" IS 'Linked local user ID';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."provider" IS 'Stable external provider ID owned by the calling plugin, e.g. google, discord';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."subject" IS 'Immutable provider-issued subject identifier, e.g. OIDC sub';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."subject_kind" IS 'Subject classification: oidc_sub, openid, unionid, custom';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."app_context" IS 'Multi-app context such as WeChat appId or Douyin clientKey';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."plugin_id" IS 'Calling plugin ID stamped by the host when the linkage was created';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."email_snapshot" IS 'Email captured at link time for audit only, never used as a resolution key';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."phone_snapshot" IS 'Phone captured at link time for audit only';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."display_name_snapshot" IS 'Display name captured at link time';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."avatar_url_snapshot" IS 'Avatar URL captured at link time';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_linapro_extlogin_core_user_external_identity."deleted_at" IS 'Soft delete time; live rows keep NULL';

-- Authoritative resolution and de-duplication key for live linkages only.
-- Index names are shortened so they stay within PostgreSQL's 63-char identifier limit.
CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_linapro_extlogin_core_identity_provider_subject
    ON plugin_linapro_extlogin_core_user_external_identity ("provider", "subject") WHERE "deleted_at" IS NULL;
-- Supports listing/unbinding all external identities for one user.
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_extlogin_core_identity_user
    ON plugin_linapro_extlogin_core_user_external_identity ("user_id");

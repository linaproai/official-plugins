-- ------------------------------------------------------------
-- 001 linapro-oidc-core external identity linkage SQL file
-- 001 linapro-oidc-core 外部身份链接 SQL 文件
-- Purpose: Stores plugin-owned linkages between a verified third-party identity
-- (provider + immutable subject) and a local user account. The (provider,
-- subject) partial unique index is the authoritative de-duplication anchor for
-- external login resolution and idempotent provisioning. Rows are soft-deleted
-- (deleted_at) for audit; the unique index only covers live rows so unbinding
-- frees the (provider, subject) key for a future relink. The table is
-- platform-scoped (no tenant_id): the identity binding is a property of the
-- user account, and tenant selection happens after login.
-- 用途：存储插件私有的「已验证第三方身份（provider + 不可变 subject）」与本地用户
-- 账号之间的链接。(provider, subject) 部分唯一索引是外部登录解析与幂等开户的权威
-- 去重锚点。行采用软删除（deleted_at）保留审计痕迹；唯一索引仅覆盖存活行，因此解绑
-- 会释放 (provider, subject) 键以便未来重新绑定。表为平台级（无 tenant_id）：身份
-- 绑定属于用户账号本身，租户在登录后选择。
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS user_external_identity (
    "id"             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id"        INT NOT NULL,
    "provider"       VARCHAR(64) NOT NULL,
    "subject"        VARCHAR(191) NOT NULL,
    "plugin_id"      VARCHAR(128) NOT NULL DEFAULT '',
    "email_snapshot" VARCHAR(191) NOT NULL DEFAULT '',
    "created_at"     TIMESTAMPTZ,
    "updated_at"     TIMESTAMPTZ,
    "deleted_at"     TIMESTAMPTZ
);

COMMENT ON TABLE user_external_identity IS 'Verified external identity to local user linkage owned by linapro-oidc-core';
COMMENT ON COLUMN user_external_identity."id" IS 'External identity linkage ID';
COMMENT ON COLUMN user_external_identity."user_id" IS 'Linked local user ID';
COMMENT ON COLUMN user_external_identity."provider" IS 'Stable external provider ID owned by the calling plugin, e.g. google, discord';
COMMENT ON COLUMN user_external_identity."subject" IS 'Immutable provider-issued subject identifier, e.g. OIDC sub';
COMMENT ON COLUMN user_external_identity."plugin_id" IS 'Calling plugin ID stamped by the host when the linkage was created';
COMMENT ON COLUMN user_external_identity."email_snapshot" IS 'Email captured at link time for audit only, never used as a resolution key';
COMMENT ON COLUMN user_external_identity."created_at" IS 'Creation time';
COMMENT ON COLUMN user_external_identity."updated_at" IS 'Update time';
COMMENT ON COLUMN user_external_identity."deleted_at" IS 'Soft delete time; live rows keep NULL';

-- The partial (provider, subject) unique index is the authoritative resolution
-- and de-duplication key. It only covers live rows so a soft-deleted (unbound)
-- linkage never blocks a future relink of the same external identity.
CREATE UNIQUE INDEX IF NOT EXISTS uk_user_external_identity_provider_subject
    ON user_external_identity ("provider", "subject") WHERE "deleted_at" IS NULL;
-- Supports listing/unbinding all external identities for one user.
CREATE INDEX IF NOT EXISTS idx_user_external_identity_user
    ON user_external_identity ("user_id");

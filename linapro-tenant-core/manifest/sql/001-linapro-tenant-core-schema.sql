-- Purpose: Stores tenant master records, including tenant code, display name, lifecycle status, and audit metadata.
-- 用途：存储租户主数据，包括租户编码、展示名称、生命周期状态与审计元数据。
CREATE TABLE IF NOT EXISTS plugin_linapro_tenant_core_tenant (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "code" VARCHAR(64) NOT NULL,
    "name" VARCHAR(128) NOT NULL,
    "status" VARCHAR(32) NOT NULL DEFAULT 'active',
    "remark" VARCHAR(512) NOT NULL DEFAULT '',
    "created_by" BIGINT NOT NULL DEFAULT 0,
    "updated_by" BIGINT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    CONSTRAINT uk_plugin_linapro_tenant_core_tenant_code UNIQUE ("code")
);

CREATE INDEX IF NOT EXISTS idx_plugin_linapro_tenant_core_tenant_status
    ON plugin_linapro_tenant_core_tenant ("status");

-- Purpose: Records which platform users belong to which tenants.
-- 用途：记录平台用户与租户的成员关系。
CREATE TABLE IF NOT EXISTS plugin_linapro_tenant_core_user_membership (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id" BIGINT NOT NULL,
    "tenant_id" BIGINT NOT NULL,
    "status" SMALLINT NOT NULL DEFAULT 1,
    "joined_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" BIGINT NOT NULL DEFAULT 0,
    "updated_by" BIGINT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    CONSTRAINT uk_plugin_linapro_tenant_core_membership_user_tenant UNIQUE ("user_id", "tenant_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_linapro_tenant_core_membership_user_tenant
    ON plugin_linapro_tenant_core_user_membership ("user_id", "tenant_id");
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_tenant_core_membership_tenant
    ON plugin_linapro_tenant_core_user_membership ("tenant_id", "status");
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_tenant_core_membership_user
    ON plugin_linapro_tenant_core_user_membership ("user_id", "status");

-- Purpose: Stores tenant-specific configuration overrides so tenant values can supersede platform defaults.
-- 用途：存储租户级配置覆盖值，使租户配置可以覆盖平台默认配置。
CREATE TABLE IF NOT EXISTS plugin_linapro_tenant_core_config_override (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id" BIGINT NOT NULL,
    "config_key" VARCHAR(128) NOT NULL,
    "config_value" TEXT NOT NULL DEFAULT '',
    "enabled" BOOLEAN NOT NULL DEFAULT TRUE,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    CONSTRAINT uk_plugin_linapro_tenant_core_config_override UNIQUE ("tenant_id", "config_key")
);

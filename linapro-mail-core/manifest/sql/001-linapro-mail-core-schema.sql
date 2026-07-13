-- LinaPro mail-core schema: Connection and Account tables.
-- Table names use plugin id prefix: linapro-mail-core → plugin_linapro_mail_core_*.

-- Purpose: Stores mail protocol connection configurations owned by mail-core.
-- 用途：存储 mail-core 拥有的邮件协议连接配置。
CREATE TABLE IF NOT EXISTS plugin_linapro_mail_core_connection (
    "id"           BIGSERIAL PRIMARY KEY,
    "name"         VARCHAR(128) NOT NULL DEFAULT '',
    "kind"         VARCHAR(32) NOT NULL DEFAULT '',
    "host"         VARCHAR(255) NOT NULL DEFAULT '',
    "port"         INTEGER NOT NULL DEFAULT 0,
    "username"     VARCHAR(255) NOT NULL DEFAULT '',
    "secret_ref"   VARCHAR(256) NOT NULL DEFAULT '',
    "tls_mode"     VARCHAR(32) NOT NULL DEFAULT 'starttls',
    "auth_mode"    VARCHAR(32) NOT NULL DEFAULT 'password',
    "extra_json"   TEXT NOT NULL DEFAULT '{}',
    "status"       SMALLINT NOT NULL DEFAULT 1,
    "tenant_id"    BIGINT NOT NULL DEFAULT 0,
    "remark"       VARCHAR(512) NOT NULL DEFAULT '',
    "created_at"   TIMESTAMPTZ,
    "updated_at"   TIMESTAMPTZ,
    "deleted_at"   TIMESTAMPTZ
);

COMMENT ON TABLE plugin_linapro_mail_core_connection IS 'Mail connection configuration table';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."name" IS 'Connection display name';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."kind" IS 'Transport kind: smtp, imap, pop3';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."host" IS 'Mail server host';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."port" IS 'Mail server port';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."username" IS 'Authentication username';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."secret_ref" IS 'Secret reference for password or token';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."tls_mode" IS 'TLS mode: disable, starttls, tls';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."auth_mode" IS 'Auth mode: password, oauth2';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."extra_json" IS 'Protocol extension JSON without secrets';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."status" IS 'Status: 1=enabled, 0=disabled';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."tenant_id" IS 'Tenant ID; 0 means platform scope';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."remark" IS 'Remark';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_linapro_mail_core_connection."deleted_at" IS 'Deletion time';

CREATE INDEX IF NOT EXISTS idx_plugin_linapro_mail_core_connection_kind_status
    ON plugin_linapro_mail_core_connection ("kind", "status")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_mail_core_connection_tenant
    ON plugin_linapro_mail_core_connection ("tenant_id")
    WHERE "deleted_at" IS NULL;

-- Purpose: Stores business mail account identities bound to connections.
-- 用途：存储绑定到 Connection 的业务邮箱账号身份。
CREATE TABLE IF NOT EXISTS plugin_linapro_mail_core_account (
    "id"                      BIGSERIAL PRIMARY KEY,
    "name"                    VARCHAR(128) NOT NULL DEFAULT '',
    "from_address"            VARCHAR(255) NOT NULL DEFAULT '',
    "outbound_connection_id"  BIGINT NOT NULL DEFAULT 0,
    "inbound_connection_id"   BIGINT NOT NULL DEFAULT 0,
    "is_default"              SMALLINT NOT NULL DEFAULT 0,
    "status"                  SMALLINT NOT NULL DEFAULT 1,
    "tenant_id"               BIGINT NOT NULL DEFAULT 0,
    "remark"                  VARCHAR(512) NOT NULL DEFAULT '',
    "created_at"              TIMESTAMPTZ,
    "updated_at"              TIMESTAMPTZ,
    "deleted_at"              TIMESTAMPTZ
);

COMMENT ON TABLE plugin_linapro_mail_core_account IS 'Mail account identity table';
COMMENT ON COLUMN plugin_linapro_mail_core_account."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_linapro_mail_core_account."name" IS 'Account display name';
COMMENT ON COLUMN plugin_linapro_mail_core_account."from_address" IS 'Default From address';
COMMENT ON COLUMN plugin_linapro_mail_core_account."outbound_connection_id" IS 'Outbound connection ID; 0 means none';
COMMENT ON COLUMN plugin_linapro_mail_core_account."inbound_connection_id" IS 'Inbound connection ID; 0 means none';
COMMENT ON COLUMN plugin_linapro_mail_core_account."is_default" IS 'Default account flag: 1=default, 0=normal';
COMMENT ON COLUMN plugin_linapro_mail_core_account."status" IS 'Status: 1=enabled, 0=disabled';
COMMENT ON COLUMN plugin_linapro_mail_core_account."tenant_id" IS 'Tenant ID; 0 means platform scope';
COMMENT ON COLUMN plugin_linapro_mail_core_account."remark" IS 'Remark';
COMMENT ON COLUMN plugin_linapro_mail_core_account."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_linapro_mail_core_account."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_linapro_mail_core_account."deleted_at" IS 'Deletion time';

CREATE INDEX IF NOT EXISTS idx_plugin_linapro_mail_core_account_default
    ON plugin_linapro_mail_core_account ("tenant_id", "is_default", "status")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_mail_core_account_outbound
    ON plugin_linapro_mail_core_account ("outbound_connection_id")
    WHERE "deleted_at" IS NULL;

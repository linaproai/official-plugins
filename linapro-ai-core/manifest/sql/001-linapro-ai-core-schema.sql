-- LinaPro AI core plugin schema.
-- Creates provider, model, tier, binding, and invocation tables owned by linapro-ai-core.

CREATE TABLE IF NOT EXISTS plugin_linapro_ai_provider (
    "id"                 BIGSERIAL PRIMARY KEY,
    "name"               VARCHAR(128) NOT NULL,
    "website_url"        VARCHAR(512) NOT NULL DEFAULT '',
    "remark"             VARCHAR(512) NOT NULL DEFAULT '',
    "openai_base_url"    VARCHAR(512) NOT NULL DEFAULT '',
    "anthropic_base_url" VARCHAR(512) NOT NULL DEFAULT '',
    "api_key_secret_ref" VARCHAR(256) NOT NULL DEFAULT '',
    "enabled"            SMALLINT NOT NULL DEFAULT 1,
    "created_at"         TIMESTAMP,
    "updated_at"         TIMESTAMP,
    "deleted_at"         TIMESTAMP
);

COMMENT ON TABLE plugin_linapro_ai_provider IS 'AI provider table';
COMMENT ON COLUMN plugin_linapro_ai_provider."id" IS 'Provider ID';
COMMENT ON COLUMN plugin_linapro_ai_provider."name" IS 'Provider display name';
COMMENT ON COLUMN plugin_linapro_ai_provider."website_url" IS 'Provider website URL';
COMMENT ON COLUMN plugin_linapro_ai_provider."remark" IS 'Provider remark';
COMMENT ON COLUMN plugin_linapro_ai_provider."openai_base_url" IS 'OpenAI-compatible base URL';
COMMENT ON COLUMN plugin_linapro_ai_provider."anthropic_base_url" IS 'Anthropic-compatible base URL';
COMMENT ON COLUMN plugin_linapro_ai_provider."api_key_secret_ref" IS 'API key secret reference or masked secret reference';
COMMENT ON COLUMN plugin_linapro_ai_provider."enabled" IS 'Enabled flag: 0=disabled 1=enabled';
COMMENT ON COLUMN plugin_linapro_ai_provider."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_linapro_ai_provider."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_linapro_ai_provider."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS idx_plugin_linapro_ai_provider_name_alive
    ON plugin_linapro_ai_provider ("name")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_provider_enabled_alive
    ON plugin_linapro_ai_provider ("enabled")
    WHERE "deleted_at" IS NULL;

CREATE TABLE IF NOT EXISTS plugin_linapro_ai_model (
    "id"                BIGSERIAL PRIMARY KEY,
    "provider_id"       BIGINT NOT NULL,
    "capability_type"   VARCHAR(32) NOT NULL DEFAULT 'text',
    "model_name"        VARCHAR(128) NOT NULL,
    "protocol"          VARCHAR(32) NOT NULL,
    "source"            VARCHAR(32) NOT NULL DEFAULT 'manual',
    "supports_thinking" SMALLINT NOT NULL DEFAULT 0,
    "supported_efforts" VARCHAR(128) NOT NULL DEFAULT '',
    "max_input_tokens"  INTEGER NOT NULL DEFAULT 0,
    "max_output_tokens" INTEGER NOT NULL DEFAULT 0,
    "enabled"           SMALLINT NOT NULL DEFAULT 1,
    "created_at"        TIMESTAMP,
    "updated_at"        TIMESTAMP,
    "deleted_at"        TIMESTAMP
);

COMMENT ON TABLE plugin_linapro_ai_model IS 'AI provider model table';
COMMENT ON COLUMN plugin_linapro_ai_model."id" IS 'Model ID';
COMMENT ON COLUMN plugin_linapro_ai_model."provider_id" IS 'Provider ID';
COMMENT ON COLUMN plugin_linapro_ai_model."capability_type" IS 'Capability type: text';
COMMENT ON COLUMN plugin_linapro_ai_model."model_name" IS 'Provider model name';
COMMENT ON COLUMN plugin_linapro_ai_model."protocol" IS 'Protocol: openai or anthropic';
COMMENT ON COLUMN plugin_linapro_ai_model."source" IS 'Model source: manual or api';
COMMENT ON COLUMN plugin_linapro_ai_model."supports_thinking" IS 'Thinking effort support flag: 0=no 1=yes';
COMMENT ON COLUMN plugin_linapro_ai_model."supported_efforts" IS 'Comma-separated supported thinking efforts';
COMMENT ON COLUMN plugin_linapro_ai_model."max_input_tokens" IS 'Maximum input tokens, 0 means unspecified';
COMMENT ON COLUMN plugin_linapro_ai_model."max_output_tokens" IS 'Maximum output tokens, 0 means unspecified';
COMMENT ON COLUMN plugin_linapro_ai_model."enabled" IS 'Enabled flag: 0=disabled 1=enabled';
COMMENT ON COLUMN plugin_linapro_ai_model."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_linapro_ai_model."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_linapro_ai_model."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS idx_plugin_linapro_ai_model_identity_alive
    ON plugin_linapro_ai_model ("provider_id", "capability_type", "protocol", "model_name")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_model_provider_alive
    ON plugin_linapro_ai_model ("provider_id", "enabled", "capability_type")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_model_capability_alive
    ON plugin_linapro_ai_model ("capability_type", "enabled")
    WHERE "deleted_at" IS NULL;

CREATE TABLE IF NOT EXISTS plugin_linapro_ai_tier (
    "id"                       BIGSERIAL PRIMARY KEY,
    "capability_type"          VARCHAR(32) NOT NULL DEFAULT 'text',
    "code"                     VARCHAR(32) NOT NULL,
    "display_name"             VARCHAR(64) NOT NULL,
    "description"              VARCHAR(512) NOT NULL DEFAULT '',
    "default_effort"           VARCHAR(32) NOT NULL DEFAULT '',
    "enabled"                  SMALLINT NOT NULL DEFAULT 1,
    "sort_order"               INTEGER NOT NULL DEFAULT 0,
    "last_test_status"         VARCHAR(32) NOT NULL DEFAULT '',
    "last_test_latency_ms"     INTEGER NOT NULL DEFAULT 0,
    "last_test_error_summary"  VARCHAR(512) NOT NULL DEFAULT '',
    "last_test_at"             TIMESTAMP,
    "created_at"               TIMESTAMP,
    "updated_at"               TIMESTAMP
);

COMMENT ON TABLE plugin_linapro_ai_tier IS 'AI capability tier table';
COMMENT ON COLUMN plugin_linapro_ai_tier."id" IS 'Tier ID';
COMMENT ON COLUMN plugin_linapro_ai_tier."capability_type" IS 'Capability type: text';
COMMENT ON COLUMN plugin_linapro_ai_tier."code" IS 'Tier code: basic, standard, advanced';
COMMENT ON COLUMN plugin_linapro_ai_tier."display_name" IS 'Tier display name';
COMMENT ON COLUMN plugin_linapro_ai_tier."description" IS 'Tier description';
COMMENT ON COLUMN plugin_linapro_ai_tier."default_effort" IS 'Default thinking effort';
COMMENT ON COLUMN plugin_linapro_ai_tier."enabled" IS 'Enabled flag: 0=disabled 1=enabled';
COMMENT ON COLUMN plugin_linapro_ai_tier."sort_order" IS 'Stable sort order';
COMMENT ON COLUMN plugin_linapro_ai_tier."last_test_status" IS 'Last tier test status';
COMMENT ON COLUMN plugin_linapro_ai_tier."last_test_latency_ms" IS 'Last tier test latency in milliseconds';
COMMENT ON COLUMN plugin_linapro_ai_tier."last_test_error_summary" IS 'Last tier test masked error summary';
COMMENT ON COLUMN plugin_linapro_ai_tier."last_test_at" IS 'Last tier test time';
COMMENT ON COLUMN plugin_linapro_ai_tier."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_linapro_ai_tier."updated_at" IS 'Update time';

CREATE UNIQUE INDEX IF NOT EXISTS idx_plugin_linapro_ai_tier_capability_code
    ON plugin_linapro_ai_tier ("capability_type", "code");
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_tier_sort
    ON plugin_linapro_ai_tier ("capability_type", "sort_order");

CREATE TABLE IF NOT EXISTS plugin_linapro_ai_tier_binding (
    "id"          BIGSERIAL PRIMARY KEY,
    "tier_id"     BIGINT NOT NULL,
    "provider_id" BIGINT NOT NULL,
    "model_id"    BIGINT NOT NULL,
    "priority"    INTEGER NOT NULL DEFAULT 0,
    "enabled"     SMALLINT NOT NULL DEFAULT 1,
    "created_at"  TIMESTAMP,
    "updated_at"  TIMESTAMP,
    "deleted_at"  TIMESTAMP
);

COMMENT ON TABLE plugin_linapro_ai_tier_binding IS 'AI tier provider-model binding table';
COMMENT ON COLUMN plugin_linapro_ai_tier_binding."id" IS 'Binding ID';
COMMENT ON COLUMN plugin_linapro_ai_tier_binding."tier_id" IS 'Tier ID';
COMMENT ON COLUMN plugin_linapro_ai_tier_binding."provider_id" IS 'Provider ID';
COMMENT ON COLUMN plugin_linapro_ai_tier_binding."model_id" IS 'Model ID';
COMMENT ON COLUMN plugin_linapro_ai_tier_binding."priority" IS 'Binding priority, 0 is primary';
COMMENT ON COLUMN plugin_linapro_ai_tier_binding."enabled" IS 'Enabled flag: 0=disabled 1=enabled';
COMMENT ON COLUMN plugin_linapro_ai_tier_binding."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_linapro_ai_tier_binding."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_linapro_ai_tier_binding."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS idx_plugin_linapro_ai_tier_binding_primary_alive
    ON plugin_linapro_ai_tier_binding ("tier_id", "priority")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_tier_binding_model_alive
    ON plugin_linapro_ai_tier_binding ("model_id", "enabled")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_tier_binding_provider_alive
    ON plugin_linapro_ai_tier_binding ("provider_id", "enabled")
    WHERE "deleted_at" IS NULL;

CREATE TABLE IF NOT EXISTS plugin_linapro_ai_invocation (
    "id"               BIGSERIAL PRIMARY KEY,
    "request_id"       VARCHAR(64) NOT NULL DEFAULT '',
    "capability_type"  VARCHAR(32) NOT NULL DEFAULT 'text',
    "purpose"          VARCHAR(128) NOT NULL DEFAULT '',
    "tier_code"        VARCHAR(32) NOT NULL DEFAULT '',
    "source_plugin_id" VARCHAR(128) NOT NULL DEFAULT '',
    "tenant_id"        INTEGER NOT NULL DEFAULT 0,
    "user_id"          INTEGER NOT NULL DEFAULT 0,
    "provider_id"      BIGINT NOT NULL DEFAULT 0,
    "model_id"         BIGINT NOT NULL DEFAULT 0,
    "provider_name"    VARCHAR(128) NOT NULL DEFAULT '',
    "model_name"       VARCHAR(128) NOT NULL DEFAULT '',
    "protocol"         VARCHAR(32) NOT NULL DEFAULT '',
    "thinking_effort"  VARCHAR(32) NOT NULL DEFAULT '',
    "status"           VARCHAR(32) NOT NULL DEFAULT '',
    "input_tokens"     INTEGER NOT NULL DEFAULT 0,
    "output_tokens"    INTEGER NOT NULL DEFAULT 0,
    "latency_ms"       INTEGER NOT NULL DEFAULT 0,
    "error_code"       VARCHAR(128) NOT NULL DEFAULT '',
    "error_summary"    VARCHAR(512) NOT NULL DEFAULT '',
    "created_at"       TIMESTAMP
);

COMMENT ON TABLE plugin_linapro_ai_invocation IS 'AI invocation audit log table';
COMMENT ON COLUMN plugin_linapro_ai_invocation."id" IS 'Invocation ID';
COMMENT ON COLUMN plugin_linapro_ai_invocation."request_id" IS 'Request correlation ID';
COMMENT ON COLUMN plugin_linapro_ai_invocation."capability_type" IS 'Capability type';
COMMENT ON COLUMN plugin_linapro_ai_invocation."purpose" IS 'Governed AI purpose';
COMMENT ON COLUMN plugin_linapro_ai_invocation."tier_code" IS 'Tier code';
COMMENT ON COLUMN plugin_linapro_ai_invocation."source_plugin_id" IS 'Source plugin ID';
COMMENT ON COLUMN plugin_linapro_ai_invocation."tenant_id" IS 'Tenant ID';
COMMENT ON COLUMN plugin_linapro_ai_invocation."user_id" IS 'User ID';
COMMENT ON COLUMN plugin_linapro_ai_invocation."provider_id" IS 'Provider ID';
COMMENT ON COLUMN plugin_linapro_ai_invocation."model_id" IS 'Model ID';
COMMENT ON COLUMN plugin_linapro_ai_invocation."provider_name" IS 'Provider display name snapshot';
COMMENT ON COLUMN plugin_linapro_ai_invocation."model_name" IS 'Model name snapshot';
COMMENT ON COLUMN plugin_linapro_ai_invocation."protocol" IS 'Protocol snapshot';
COMMENT ON COLUMN plugin_linapro_ai_invocation."thinking_effort" IS 'Requested or applied thinking effort';
COMMENT ON COLUMN plugin_linapro_ai_invocation."status" IS 'Invocation status: success or failed';
COMMENT ON COLUMN plugin_linapro_ai_invocation."input_tokens" IS 'Input token count';
COMMENT ON COLUMN plugin_linapro_ai_invocation."output_tokens" IS 'Output token count';
COMMENT ON COLUMN plugin_linapro_ai_invocation."latency_ms" IS 'Provider call latency in milliseconds';
COMMENT ON COLUMN plugin_linapro_ai_invocation."error_code" IS 'Stable error code';
COMMENT ON COLUMN plugin_linapro_ai_invocation."error_summary" IS 'Masked error summary';
COMMENT ON COLUMN plugin_linapro_ai_invocation."created_at" IS 'Creation time';

CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_invocation_created
    ON plugin_linapro_ai_invocation ("created_at" DESC);
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_invocation_filters
    ON plugin_linapro_ai_invocation ("capability_type", "tier_code", "status", "created_at" DESC);
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_invocation_provider_model
    ON plugin_linapro_ai_invocation ("provider_id", "model_id", "created_at" DESC);
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_invocation_purpose
    ON plugin_linapro_ai_invocation ("purpose", "created_at" DESC);
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_invocation_source_plugin
    ON plugin_linapro_ai_invocation ("source_plugin_id", "created_at" DESC);

INSERT INTO plugin_linapro_ai_tier (
    "capability_type",
    "code",
    "display_name",
    "description",
    "default_effort",
    "enabled",
    "sort_order",
    "created_at",
    "updated_at"
) VALUES
    ('text', 'basic', 'Basic', 'Low-cost AI capability tier for simple text generation and commit message generation.', 'low', 1, 1, NOW(), NOW()),
    ('text', 'standard', 'Standard', 'Default AI capability tier for regular code generation and code optimization.', 'medium', 1, 2, NOW(), NOW()),
    ('text', 'advanced', 'Advanced', 'High-capability AI tier for complex code generation and cross-file reasoning.', 'high', 1, 3, NOW(), NOW())
ON CONFLICT ("capability_type", "code") DO NOTHING;

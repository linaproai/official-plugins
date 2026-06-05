-- LinaPro AI core plugin schema.
-- Creates the final Smart Center provider, model, capability, tier, invocation, and operation tables.

-- Purpose: Stores AI provider metadata for Smart Center.
-- 用途：存储智能中心的 AI 渠道元数据。
CREATE TABLE IF NOT EXISTS plugin_linapro_ai_provider (
    "id"          BIGSERIAL PRIMARY KEY,
    "name"        VARCHAR(128) NOT NULL,
    "website_url" VARCHAR(512) NOT NULL DEFAULT '',
    "remark"      VARCHAR(512) NOT NULL DEFAULT '',
    "enabled"     SMALLINT NOT NULL DEFAULT 1,
    "created_at"  TIMESTAMP,
    "updated_at"  TIMESTAMP,
    "deleted_at"  TIMESTAMP
);

COMMENT ON TABLE plugin_linapro_ai_provider IS 'AI provider table';
COMMENT ON COLUMN plugin_linapro_ai_provider."id" IS 'Provider ID';
COMMENT ON COLUMN plugin_linapro_ai_provider."name" IS 'Provider display name';
COMMENT ON COLUMN plugin_linapro_ai_provider."website_url" IS 'Provider website URL';
COMMENT ON COLUMN plugin_linapro_ai_provider."remark" IS 'Provider remark';
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

-- Purpose: Stores extensible protocol endpoints owned by one AI provider.
-- 用途：存储单个 AI 渠道拥有的可扩展协议端点。
CREATE TABLE IF NOT EXISTS plugin_linapro_ai_provider_endpoint (
    "id"            BIGSERIAL PRIMARY KEY,
    "provider_id"   BIGINT NOT NULL,
    "protocol"      VARCHAR(64) NOT NULL,
    "base_url"      VARCHAR(512) NOT NULL DEFAULT '',
    "secret_ref"    VARCHAR(256) NOT NULL DEFAULT '',
    "enabled"       SMALLINT NOT NULL DEFAULT 1,
    "metadata_json" TEXT NOT NULL DEFAULT '{}',
    "created_at"    TIMESTAMP,
    "updated_at"    TIMESTAMP,
    "deleted_at"    TIMESTAMP
);

COMMENT ON TABLE plugin_linapro_ai_provider_endpoint IS 'AI provider protocol endpoint table';
COMMENT ON COLUMN plugin_linapro_ai_provider_endpoint."id" IS 'Provider endpoint ID';
COMMENT ON COLUMN plugin_linapro_ai_provider_endpoint."provider_id" IS 'Provider ID';
COMMENT ON COLUMN plugin_linapro_ai_provider_endpoint."protocol" IS 'Protocol: openai, anthropic, voyage, openai-compatible, or provider-specific';
COMMENT ON COLUMN plugin_linapro_ai_provider_endpoint."base_url" IS 'Provider protocol base URL';
COMMENT ON COLUMN plugin_linapro_ai_provider_endpoint."secret_ref" IS 'Secret reference or masked secret reference';
COMMENT ON COLUMN plugin_linapro_ai_provider_endpoint."enabled" IS 'Enabled flag: 0=disabled 1=enabled';
COMMENT ON COLUMN plugin_linapro_ai_provider_endpoint."metadata_json" IS 'Endpoint metadata JSON without secret values';
COMMENT ON COLUMN plugin_linapro_ai_provider_endpoint."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_linapro_ai_provider_endpoint."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_linapro_ai_provider_endpoint."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS idx_plugin_linapro_ai_endpoint_identity_alive
    ON plugin_linapro_ai_provider_endpoint ("provider_id", "protocol", "base_url")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_endpoint_provider_alive
    ON plugin_linapro_ai_provider_endpoint ("provider_id", "enabled", "protocol")
    WHERE "deleted_at" IS NULL;

-- Purpose: Stores provider model identity metadata for Smart Center binding.
-- 用途：存储用于智能中心绑定的渠道模型身份元数据。
CREATE TABLE IF NOT EXISTS plugin_linapro_ai_model (
    "id"          BIGSERIAL PRIMARY KEY,
    "provider_id" BIGINT NOT NULL,
    "endpoint_id" BIGINT NOT NULL,
    "model_name"  VARCHAR(128) NOT NULL,
    "protocol"    VARCHAR(32) NOT NULL,
    "source"      VARCHAR(32) NOT NULL DEFAULT 'manual',
    "enabled"     SMALLINT NOT NULL DEFAULT 1,
    "created_at"  TIMESTAMP,
    "updated_at"  TIMESTAMP,
    "deleted_at"  TIMESTAMP
);

COMMENT ON TABLE plugin_linapro_ai_model IS 'AI provider model table';
COMMENT ON COLUMN plugin_linapro_ai_model."id" IS 'Model ID';
COMMENT ON COLUMN plugin_linapro_ai_model."provider_id" IS 'Provider ID';
COMMENT ON COLUMN plugin_linapro_ai_model."endpoint_id" IS 'Provider endpoint ID used by the model';
COMMENT ON COLUMN plugin_linapro_ai_model."model_name" IS 'Provider model name';
COMMENT ON COLUMN plugin_linapro_ai_model."protocol" IS 'Protocol: openai or anthropic';
COMMENT ON COLUMN plugin_linapro_ai_model."source" IS 'Model source: manual or api';
COMMENT ON COLUMN plugin_linapro_ai_model."enabled" IS 'Enabled flag: 0=disabled 1=enabled';
COMMENT ON COLUMN plugin_linapro_ai_model."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_linapro_ai_model."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_linapro_ai_model."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS idx_plugin_linapro_ai_model_identity_alive
    ON plugin_linapro_ai_model ("provider_id", "protocol", "model_name")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_model_provider_alive
    ON plugin_linapro_ai_model ("provider_id", "enabled", "protocol")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_model_endpoint_alive
    ON plugin_linapro_ai_model ("endpoint_id", "enabled")
    WHERE "deleted_at" IS NULL;

-- Purpose: Stores explicit multimodal capability declarations for provider models.
-- 用途：存储渠道模型显式声明的多模态能力。
CREATE TABLE IF NOT EXISTS plugin_linapro_ai_model_capability (
    "id"                  BIGSERIAL PRIMARY KEY,
    "model_id"            BIGINT NOT NULL,
    "endpoint_id"         BIGINT NOT NULL DEFAULT 0,
    "capability_type"     VARCHAR(32) NOT NULL,
    "capability_method"   VARCHAR(32) NOT NULL,
    "input_modalities"    VARCHAR(128) NOT NULL DEFAULT '',
    "output_modalities"   VARCHAR(128) NOT NULL DEFAULT '',
    "max_input_tokens"    INTEGER NOT NULL DEFAULT 0,
    "max_output_tokens"   INTEGER NOT NULL DEFAULT 0,
    "max_input_assets"    INTEGER NOT NULL DEFAULT 0,
    "max_output_assets"   INTEGER NOT NULL DEFAULT 0,
    "max_asset_bytes"     BIGINT NOT NULL DEFAULT 0,
    "supports_thinking"   SMALLINT NOT NULL DEFAULT 0,
    "supported_efforts"   VARCHAR(128) NOT NULL DEFAULT '',
    "supports_streaming"  SMALLINT NOT NULL DEFAULT 0,
    "supports_operation"  SMALLINT NOT NULL DEFAULT 0,
    "default_params_json" TEXT NOT NULL DEFAULT '{}',
    "enabled"             SMALLINT NOT NULL DEFAULT 1,
    "created_at"          TIMESTAMP,
    "updated_at"          TIMESTAMP,
    "deleted_at"          TIMESTAMP
);

COMMENT ON TABLE plugin_linapro_ai_model_capability IS 'AI model method capability declaration table';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."id" IS 'Model capability ID';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."model_id" IS 'Model ID';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."endpoint_id" IS 'Preferred provider endpoint ID, 0 means model default';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."capability_type" IS 'Capability type';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."capability_method" IS 'Capability method';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."input_modalities" IS 'Comma-separated input modality list';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."output_modalities" IS 'Comma-separated output modality list';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."max_input_tokens" IS 'Maximum input tokens, 0 means unspecified';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."max_output_tokens" IS 'Maximum output tokens, 0 means unspecified';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."max_input_assets" IS 'Maximum input assets, 0 means unspecified';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."max_output_assets" IS 'Maximum output assets, 0 means unspecified';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."max_asset_bytes" IS 'Maximum single asset bytes, 0 means unspecified';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."supports_thinking" IS 'Thinking effort support flag for this model method: 0=no 1=yes';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."supported_efforts" IS 'Comma-separated thinking efforts supported by this model method';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."supports_streaming" IS 'Streaming support flag: 0=no 1=yes';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."supports_operation" IS 'Provider operation support flag: 0=no 1=yes';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."default_params_json" IS 'Method-specific default params JSON';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."enabled" IS 'Enabled flag: 0=disabled 1=enabled';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_linapro_ai_model_capability."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS idx_plugin_linapro_ai_model_capability_identity_alive
    ON plugin_linapro_ai_model_capability ("model_id", "capability_type", "capability_method")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_model_capability_method_alive
    ON plugin_linapro_ai_model_capability ("capability_type", "capability_method", "enabled")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_model_capability_endpoint_alive
    ON plugin_linapro_ai_model_capability ("endpoint_id", "enabled")
    WHERE "deleted_at" IS NULL;

-- Purpose: Stores fixed AI capability tiers and their latest diagnostic test summaries.
-- 用途：存储固定 AI 能力档位及其最近一次诊断测试摘要。
CREATE TABLE IF NOT EXISTS plugin_linapro_ai_tier (
    "id"                      BIGSERIAL PRIMARY KEY,
    "capability_type"         VARCHAR(32) NOT NULL DEFAULT 'text',
    "capability_method"       VARCHAR(32) NOT NULL DEFAULT 'generate',
    "code"                    VARCHAR(32) NOT NULL,
    "display_name"            VARCHAR(64) NOT NULL,
    "description"             VARCHAR(512) NOT NULL DEFAULT '',
    "default_effort"          VARCHAR(32) NOT NULL DEFAULT '',
    "enabled"                 SMALLINT NOT NULL DEFAULT 1,
    "sort_order"              INTEGER NOT NULL DEFAULT 0,
    "last_test_status"        VARCHAR(32) NOT NULL DEFAULT '',
    "last_test_latency_ms"    INTEGER NOT NULL DEFAULT 0,
    "last_test_error_summary" VARCHAR(512) NOT NULL DEFAULT '',
    "last_test_at"            TIMESTAMP,
    "created_at"              TIMESTAMP,
    "updated_at"              TIMESTAMP,
    CONSTRAINT uk_plugin_linapro_ai_tier_capability_code UNIQUE ("capability_type", "capability_method", "code")
);

COMMENT ON TABLE plugin_linapro_ai_tier IS 'AI capability tier table';
COMMENT ON COLUMN plugin_linapro_ai_tier."id" IS 'Tier ID';
COMMENT ON COLUMN plugin_linapro_ai_tier."capability_type" IS 'Capability type';
COMMENT ON COLUMN plugin_linapro_ai_tier."capability_method" IS 'Capability method';
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

CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_tier_sort
    ON plugin_linapro_ai_tier ("capability_type", "capability_method", "sort_order");

-- Purpose: Stores provider-model bindings for each AI capability tier.
-- 用途：存储每个 AI 能力档位对应的渠道模型绑定关系。
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

-- Purpose: Stores masked AI invocation audit summaries for Smart Center diagnostics.
-- 用途：存储用于智能中心诊断的脱敏 AI 调用审计摘要。
CREATE TABLE IF NOT EXISTS plugin_linapro_ai_invocation (
    "id"                     BIGSERIAL PRIMARY KEY,
    "request_id"             VARCHAR(64) NOT NULL DEFAULT '',
    "capability_type"        VARCHAR(32) NOT NULL DEFAULT 'text',
    "capability_method"      VARCHAR(32) NOT NULL DEFAULT 'generate',
    "purpose"                VARCHAR(128) NOT NULL DEFAULT '',
    "tier_code"              VARCHAR(32) NOT NULL DEFAULT '',
    "source_plugin_id"       VARCHAR(128) NOT NULL DEFAULT '',
    "tenant_id"              INTEGER NOT NULL DEFAULT 0,
    "user_id"                INTEGER NOT NULL DEFAULT 0,
    "provider_id"            BIGINT NOT NULL DEFAULT 0,
    "model_id"               BIGINT NOT NULL DEFAULT 0,
    "provider_name"          VARCHAR(128) NOT NULL DEFAULT '',
    "model_name"             VARCHAR(128) NOT NULL DEFAULT '',
    "protocol"               VARCHAR(32) NOT NULL DEFAULT '',
    "thinking_effort"        VARCHAR(32) NOT NULL DEFAULT '',
    "status"                 VARCHAR(32) NOT NULL DEFAULT '',
    "input_tokens"           INTEGER NOT NULL DEFAULT 0,
    "output_tokens"          INTEGER NOT NULL DEFAULT 0,
    "latency_ms"             INTEGER NOT NULL DEFAULT 0,
    "error_code"             VARCHAR(128) NOT NULL DEFAULT '',
    "error_summary"          VARCHAR(512) NOT NULL DEFAULT '',
    "asset_summary_json"     TEXT NOT NULL DEFAULT '{}',
    "operation_summary_json" TEXT NOT NULL DEFAULT '{}',
    "metadata_summary_json"  TEXT NOT NULL DEFAULT '{}',
    "created_at"             TIMESTAMP
);

COMMENT ON TABLE plugin_linapro_ai_invocation IS 'AI invocation audit log table';
COMMENT ON COLUMN plugin_linapro_ai_invocation."id" IS 'Invocation ID';
COMMENT ON COLUMN plugin_linapro_ai_invocation."request_id" IS 'Request correlation ID';
COMMENT ON COLUMN plugin_linapro_ai_invocation."capability_type" IS 'Capability type';
COMMENT ON COLUMN plugin_linapro_ai_invocation."capability_method" IS 'Capability method';
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
COMMENT ON COLUMN plugin_linapro_ai_invocation."asset_summary_json" IS 'Asset reference summary JSON without file contents';
COMMENT ON COLUMN plugin_linapro_ai_invocation."operation_summary_json" IS 'Provider operation summary JSON without provider secrets';
COMMENT ON COLUMN plugin_linapro_ai_invocation."metadata_summary_json" IS 'Bounded metadata summary JSON without request or response bodies';
COMMENT ON COLUMN plugin_linapro_ai_invocation."created_at" IS 'Creation time';

CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_invocation_created
    ON plugin_linapro_ai_invocation ("created_at" DESC);
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_invocation_filters
    ON plugin_linapro_ai_invocation ("capability_type", "capability_method", "tier_code", "status", "created_at" DESC);
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_invocation_provider_model
    ON plugin_linapro_ai_invocation ("provider_id", "model_id", "created_at" DESC);
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_invocation_purpose
    ON plugin_linapro_ai_invocation ("purpose", "created_at" DESC);
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_invocation_source_plugin
    ON plugin_linapro_ai_invocation ("source_plugin_id", "created_at" DESC);
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_invocation_method_created
    ON plugin_linapro_ai_invocation ("capability_type", "capability_method", "created_at" DESC);

-- Purpose: Stores method-specific default parameter templates for AI capability calls.
-- 用途：存储 AI 能力调用的方法级默认参数模板。
CREATE TABLE IF NOT EXISTS plugin_linapro_ai_method_default_param (
    "id"                  BIGSERIAL PRIMARY KEY,
    "capability_type"     VARCHAR(32) NOT NULL,
    "capability_method"   VARCHAR(32) NOT NULL,
    "default_params_json" TEXT NOT NULL DEFAULT '{}',
    "enabled"             SMALLINT NOT NULL DEFAULT 1,
    "created_at"          TIMESTAMP,
    "updated_at"          TIMESTAMP,
    CONSTRAINT uk_plugin_linapro_ai_method_default_identity UNIQUE ("capability_type", "capability_method")
);

COMMENT ON TABLE plugin_linapro_ai_method_default_param IS 'AI capability method default params table';
COMMENT ON COLUMN plugin_linapro_ai_method_default_param."id" IS 'Method default param ID';
COMMENT ON COLUMN plugin_linapro_ai_method_default_param."capability_type" IS 'Capability type';
COMMENT ON COLUMN plugin_linapro_ai_method_default_param."capability_method" IS 'Capability method';
COMMENT ON COLUMN plugin_linapro_ai_method_default_param."default_params_json" IS 'Method-specific default params JSON';
COMMENT ON COLUMN plugin_linapro_ai_method_default_param."enabled" IS 'Enabled flag: 0=disabled 1=enabled';
COMMENT ON COLUMN plugin_linapro_ai_method_default_param."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_linapro_ai_method_default_param."updated_at" IS 'Update time';

-- Purpose: Stores minimal provider operation projections without business task ownership.
-- 用途：存储不承载业务任务归属的最小渠道 Operation 投影。
CREATE TABLE IF NOT EXISTS plugin_linapro_ai_provider_operation (
    "id"                 BIGSERIAL PRIMARY KEY,
    "operation_ref"      VARCHAR(128) NOT NULL,
    "capability_type"    VARCHAR(32) NOT NULL,
    "capability_method"  VARCHAR(32) NOT NULL,
    "purpose"            VARCHAR(128) NOT NULL DEFAULT '',
    "source_plugin_id"   VARCHAR(128) NOT NULL DEFAULT '',
    "provider_id"        BIGINT NOT NULL DEFAULT 0,
    "model_id"           BIGINT NOT NULL DEFAULT 0,
    "provider_name"      VARCHAR(128) NOT NULL DEFAULT '',
    "model_name"         VARCHAR(128) NOT NULL DEFAULT '',
    "protocol"           VARCHAR(64) NOT NULL DEFAULT '',
    "status"             VARCHAR(32) NOT NULL,
    "next_poll_after_ms" BIGINT NOT NULL DEFAULT 0,
    "expires_at"         TIMESTAMP,
    "asset_summary_json" TEXT NOT NULL DEFAULT '{}',
    "error_code"         VARCHAR(128) NOT NULL DEFAULT '',
    "error_summary"      VARCHAR(512) NOT NULL DEFAULT '',
    "created_at"         TIMESTAMP,
    "updated_at"         TIMESTAMP,
    "deleted_at"         TIMESTAMP
);

COMMENT ON TABLE plugin_linapro_ai_provider_operation IS 'AI provider operation projection table';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."id" IS 'Provider operation row ID';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."operation_ref" IS 'Opaque provider operation reference';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."capability_type" IS 'Capability type';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."capability_method" IS 'Capability method';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."purpose" IS 'Governed AI purpose';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."source_plugin_id" IS 'Source plugin ID';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."provider_id" IS 'Provider ID';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."model_id" IS 'Model ID';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."provider_name" IS 'Provider display name snapshot';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."model_name" IS 'Model name snapshot';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."protocol" IS 'Protocol snapshot';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."status" IS 'Provider operation status';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."next_poll_after_ms" IS 'Recommended next poll delay in milliseconds';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."expires_at" IS 'Operation reference expiration time';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."asset_summary_json" IS 'Asset reference summary JSON without file contents';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."error_code" IS 'Stable error code';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."error_summary" IS 'Masked error summary';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_linapro_ai_provider_operation."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS idx_plugin_linapro_ai_operation_ref_alive
    ON plugin_linapro_ai_provider_operation ("operation_ref")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_operation_filters_alive
    ON plugin_linapro_ai_provider_operation ("capability_type", "capability_method", "status", "created_at" DESC)
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_linapro_ai_operation_provider_model_alive
    ON plugin_linapro_ai_provider_operation ("provider_id", "model_id", "created_at" DESC)
    WHERE "deleted_at" IS NULL;

WITH capability_methods("capability_type", "capability_method", "tier_description") AS (
    VALUES
        ('text', 'generate', ''),
        ('image', 'generate', 'Image generation capability tier.'),
        ('image', 'edit', 'Image editing capability tier.'),
        ('embedding', 'create', 'Embedding creation capability tier.'),
        ('audio', 'transcribe', 'Audio transcription capability tier.'),
        ('audio', 'synthesize', 'Audio synthesis capability tier.'),
        ('vision', 'analyze', 'Vision analysis capability tier.'),
        ('document', 'analyze', 'Document analysis capability tier.'),
        ('document', 'cite', 'Citation-aware document capability tier.'),
        ('safety', 'moderate', 'Safety moderation capability tier.'),
        ('video', 'generate', 'Video generation capability tier.'),
        ('video', 'edit', 'Video editing capability tier.'),
        ('video', 'extend', 'Video extension capability tier.'),
        ('video', 'operation.get', 'Provider operation lookup capability tier.'),
        ('video', 'operation.cancel', 'Provider operation cancellation capability tier.')
),
tier_catalog("code", "display_name", "sort_order", "text_description") AS (
    VALUES
        ('basic', 'Basic', 1, 'Low-cost AI capability tier for simple text generation and commit message generation.'),
        ('standard', 'Standard', 2, 'Default AI capability tier for regular code generation and code optimization.'),
        ('advanced', 'Advanced', 3, 'High-capability AI tier for complex code generation and cross-file reasoning.')
),
method_default_insert AS (
    INSERT INTO plugin_linapro_ai_method_default_param (
        "capability_type",
        "capability_method",
        "enabled"
    )
    SELECT
        methods."capability_type",
        methods."capability_method",
        1
    FROM capability_methods AS methods
    ON CONFLICT ("capability_type", "capability_method") DO NOTHING
    RETURNING 1
),
method_default_insert_marker AS (
    SELECT COUNT(*) AS "ignored" FROM method_default_insert
)
INSERT INTO plugin_linapro_ai_tier (
    "capability_type",
    "capability_method",
    "code",
    "display_name",
    "description",
    "default_effort",
    "enabled",
    "sort_order"
)
SELECT
    methods."capability_type",
    methods."capability_method",
    tiers."code",
    tiers."display_name",
    CASE
        WHEN methods."capability_type" = 'text' AND methods."capability_method" = 'generate' THEN tiers."text_description"
        ELSE methods."tier_description"
    END,
    '',
    1,
    tiers."sort_order"
FROM capability_methods AS methods
CROSS JOIN tier_catalog AS tiers
CROSS JOIN method_default_insert_marker AS marker
ON CONFLICT ("capability_type", "capability_method", "code") DO NOTHING;

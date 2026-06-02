-- LinaPro AI core plugin uninstall schema.
-- Drops plugin-owned AI tables in dependency order.

DROP TABLE IF EXISTS plugin_linapro_ai_invocation;
DROP TABLE IF EXISTS plugin_linapro_ai_tier_binding;
DROP TABLE IF EXISTS plugin_linapro_ai_tier;
DROP TABLE IF EXISTS plugin_linapro_ai_model;
DROP TABLE IF EXISTS plugin_linapro_ai_provider;

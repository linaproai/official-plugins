-- Mock data: tenant-scoped scheduled-job groups and demo jobs.
-- 模拟数据：租户级定时任务分组与演示任务。

-- Each demo tenant gets its own default group so tenant job-group pages never
-- depend on the platform tenant_id=0 default group.
INSERT INTO sys_job_group ("tenant_id", "code", "name", "remark", "sort_order", "is_default", "created_at", "updated_at")
SELECT t."id", 'default', 'Default Group', 'Tenant default job group for multi-tenant demos.', 0, 1, NOW(), NOW()
FROM plugin_multi_tenant_tenant t
WHERE t."status" IN ('active', 'suspended')
ON CONFLICT ("tenant_id", "code") DO NOTHING;

-- A separate non-default group gives tenant job-group isolation demos visible
-- data without reusing platform mock groups.
INSERT INTO sys_job_group ("tenant_id", "code", "name", "remark", "sort_order", "is_default", "created_at", "updated_at")
SELECT t."id", 'tenant-maintenance', 'Tenant Maintenance', 'Tenant-scoped maintenance jobs for isolation demos.', 10, 0, NOW(), NOW()
FROM plugin_multi_tenant_tenant t
WHERE t."code" IN ('alpha-retail', 'beta-manufacturing')
ON CONFLICT ("tenant_id", "code") DO NOTHING;

INSERT INTO sys_job (
    "tenant_id",
    "group_id",
    "name",
    "description",
    "task_type",
    "handler_ref",
    "params",
    "timeout_seconds",
    "shell_cmd",
    "cron_expr",
    "timezone",
    "scope",
    "concurrency",
    "max_concurrency",
    "max_executions",
    "executed_count",
    "status",
    "is_builtin",
    "seed_version",
    "created_by",
    "updated_by",
    "created_at",
    "updated_at"
)
SELECT
    t."id",
    g."id",
    'Tenant demo cleanup',
    'Disabled tenant-scoped shell job used to verify job-group isolation.',
    'shell',
    '',
    '{}',
    60,
    'printf tenant-cleanup',
    '0 */20 * * * *',
    'Asia/Shanghai',
    'master_only',
    'singleton',
    1,
    0,
    0,
    'disabled',
    0,
    1,
    COALESCE(u."id", 0),
    COALESCE(u."id", 0),
    '2026-05-05 09:00:00',
    '2026-05-05 09:00:00'
FROM plugin_multi_tenant_tenant t
JOIN sys_job_group g ON g."tenant_id" = t."id" AND g."code" = 'tenant-maintenance'
LEFT JOIN sys_user u ON u."tenant_id" = t."id" AND u."username" IN ('tenant_alpha_admin', 'tenant_beta_admin')
WHERE t."code" IN ('alpha-retail', 'beta-manufacturing')
ON CONFLICT ("tenant_id", "group_id", "name") DO NOTHING;

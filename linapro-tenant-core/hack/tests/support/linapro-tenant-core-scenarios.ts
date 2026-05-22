import {
  createAdminApiContext,
  createUser,
  deleteUser,
  expectBusinessError,
  expectSuccess,
  getAccessibleMenus,
  getPlugin,
  type MenuNode,
  enablePlugin,
  installPlugin,
  syncPlugins,
  type APIRequestContext,
  type APIResponse,
} from "@host-tests/support/api/job";
import type { Page } from "@host-tests/support/playwright";
import { LoginPage } from "@host-tests/pages/LoginPage";
import { workspacePath } from "@host-tests/fixtures/config";
import {
  execPgSQL,
  pgEscapeLiteral,
  queryPgRows,
  queryPgScalar,
} from "@host-tests/support/postgres";
import { waitForRouteReady } from "@host-tests/support/ui";
import {
  addTenantMember,
  createTenant,
  createTenantApiContext,
  deleteTenant,
  ensureMultiTenantPluginEnabled,
  grantTenantPermissions,
  loginRaw,
  removeTenantMember,
  revokeTenantPermissionGrants,
  selectTenant,
  switchTenant,
  tenantCoreApiPath,
  type TenantUserGrant,
  updateUserPrimaryTenant,
  expect,
} from "./linapro-tenant-core";

const password = "test123456";
const tenantEnablementKey = "__tenant_enabled__";

type TenantRecord = {
  id: number;
  code: string;
  name: string;
};

type MonitorAutoEnableSnapshot = {
  autoEnableForNewTenants: string;
  currentState: string;
  desiredState: string;
  installMode: string;
  installed: string;
  pluginId: string;
  status: string;
};

type ErrorEnvelope = {
  code: number;
  errorCode?: string;
  message?: string;
  messageKey?: string;
};

type RoleMenuTreeNode = {
  children?: RoleMenuTreeNode[];
  id: number;
  label: string;
  parentId: number;
  type: string;
};

type RoleMenuTreeResponse = {
  checkedKeys: number[];
  menus: RoleMenuTreeNode[];
};

type ConfigFallbackItem = {
  canEdit: boolean;
  canOverride: boolean;
  id: number;
  isFallback: boolean;
  key: string;
  overrideMode: string;
  sourceTenantId: number;
};

type DictTypeFallbackItem = {
  canEdit: boolean;
  canOverride: boolean;
  id: number;
  isFallback: boolean;
  overrideMode: string;
  sourceTenantId: number;
  type: string;
};

type DictDataFallbackItem = {
  canEdit: boolean;
  canOverride: boolean;
  dictType: string;
  id: number;
  isFallback: boolean;
  label: string;
  overrideMode: string;
  sourceTenantId: number;
  value: string;
};

type JobGroupFallbackItem = {
  code: string;
  id: number;
  jobCount: number;
  name: string;
};

type ScenarioContext = {
  api: APIRequestContext;
  suffix: string;
};

async function withAdmin<T>(fn: (ctx: ScenarioContext) => Promise<T>) {
  const api = await createAdminApiContext();
  try {
    await ensureMultiTenantPluginEnabled(api);
    return await fn({ api, suffix: Date.now().toString() });
  } finally {
    await api.dispose();
  }
}

async function withTenant<T>(
  api: APIRequestContext,
  suffix: string,
  prefix: string,
  fn: (tenant: TenantRecord) => Promise<T>,
) {
  const tenant = await createNamedTenant(api, suffix, prefix);
  try {
    return await fn(tenant);
  } finally {
    await deleteTenant(api, tenant.id).catch(() => {});
  }
}

async function createNamedTenant(
  api: APIRequestContext,
  suffix: string,
  prefix: string,
): Promise<TenantRecord> {
  const code = `${prefix}-${suffix}`.toLowerCase().slice(0, 32);
  const result = await createTenant(api, {
    code,
    name: `${prefix.toUpperCase()} Tenant ${suffix}`,
  });
  return {
    id: result.id,
    code,
    name: `${prefix.toUpperCase()} Tenant ${suffix}`,
  };
}

async function createTenantUser(
  api: APIRequestContext,
  suffix: string,
  prefix: string,
  tenantId: number,
) {
  const username = `${prefix}_${suffix}`.replaceAll("-", "_").slice(0, 60);
  const user = await createUser(api, {
    username,
    password,
    nickname: `${prefix} user`,
  });
  updateUserPrimaryTenant(username, tenantId);
  return { id: user.id, username };
}

async function addTenantUser(
  api: APIRequestContext,
  suffix: string,
  prefix: string,
  tenantId: number,
) {
  const user = await createTenantUser(api, suffix, prefix, tenantId);
  const member = await addTenantMember(api, {
    tenantId,
    userId: user.id,
  });
  return { ...user, memberId: member.id };
}

async function cleanupTenantUser(
  api: APIRequestContext,
  userId: number,
  memberId = 0,
) {
  execPgSQL(
    `DELETE FROM plugin_linapro_tenant_core_user_membership WHERE user_id = ${userId};`,
  );
  if (memberId > 0) {
    await removeTenantMember(api, memberId).catch(() => {});
  }
  execPgSQL(`DELETE FROM sys_user_role WHERE user_id = ${userId};`);
  await deleteUser(api, userId).catch(() => {});
}

function scalarNumber(sql: string) {
  return Number(queryPgScalar(sql) || "0");
}

function pluginRow(pluginId: string) {
  const rows = queryPgRows(`
    SELECT scope_nature || '|' || install_mode || '|' || auto_enable_for_new_tenants::text || '|' || installed::text || '|' || status::text
    FROM sys_plugin
    WHERE plugin_id = '${pgEscapeLiteral(pluginId)}'
    LIMIT 1;
  `);
  expect(rows.length, `missing sys_plugin row for ${pluginId}`).toBe(1);
  const [
    scopeNature,
    installMode,
    autoEnableForNewTenants,
    installed,
    enabled,
  ] = rows[0].split("|");
  return {
    scopeNature,
    installMode,
    autoEnableForNewTenants: autoEnableForNewTenants === "true",
    installed: Number(installed),
    enabled: Number(enabled),
  };
}

function tableExists(table: string) {
  return (
    scalarNumber(`
    SELECT COUNT(1)
    FROM information_schema.tables
    WHERE table_schema = 'public'
      AND table_name = '${pgEscapeLiteral(table)}';
  `) === 1
  );
}

function cleanupRowsByPrefix(prefix: string) {
  const escaped = pgEscapeLiteral(prefix);
  execPgSQL(`
    DELETE FROM sys_role_menu WHERE role_id IN (SELECT id FROM sys_role WHERE key LIKE '${escaped}%');
    DELETE FROM sys_user_role WHERE role_id IN (SELECT id FROM sys_role WHERE key LIKE '${escaped}%');
    DELETE FROM sys_role WHERE key LIKE '${escaped}%';
    DELETE FROM sys_dict_data WHERE dict_type LIKE '${escaped}%';
    DELETE FROM sys_dict_type WHERE type LIKE '${escaped}%';
    DELETE FROM sys_config WHERE key LIKE '${escaped}%';
    DELETE FROM sys_file WHERE name LIKE '${escaped}%';
    DELETE FROM sys_notify_delivery WHERE message_id IN (SELECT id FROM sys_notify_message WHERE source_id LIKE '${escaped}%');
    DELETE FROM sys_notify_message WHERE source_id LIKE '${escaped}%';
    DELETE FROM sys_job_log WHERE job_id IN (SELECT id FROM sys_job WHERE name LIKE '${escaped}%');
    DELETE FROM sys_job WHERE name LIKE '${escaped}%';
    DELETE FROM sys_job_group WHERE code LIKE '${escaped}%';
    DELETE FROM sys_online_session WHERE token_id LIKE '${escaped}%';
    DO $$
    BEGIN
      IF to_regclass('public.plugin_linapro_monitor_loginlog') IS NOT NULL THEN
        DELETE FROM plugin_linapro_monitor_loginlog WHERE user_name LIKE '${escaped}%';
      END IF;
      IF to_regclass('public.plugin_linapro_monitor_operlog') IS NOT NULL THEN
        DELETE FROM plugin_linapro_monitor_operlog WHERE oper_name LIKE '${escaped}%';
      END IF;
      IF to_regclass('public.plugin_linapro_org_core_post') IS NOT NULL THEN
        DELETE FROM plugin_linapro_org_core_post WHERE code LIKE '${escaped}%';
      END IF;
      IF to_regclass('public.plugin_linapro_org_core_dept') IS NOT NULL THEN
        DELETE FROM plugin_linapro_org_core_dept WHERE code LIKE '${escaped}%';
      END IF;
    END $$;
  `);
}

function flattenAccessibleMenus(list: Array<any>): Array<any> {
  return list.flatMap((item) => [
    item,
    ...flattenAccessibleMenus(item?.children ?? []),
  ]);
}

function flattenRoleMenuTreeNodes(list: RoleMenuTreeNode[]): RoleMenuTreeNode[] {
  return list.flatMap((item) => [
    item,
    ...flattenRoleMenuTreeNodes(item.children ?? []),
  ]);
}

function menuIDByPermission(permission: string) {
  const menuID = scalarNumber(`
    SELECT id
    FROM sys_menu
    WHERE perms = '${pgEscapeLiteral(permission)}'
    ORDER BY id
    LIMIT 1;
  `);
  expect(menuID, `missing menu permission ${permission}`).toBeGreaterThan(0);
  return menuID;
}

function menuIDsByPermissions(permissions: string[]) {
  return permissions.map((permission) => menuIDByPermission(permission));
}

async function expectBusinessErrorCode(
  response: APIResponse,
  expectedErrorCode: string,
) {
  const payload = (await expectBusinessError(response)) as ErrorEnvelope;
  expect(payload.errorCode).toBe(expectedErrorCode);
  return payload;
}

async function loginTenantUserInBrowser(page: Page, username: string) {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.login(username, password);
  await page
    .waitForURL((url) => !url.pathname.includes("/auth/login"), {
      timeout: 15000,
    })
    .catch(() => {});
  const tenantSelector = page.getByTestId("login-tenant-selector");
  if (await tenantSelector.isVisible({ timeout: 1000 }).catch(() => false)) {
    await page.getByTestId("login-tenant-confirm").click();
    await page.waitForURL((url) => !url.pathname.includes("/auth/login"), {
      timeout: 15000,
    });
  }
  await waitForRouteReady(page, 15000);
}

function assertNoRoutePath(list: Array<any>, path: string) {
  const pathsByMenu = new Set(
    flattenAccessibleMenus(list).map((item) => item?.path),
  );
  expect(pathsByMenu.has(path), `route path ${path} should be hidden`).toBe(
    false,
  );
}

function expectRoleTreeExcludesMenuIDs(
  tree: RoleMenuTreeResponse,
  menuIDs: number[],
) {
  const treeIDs = new Set(flattenRoleMenuTreeNodes(tree.menus).map((node) => node.id));
  for (const menuID of menuIDs) {
    expect(treeIDs.has(menuID), `role tree should not expose menu ${menuID}`).toBe(
      false,
    );
    expect(
      tree.checkedKeys.includes(menuID),
      `checked keys should not expose menu ${menuID}`,
    ).toBe(false);
  }
}

function expectFallbackMetadata(
  row: {
    canEdit: boolean;
    canOverride: boolean;
    isFallback: boolean;
    overrideMode: string;
    sourceTenantId: number;
  },
  label: string,
) {
  expect(row.sourceTenantId, `${label} source tenant`).toBe(0);
  expect(row.isFallback, `${label} fallback flag`).toBe(true);
  expect(row.canEdit, `${label} direct edit flag`).toBe(false);
  expect(row.canOverride, `${label} override flag`).toBe(true);
  expect(row.overrideMode, `${label} override mode`).toBe(
    "createTenantOverride",
  );
}

function insertTenantDefaultJobGroup(tenantId: number) {
  return scalarNumber(`
    INSERT INTO sys_job_group (tenant_id, code, name, remark, sort_order, is_default, created_at, updated_at)
    VALUES (${tenantId}, 'default', 'Default group', 'Tenant default scheduled-job group', 0, 1, NOW(), NOW())
    ON CONFLICT (tenant_id, code) DO UPDATE SET is_default = 1, updated_at = NOW()
    RETURNING id;
  `);
}

function insertJobGroup(
  tenantId: number,
  code: string,
  name: string,
  isDefault = false,
) {
  return scalarNumber(`
    INSERT INTO sys_job_group (tenant_id, code, name, remark, sort_order, is_default, created_at, updated_at)
    VALUES (${tenantId}, '${pgEscapeLiteral(code)}', '${pgEscapeLiteral(name)}', 'linapro-tenant-core e2e', 10, ${isDefault ? 1 : 0}, NOW(), NOW())
    RETURNING id;
  `);
}

function insertShellJob(tenantId: number, groupId: number, name: string) {
  return scalarNumber(`
    INSERT INTO sys_job (
      tenant_id,
      group_id,
      name,
      description,
      task_type,
      handler_ref,
      params,
      timeout_seconds,
      shell_cmd,
      work_dir,
      env,
      cron_expr,
      timezone,
      scope,
      concurrency,
      max_concurrency,
      max_executions,
      status,
      is_builtin,
      created_at,
      updated_at
    )
    VALUES (
      ${tenantId},
      ${groupId},
      '${pgEscapeLiteral(name)}',
      'Tenant job-group isolation fixture',
      'shell',
      '',
      '{}',
      300,
      'echo tc',
      '',
      '{}',
      '0 0 1 1 *',
      'Asia/Shanghai',
      'master_only',
      'singleton',
      1,
      0,
      'disabled',
      0,
      NOW(),
      NOW()
    )
    RETURNING id;
  `);
}

function monitorPluginSnapshot(pluginId: string): MonitorAutoEnableSnapshot {
  const rows = queryPgRows(`
    SELECT plugin_id || '|' || installed::text || '|' || status::text || '|' || install_mode || '|' || auto_enable_for_new_tenants::text || '|' || desired_state || '|' || current_state
    FROM sys_plugin
    WHERE plugin_id = '${pgEscapeLiteral(pluginId)}'
    LIMIT 1;
  `);
  expect(rows.length, `missing sys_plugin row for ${pluginId}`).toBe(1);
  const [
    snapshotPluginId,
    installed,
    status,
    installMode,
    autoEnableForNewTenants,
    desiredState,
    currentState,
  ] = rows[0].split("|");
  return {
    pluginId: snapshotPluginId,
    installed,
    status,
    installMode,
    autoEnableForNewTenants,
    desiredState,
    currentState,
  };
}

function restoreMonitorPluginSnapshot(snapshot: MonitorAutoEnableSnapshot) {
  execPgSQL(`
    UPDATE sys_plugin
    SET installed = ${Number(snapshot.installed) || 0},
        status = ${Number(snapshot.status) || 0},
        install_mode = '${pgEscapeLiteral(snapshot.installMode)}',
        auto_enable_for_new_tenants = ${snapshot.autoEnableForNewTenants === "true" ? "TRUE" : "FALSE"},
        desired_state = '${pgEscapeLiteral(snapshot.desiredState)}',
        current_state = '${pgEscapeLiteral(snapshot.currentState)}',
        updated_at = NOW()
    WHERE plugin_id = '${pgEscapeLiteral(snapshot.pluginId)}';
  `);
}

function enableMonitorTenantStateForTenant(pluginId: string, tenantId: number) {
  execPgSQL(`
    INSERT INTO sys_plugin_state (plugin_id, tenant_id, state_key, state_value, enabled, created_at, updated_at)
    VALUES ('${pgEscapeLiteral(pluginId)}', ${tenantId}, '${tenantEnablementKey}', 'enabled', TRUE, NOW(), NOW())
    ON CONFLICT (plugin_id, tenant_id, state_key) DO NOTHING;
  `);
}

function setMonitorPluginsAutoEnabledForTenant(
  pluginIds: string[],
  tenantId: number,
) {
  for (const pluginId of pluginIds) {
    execPgSQL(`
      UPDATE sys_plugin
      SET installed = 1,
          status = 1,
          install_mode = 'tenant_scoped',
          auto_enable_for_new_tenants = TRUE,
          updated_at = NOW()
      WHERE plugin_id = '${pgEscapeLiteral(pluginId)}';
    `);
    enableMonitorTenantStateForTenant(pluginId, tenantId);
  }
}

function removeTenantMonitorStates(tenantId: number, pluginIds: string[]) {
  execPgSQL(`
    DELETE FROM sys_plugin_state
    WHERE tenant_id = ${tenantId}
      AND state_key = '${tenantEnablementKey}'
      AND plugin_id IN (${pluginIds.map((pluginId) => `'${pgEscapeLiteral(pluginId)}'`).join(", ")});
  `);
}

function assertAccessibleMenuPaths(list: Array<any>, paths: string[]) {
  const pathsByMenu = new Set(
    flattenAccessibleMenus(list).map((item) => item?.path),
  );
  for (const path of paths) {
    expect(
      pathsByMenu.has(path),
      `missing tenant monitor menu path ${path}`,
    ).toBe(true);
  }
}

function flattenMenuNodes(list: MenuNode[]): MenuNode[] {
  return list.flatMap((item) => [
    item,
    ...flattenMenuNodes(item.children ?? []),
  ]);
}

function assertTenantManagementButtonPermissions(list: MenuNode[]) {
  const flatMenus = flattenMenuNodes(list);
  const tenantMenu = flatMenus.find(
    (item) => item.perms === "system:tenant:list",
  );
  expect(tenantMenu, "missing tenant management menu").toBeTruthy();

  const buttonPerms = new Set(
    (tenantMenu?.children ?? [])
      .filter((item) => item.type === "B")
      .map((item) => item.perms),
  );
  expect([...buttonPerms].sort()).toEqual(
    [
      "system:tenant:add",
      "system:tenant:edit",
      "system:tenant:impersonate",
      "system:tenant:query",
      "system:tenant:remove",
    ].sort(),
  );
  for (const stalePrefix of [
    "system:tenant:resolver:",
    "system:tenant:member:",
    "system:tenant:plugin:",
  ]) {
    expect([...buttonPerms].some((perm) => perm.startsWith(stalePrefix))).toBe(
      false,
    );
  }
}

async function loginAndSelect(username: string, tenantId: number) {
  const login = await loginRaw(username, password);
  if (login.accessToken) {
    return login.accessToken;
  }
  expect(login.preToken).toBeTruthy();
  return selectTenant(login.preToken!, tenantId);
}

async function expectUserListContains(
  api: APIRequestContext,
  tenantId: number,
  username: string,
) {
  const users = await expectSuccess<{
    list: Array<{ username: string; tenantIds?: number[] }>;
  }>(
    await api.get(
      `user?pageNum=1&pageSize=100&tenantId=${tenantId}&username=${encodeURIComponent(username)}`,
    ),
  );
  expect(users.list.map((item) => item.username)).toContain(username);
}

function expectMembershipRow(userId: number, tenantId: number) {
  expect(
    scalarNumber(`
      SELECT COUNT(1)
      FROM plugin_linapro_tenant_core_user_membership
      WHERE user_id = ${userId}
        AND tenant_id = ${tenantId}
        AND status = 1
        AND deleted_at IS NULL;
    `),
  ).toBe(1);
}

function expectBuiltInResolverPolicyRemainsCodeOwned() {
  expect(tableExists("plugin_linapro_tenant_core_resolver_config")).toBeFalsy();
  expect(
    scalarNumber(`
      SELECT COUNT(1)
      FROM sys_cache_revision
      WHERE domain = 'tenant-resolution'
        AND scope = 'resolver-config';
    `),
  ).toBe(0);
}

async function expectResolverConfigEndpointRemoved(api: APIRequestContext) {
  expect(
    (
      await api.get("platform/tenant/resolver-config", { maxRedirects: 0 })
    ).ok(),
  ).toBe(false);
  expect(
    (
      await api.put("platform/tenant/resolver-config", {
        data: {},
        maxRedirects: 0,
      })
    ).ok(),
  ).toBe(false);
}

function insertTenantRole(tenantId: number, suffix: string, keyPrefix: string) {
  const dataScope = tenantId === 0 ? 1 : 2;
  return scalarNumber(`
    INSERT INTO sys_role (name, key, sort, data_scope, status, remark, tenant_id, created_at, updated_at)
    VALUES (
      '${pgEscapeLiteral(keyPrefix)} role',
      '${pgEscapeLiteral(keyPrefix)}_${pgEscapeLiteral(suffix)}',
      1,
      ${dataScope},
      1,
      'linapro-tenant-core e2e',
      ${tenantId},
      NOW(),
      NOW()
    )
    RETURNING id;
  `);
}

function assertTenantScopedRowCounts(
  table: string,
  tenantAId: number,
  tenantBId: number,
  label: string,
) {
  expect(
    scalarNumber(
      `SELECT COUNT(1) FROM ${table} WHERE tenant_id = ${tenantAId};`,
    ),
    `${label} tenant A row`,
  ).toBeGreaterThan(0);
  expect(
    scalarNumber(
      `SELECT COUNT(1) FROM ${table} WHERE tenant_id = ${tenantBId};`,
    ),
    `${label} tenant B row`,
  ).toBeGreaterThan(0);
}

async function ensurePluginInstalled(api: APIRequestContext, pluginId: string) {
  await syncPlugins(api);
  const plugin = await getPlugin(api, pluginId);
  if (plugin.installed !== 1) {
    await installPlugin(api, pluginId);
  }
}

async function ensurePluginInstalledAndEnabled(
  api: APIRequestContext,
  pluginId: string,
) {
  await syncPlugins(api);
  const plugin = await getPlugin(api, pluginId);
  if (plugin.installed !== 1) {
    await installPlugin(api, pluginId);
  }
  if (plugin.enabled !== 1) {
    await enablePlugin(api, pluginId);
  }
}

async function expectMultiTenantUninstallBlockedWithExistingTenant(
  api: APIRequestContext,
  suffix: string,
  prefix: string,
) {
  await withTenant(api, suffix, prefix, async () => {
    await expectBusinessError(
      await api.delete("plugins/linapro-tenant-core", {
        data: { purgeStorageData: 0, force: false },
      }),
    );
    expect(pluginRow("linapro-tenant-core").installed).toBe(1);
  });
}

async function expectMultiTenantForceUninstallBypassesGuard(
  api: APIRequestContext,
  suffix: string,
  prefix: string,
) {
  await withTenant(api, suffix, prefix, async (tenant) => {
    await expectSuccess(
      await api.delete("plugins/linapro-tenant-core?force=true", {
        data: { purgeStorageData: 0, force: true },
      }),
    );
    expect(pluginRow("linapro-tenant-core").installed).toBe(0);
    await ensureMultiTenantPluginEnabled(api);
    await deleteTenant(api, tenant.id).catch(() => {});
    expect(pluginRow("linapro-tenant-core").installed).toBe(1);
    expect(pluginRow("linapro-tenant-core").enabled).toBe(1);
  });
}

async function createTenantPluginApi(
  adminApi: APIRequestContext,
  tenantId: number,
  suffix: string,
) {
  const user = await addTenantUser(
    adminApi,
    suffix,
    `tenant_plugin_${tenantId}`,
    tenantId,
  );
  const grants = [
    await grantTenantPermissions(adminApi, {
      roleKey: `tenant_plugin_${tenantId}_${suffix}`,
      roleName: `Tenant plugin ${tenantId} ${suffix}`,
      tenantId,
      userId: user.id,
      permissions: [
        "system:tenant:plugin:list",
        "system:tenant:plugin:enable",
        "system:tenant:plugin:disable",
      ],
    }),
  ];
  const token = await loginAndSelect(user.username, tenantId);
  const api = await createTenantApiContext(token);
  return { api, grants, user };
}

export async function scenarioTC0178() {
  await withAdmin(async ({ api }) => {
    const plugin = pluginRow("linapro-tenant-core");
    expect(plugin.installed).toBe(1);
    expect(plugin.enabled).toBe(1);
    expect(plugin.scopeNature).toBe("platform_only");
    expect(tableExists("plugin_linapro_tenant_core_tenant")).toBeTruthy();
    expect(tableExists("plugin_linapro_tenant_core_user_membership")).toBeTruthy();
    const menuData = await expectSuccess<{ list: MenuNode[] }>(
      await api.get("menu"),
    );
    assertTenantManagementButtonPermissions(menuData.list);
    await expectSuccess(
      await api.get(tenantCoreApiPath("platform/tenants?pageNum=1&pageSize=1")),
    );
    const accessibleMenus = await getAccessibleMenus(api);
    const routesByPath = new Map(
      flattenAccessibleMenus(accessibleMenus.list).map((item) => [
        item.path,
        item,
      ]),
    );
    expect(routesByPath.get("/platform/tenants")?.component).toBe(
      "#/views/system/plugin/dynamic-page",
    );
    expect(routesByPath.get("/platform/tenant-members")).toBeUndefined();
    expect(routesByPath.get("/tenant")).toBeUndefined();
    expect(routesByPath.get("/tenant/members")).toBeUndefined();
    expect(routesByPath.get("/tenant/plugins")).toBeUndefined();
    expect(routesByPath.get("/platform/users")).toBeUndefined();
    expect(routesByPath.get("/platform/resolver-config")).toBeUndefined();
    expect(routesByPath.get("/system/user")?.component).toBe(
      "#/views/system/user/index",
    );
  });
}

export async function scenarioTC0179() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc179", async (tenant) => {
      const detail = await expectSuccess<{ id: number; code: string }>(
        await api.get(tenantCoreApiPath(`platform/tenants/${tenant.id}`)),
      );
      expect(detail.code).toBe(tenant.code);
      await expectBusinessError(
        await api.post(tenantCoreApiPath("platform/tenants"), {
          data: { code: tenant.code, name: "Duplicate" },
        }),
      );
      await expectSuccess(
        await api.put(tenantCoreApiPath(`platform/tenants/${tenant.id}`), {
          data: { name: `${tenant.name} updated` },
        }),
      );
      await expectSuccess(
        await api.delete(tenantCoreApiPath(`platform/tenants/${tenant.id}`)),
      );
      await expectBusinessError(
        await api.post(tenantCoreApiPath("platform/tenants"), {
          data: { code: tenant.code, name: "Tombstone" },
        }),
      );
    });
  });
}

export async function scenarioTC0180() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc180", async (tenant) => {
      const user = await addTenantUser(api, suffix, "tc180_user", tenant.id);
      try {
        await expectSuccess(
          await api.put(tenantCoreApiPath(`platform/tenants/${tenant.id}/status`), {
            data: { status: "suspended" },
          }),
        );
        await expectBusinessError(
          await api.post("auth/login", {
            data: { username: user.username, password },
          }),
        );
        await expectSuccess(
          await api.put(tenantCoreApiPath(`platform/tenants/${tenant.id}/status`), {
            data: { status: "active" },
          }),
        );
        const activeLogin = await loginRaw(user.username, password);
        expect(activeLogin.accessToken || activeLogin.preToken).toBeTruthy();
      } finally {
        await cleanupTenantUser(api, user.id, user.memberId);
      }
    });
  });
}

export async function scenarioTC0181() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc181", async (tenant) => {
      await expectBusinessError(
        await api.put(tenantCoreApiPath(`platform/tenants/${tenant.id}/status`), {
          data: { status: "archived" },
        }),
      );
      const status = queryPgScalar(
        `SELECT status FROM plugin_linapro_tenant_core_tenant WHERE id = ${tenant.id};`,
      );
      expect(status).toBe("active");
    });
  });
}

export async function scenarioTC0182() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc182", async (tenant) => {
      await expectSuccess(
        await api.delete(tenantCoreApiPath(`platform/tenants/${tenant.id}`)),
      );
      const deleted = scalarNumber(
        `SELECT COUNT(1) FROM plugin_linapro_tenant_core_tenant WHERE id = ${tenant.id} AND deleted_at IS NOT NULL;`,
      );
      expect(deleted).toBe(1);
    });
  });
}

export async function scenarioTC0183() {
  await withAdmin(async ({ api, suffix }) => {
    const plugin = pluginRow("linapro-tenant-core");
    expect(plugin.scopeNature).toBe("platform_only");
    expect(plugin.installMode).toBe("global");
    await expectMultiTenantUninstallBlockedWithExistingTenant(
      api,
      suffix,
      "tc183",
    );
  });
}

export async function scenarioTC0184() {
  await withAdmin(async ({ api, suffix }) => {
    const tenantA = await createNamedTenant(api, suffix, "tc184-a");
    const tenantB = await createNamedTenant(api, suffix, "tc184-b");
    const user = await createTenantUser(api, suffix, "tc184_user", tenantA.id);
    let memberA = 0;
    let memberB = 0;
    const grants: TenantUserGrant[] = [];
    try {
      memberA = (
        await addTenantMember(api, { tenantId: tenantA.id, userId: user.id })
      ).id;
      memberB = (
        await addTenantMember(api, { tenantId: tenantB.id, userId: user.id })
      ).id;
      grants.push(
        await grantTenantPermissions(api, {
          roleKey: `tc184-user-query-a-${suffix}`,
          roleName: `TC184 User Query A ${suffix}`,
          tenantId: tenantA.id,
          userId: user.id,
          permissions: ["system:user:query"],
        }),
      );
      const login = await loginRaw(user.username, password);
      expect(login.accessToken ?? "").toBe("");
      expect(login.preToken).toBeTruthy();
      expect(login.tenants?.map((tenant) => tenant.id)).toEqual([
        tenantA.id,
        tenantB.id,
      ]);
      const token = await selectTenant(login.preToken!, tenantA.id);
      const tenantApi = await createTenantApiContext(token);
      try {
        const userInfo = await expectSuccess<{ permissions: string[] }>(
          await tenantApi.get("user/info"),
        );
        expect(userInfo.permissions).toContain("system:user:query");
        await expectUserListContains(tenantApi, tenantA.id, user.username);
        expectMembershipRow(user.id, tenantA.id);
      } finally {
        await tenantApi.dispose();
      }
    } finally {
      revokeTenantPermissionGrants(grants);
      await cleanupTenantUser(api, user.id, memberA);
      if (memberB > 0) {
        await removeTenantMember(api, memberB).catch(() => {});
      }
      await deleteTenant(api, tenantA.id).catch(() => {});
      await deleteTenant(api, tenantB.id).catch(() => {});
    }
  });
}

export async function scenarioTC0185() {
  await scenarioSwitchToken("tc185");
}

export async function scenarioTC0186() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc186", async (tenant) => {
      const out = await expectSuccess<{
        token: string;
        tenantId: number;
        actingUserId: number;
        isImpersonated: boolean;
      }>(
        await api.post(tenantCoreApiPath(`platform/tenants/${tenant.id}/impersonate`), {
          data: { reason: "TC001" },
        }),
      );
      expect(out.token).toBeTruthy();
      expect(out.tenantId).toBe(tenant.id);
      expect(out.isImpersonated).toBeTruthy();
      const impersonatedApi = await createTenantApiContext(out.token);
      try {
        const userInfo = await expectSuccess<{
          menus: Array<unknown>;
          permissions: string[];
        }>(await impersonatedApi.get("user/info"));
        expect(userInfo.menus.length).toBeGreaterThan(0);
        expect(userInfo.permissions.length).toBeGreaterThan(0);
        const routes = await getAccessibleMenus(impersonatedApi);
        expect(routes.list.length).toBeGreaterThan(0);
      } finally {
        await impersonatedApi.dispose();
      }
      expect(
        scalarNumber(
          `SELECT COUNT(1) FROM sys_online_session WHERE tenant_id = ${tenant.id} AND user_id = ${out.actingUserId};`,
        ),
      ).toBeGreaterThan(0);
      await expectSuccess(
        await api.post(tenantCoreApiPath(`platform/tenants/${tenant.id}/end-impersonate`), {
          headers: { Authorization: `Bearer ${out.token}` },
        }),
      );
    });
  });
}

export async function scenarioTC0188() {
  await withAdmin(async ({ api, suffix }) => {
    const tenantA = await createNamedTenant(api, suffix, "tc188-a");
    const tenantB = await createNamedTenant(api, suffix, "tc188-b");
    const prefix = `tc188_${suffix}`;
    try {
      cleanupRowsByPrefix(prefix);
      const roleA = insertTenantRole(tenantA.id, suffix, `${prefix}_a`);
      const roleB = insertTenantRole(tenantB.id, suffix, `${prefix}_b`);
      const platformRole = insertTenantRole(0, suffix, `${prefix}_platform`);
      expect(roleA).toBeGreaterThan(0);
      expect(roleB).toBeGreaterThan(0);
      expect(platformRole).toBeGreaterThan(0);
      expect(
        scalarNumber(
          `SELECT COUNT(1) FROM sys_role WHERE tenant_id = ${tenantA.id} AND data_scope = 2 AND key LIKE '${pgEscapeLiteral(prefix)}%';`,
        ),
      ).toBe(1);
      expect(
        scalarNumber(
          `SELECT COUNT(1) FROM sys_role WHERE tenant_id = ${tenantB.id} AND data_scope = 2 AND key LIKE '${pgEscapeLiteral(prefix)}%';`,
        ),
      ).toBe(1);
      expect(
        scalarNumber(
          `SELECT COUNT(1) FROM sys_role WHERE tenant_id = 0 AND data_scope = 1 AND key LIKE '${pgEscapeLiteral(prefix)}%';`,
        ),
      ).toBe(1);
    } finally {
      cleanupRowsByPrefix(prefix);
      await deleteTenant(api, tenantA.id).catch(() => {});
      await deleteTenant(api, tenantB.id).catch(() => {});
    }
  });
}

export async function scenarioTC0189() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc189", async (tenant) => {
      const prefix = `tc189_${suffix}`;
      try {
        cleanupRowsByPrefix(prefix);
        execPgSQL(`
          INSERT INTO sys_dict_type (tenant_id, name, type, status, is_builtin, allow_tenant_override, remark, created_at, updated_at)
          VALUES (0, 'TC189 Platform', '${pgEscapeLiteral(prefix)}', 1, 0, TRUE, '', NOW(), NOW());
          INSERT INTO sys_dict_data (tenant_id, dict_type, label, value, sort, status, is_builtin, created_at, updated_at)
          VALUES (0, '${pgEscapeLiteral(prefix)}', 'platform', 'p', 1, 1, 0, NOW(), NOW());
          INSERT INTO sys_dict_type (tenant_id, name, type, status, is_builtin, allow_tenant_override, remark, created_at, updated_at)
          VALUES (${tenant.id}, 'TC189 Tenant', '${pgEscapeLiteral(prefix)}', 1, 0, TRUE, '', NOW(), NOW());
          INSERT INTO sys_dict_data (tenant_id, dict_type, label, value, sort, status, is_builtin, created_at, updated_at)
          VALUES (${tenant.id}, '${pgEscapeLiteral(prefix)}', 'tenant', 't', 1, 1, 0, NOW(), NOW());
        `);
        expect(
          queryPgRows(
            `SELECT label FROM sys_dict_data WHERE tenant_id = ${tenant.id} AND dict_type = '${pgEscapeLiteral(prefix)}';`,
          ),
        ).toEqual(["tenant"]);
        expect(
          queryPgRows(
            `SELECT label FROM sys_dict_data WHERE tenant_id = 0 AND dict_type = '${pgEscapeLiteral(prefix)}';`,
          ),
        ).toEqual(["platform"]);
      } finally {
        cleanupRowsByPrefix(prefix);
      }
    });
  });
}

export async function scenarioTC0190() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc190", async (tenant) => {
      const key = `tc190.${suffix}`;
      try {
        cleanupRowsByPrefix("tc190.");
        execPgSQL(`
          INSERT INTO sys_config (tenant_id, name, key, value, is_builtin, remark, created_at, updated_at)
          VALUES (0, 'TC190 Platform', '${pgEscapeLiteral(key)}', 'platform', 0, '', NOW(), NOW());
          INSERT INTO sys_config (tenant_id, name, key, value, is_builtin, remark, created_at, updated_at)
          VALUES (${tenant.id}, 'TC190 Tenant', '${pgEscapeLiteral(key)}', 'tenant', 0, '', NOW(), NOW());
        `);
        expect(
          queryPgScalar(
            `SELECT value FROM sys_config WHERE tenant_id = ${tenant.id} AND key = '${pgEscapeLiteral(key)}';`,
          ),
        ).toBe("tenant");
        expect(
          queryPgScalar(
            `SELECT value FROM sys_config WHERE tenant_id = 0 AND key = '${pgEscapeLiteral(key)}';`,
          ),
        ).toBe("platform");
      } finally {
        cleanupRowsByPrefix("tc190.");
      }
    });
  });
}

export async function scenarioTC0191() {
  await withAdmin(async ({ api, suffix }) => {
    const tenantA = await createNamedTenant(api, suffix, "tc191-a");
    const tenantB = await createNamedTenant(api, suffix, "tc191-b");
    const prefix = `tc191_${suffix}`;
    try {
      cleanupRowsByPrefix(prefix);
      execPgSQL(`
        INSERT INTO sys_file (tenant_id, name, original, suffix, scene, size, hash, url, path, engine, created_at, updated_at)
        VALUES
          (${tenantA.id}, '${pgEscapeLiteral(prefix)}_a', 'a.txt', '.txt', 'other', 1, '${pgEscapeLiteral(prefix)}a', '/storage/t/${tenantA.id}/a.txt', '/storage/t/${tenantA.id}/a.txt', 'local', NOW(), NOW()),
          (${tenantB.id}, '${pgEscapeLiteral(prefix)}_b', 'b.txt', '.txt', 'other', 1, '${pgEscapeLiteral(prefix)}b', '/storage/t/${tenantB.id}/b.txt', '/storage/t/${tenantB.id}/b.txt', 'local', NOW(), NOW()),
          (0, '${pgEscapeLiteral(prefix)}_p', 'p.txt', '.txt', 'other', 1, '${pgEscapeLiteral(prefix)}p', '/storage/platform/p.txt', '/storage/platform/p.txt', 'local', NOW(), NOW());
      `);
      expect(
        queryPgScalar(
          `SELECT path FROM sys_file WHERE tenant_id = ${tenantA.id} AND name = '${pgEscapeLiteral(prefix)}_a';`,
        ),
      ).toContain(`/t/${tenantA.id}/`);
      expect(
        queryPgScalar(
          `SELECT path FROM sys_file WHERE tenant_id = 0 AND name = '${pgEscapeLiteral(prefix)}_p';`,
        ),
      ).toContain("/platform/");
    } finally {
      cleanupRowsByPrefix(prefix);
      await deleteTenant(api, tenantA.id).catch(() => {});
      await deleteTenant(api, tenantB.id).catch(() => {});
    }
  });
}

export async function scenarioTC0192() {
  await withAdmin(async ({ api, suffix }) => {
    const tenantA = await createNamedTenant(api, suffix, "tc192-a");
    const tenantB = await createNamedTenant(api, suffix, "tc192-b");
    const prefix = `tc192_${suffix}`;
    try {
      cleanupRowsByPrefix(prefix);
      execPgSQL(`
        INSERT INTO sys_notify_message (tenant_id, plugin_id, source_type, source_id, category_code, title, content, payload_json, sender_user_id, created_at)
        VALUES
          (${tenantA.id}, '', 'system', '${pgEscapeLiteral(prefix)}_a', 'notice', 'A', 'A', '{}', 0, NOW()),
          (${tenantB.id}, '', 'system', '${pgEscapeLiteral(prefix)}_b', 'notice', 'B', 'B', '{}', 0, NOW()),
          (0, '', 'system', '${pgEscapeLiteral(prefix)}_platform', 'notice', 'P', 'P', '{}', 0, NOW());
      `);
      assertTenantScopedRowCounts(
        "sys_notify_message",
        tenantA.id,
        tenantB.id,
        "notify",
      );
      expect(
        scalarNumber(
          `SELECT COUNT(1) FROM sys_notify_message WHERE tenant_id = 0 AND source_id = '${pgEscapeLiteral(prefix)}_platform';`,
        ),
      ).toBe(1);
    } finally {
      cleanupRowsByPrefix(prefix);
      await deleteTenant(api, tenantA.id).catch(() => {});
      await deleteTenant(api, tenantB.id).catch(() => {});
    }
  });
}

export async function scenarioTC0193() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc193", async (tenant) => {
      const prefix = `tc193_${suffix}`;
      try {
        cleanupRowsByPrefix(prefix);
        execPgSQL(`
          INSERT INTO sys_job_group (tenant_id, code, name, sort_order, is_default, created_at, updated_at)
          VALUES (${tenant.id}, '${pgEscapeLiteral(prefix)}', 'TC193 Group', 1, 0, NOW(), NOW());
          INSERT INTO sys_job (tenant_id, group_id, name, description, task_type, handler_ref, params, timeout_seconds, cron_expr, timezone, scope, concurrency, status, is_builtin, created_at, updated_at)
          SELECT ${tenant.id}, id, '${pgEscapeLiteral(prefix)}_job', '', 'handler', 'host:cleanup-job-logs', '{}', 300, '0 0 1 1 *', 'Asia/Shanghai', 'master_only', 'singleton', 'disabled', 0, NOW(), NOW()
          FROM sys_job_group WHERE tenant_id = ${tenant.id} AND code = '${pgEscapeLiteral(prefix)}';
          INSERT INTO sys_job (tenant_id, group_id, name, description, task_type, handler_ref, params, timeout_seconds, cron_expr, timezone, scope, concurrency, status, is_builtin, created_at, updated_at)
          VALUES (0, 0, '${pgEscapeLiteral(prefix)}_platform_job', '', 'handler', 'host:cleanup-job-logs', '{}', 300, '0 0 1 1 *', 'Asia/Shanghai', 'master_only', 'singleton', 'disabled', 1, NOW(), NOW());
        `);
        expect(
          scalarNumber(
            `SELECT COUNT(1) FROM sys_job WHERE tenant_id = ${tenant.id} AND is_builtin = 0 AND name = '${pgEscapeLiteral(prefix)}_job';`,
          ),
        ).toBe(1);
        expect(
          scalarNumber(
            `SELECT COUNT(1) FROM sys_job WHERE tenant_id = 0 AND is_builtin = 1 AND name = '${pgEscapeLiteral(prefix)}_platform_job';`,
          ),
        ).toBe(1);
      } finally {
        cleanupRowsByPrefix(prefix);
      }
    });
  });
}

export async function scenarioTC0194() {
  await withAdmin(async ({ api, suffix }) => {
    const tenantA = await createNamedTenant(api, suffix, "tc194-a");
    const tenantB = await createNamedTenant(api, suffix, "tc194-b");
    const prefix = `tc194_${suffix}`;
    try {
      cleanupRowsByPrefix(prefix);
      execPgSQL(`
        INSERT INTO sys_online_session (tenant_id, token_id, user_id, username, login_time, last_active_time)
        VALUES
          (${tenantA.id}, '${pgEscapeLiteral(prefix)}_a', 1, 'admin', NOW(), NOW()),
          (${tenantB.id}, '${pgEscapeLiteral(prefix)}_b', 1, 'admin', NOW(), NOW());
      `);
      assertTenantScopedRowCounts(
        "sys_online_session",
        tenantA.id,
        tenantB.id,
        "online session",
      );
      execPgSQL(
        `DELETE FROM sys_online_session WHERE tenant_id = ${tenantA.id} AND token_id = '${pgEscapeLiteral(prefix)}_a';`,
      );
      expect(
        scalarNumber(
          `SELECT COUNT(1) FROM sys_online_session WHERE tenant_id = ${tenantA.id} AND token_id = '${pgEscapeLiteral(prefix)}_a';`,
        ),
      ).toBe(0);
      expect(
        scalarNumber(
          `SELECT COUNT(1) FROM sys_online_session WHERE tenant_id = ${tenantB.id} AND token_id = '${pgEscapeLiteral(prefix)}_b';`,
        ),
      ).toBe(1);
    } finally {
      cleanupRowsByPrefix(prefix);
      await deleteTenant(api, tenantA.id).catch(() => {});
      await deleteTenant(api, tenantB.id).catch(() => {});
    }
  });
}

export async function scenarioTC0195() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc195", async (tenant) => {
      const prefix = `tc195_${suffix}`;
      try {
        cleanupRowsByPrefix(prefix);
        execPgSQL(`
          INSERT INTO plugin_linapro_monitor_loginlog (tenant_id, acting_user_id, on_behalf_of_tenant_id, is_impersonation, user_name, status, ip, browser, os, msg, login_time)
          VALUES (${tenant.id}, 1, ${tenant.id}, TRUE, '${pgEscapeLiteral(prefix)}_user', 0, '127.0.0.1', 'e2e', 'e2e', 'impersonation', NOW());
          INSERT INTO plugin_linapro_monitor_operlog (tenant_id, acting_user_id, on_behalf_of_tenant_id, is_impersonation, title, oper_summary, route_owner, route_method, route_path, route_doc_key, oper_type, method, request_method, oper_name, oper_url, oper_ip, oper_param, json_result, status, error_msg, cost_time, oper_time)
          VALUES (${tenant.id}, 1, ${tenant.id}, TRUE, 'TC195', 'impersonation', 'core', 'POST', '/platform/tenants/${tenant.id}/impersonate', '', 'other', '', 'POST', '${pgEscapeLiteral(prefix)}_oper', '', '127.0.0.1', '{}', '{}', 0, '', 1, NOW());
        `);
        expect(
          scalarNumber(
            `SELECT COUNT(1) FROM plugin_linapro_monitor_loginlog WHERE tenant_id = ${tenant.id} AND is_impersonation = TRUE AND on_behalf_of_tenant_id = ${tenant.id};`,
          ),
        ).toBeGreaterThan(0);
        expect(
          scalarNumber(
            `SELECT COUNT(1) FROM plugin_linapro_monitor_operlog WHERE tenant_id = ${tenant.id} AND is_impersonation = TRUE AND acting_user_id = 1;`,
          ),
        ).toBeGreaterThan(0);
      } finally {
        cleanupRowsByPrefix(prefix);
      }
    });
  });
}

export async function scenarioTC0196() {
  await withAdmin(async ({ api, suffix }) => {
    await ensurePluginInstalledAndEnabled(api, "linapro-org-core");
    await withTenant(api, suffix, "tc196", async (tenant) => {
      const prefix = `tc196_${suffix}`;
      try {
        cleanupRowsByPrefix(prefix);
        execPgSQL(`
          INSERT INTO plugin_linapro_org_core_dept (tenant_id, parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
          VALUES (${tenant.id}, 0, '', 'TC196 Root', '${pgEscapeLiteral(prefix)}', 1, 1, NOW(), NOW());
        `);
        expect(
          scalarNumber(
            `SELECT COUNT(1) FROM plugin_linapro_org_core_dept WHERE tenant_id = ${tenant.id} AND code = '${pgEscapeLiteral(prefix)}';`,
          ),
        ).toBe(1);
        expect(tableExists("plugin_linapro_tenant_core_event_outbox")).toBeFalsy();
      } finally {
        cleanupRowsByPrefix(prefix);
      }
    });
  });
}

export async function scenarioTC0197() {
  await withAdmin(async ({ api, suffix }) => {
    await ensurePluginInstalledAndEnabled(api, "linapro-org-core");
    const tenantA = await createNamedTenant(api, suffix, "tc197-a");
    const tenantB = await createNamedTenant(api, suffix, "tc197-b");
    const prefix = `tc197_${suffix}`;
    try {
      cleanupRowsByPrefix(prefix);
      execPgSQL(`
        INSERT INTO plugin_linapro_org_core_post (tenant_id, code, name, sort, status, created_at, updated_at)
        VALUES
          (${tenantA.id}, '${pgEscapeLiteral(prefix)}', 'TC197 A', 1, 1, NOW(), NOW()),
          (${tenantB.id}, '${pgEscapeLiteral(prefix)}', 'TC197 B', 1, 1, NOW(), NOW());
      `);
      expect(
        scalarNumber(
          `SELECT COUNT(1) FROM plugin_linapro_org_core_post WHERE code = '${pgEscapeLiteral(prefix)}';`,
        ),
      ).toBe(2);
      expect(
        scalarNumber(
          `SELECT COUNT(1) FROM plugin_linapro_org_core_post WHERE tenant_id = ${tenantA.id} AND code = '${pgEscapeLiteral(prefix)}';`,
        ),
      ).toBe(1);
      expect(
        scalarNumber(
          `SELECT COUNT(1) FROM plugin_linapro_org_core_post WHERE tenant_id = ${tenantB.id} AND code = '${pgEscapeLiteral(prefix)}';`,
        ),
      ).toBe(1);
    } finally {
      cleanupRowsByPrefix(prefix);
      await deleteTenant(api, tenantA.id).catch(() => {});
      await deleteTenant(api, tenantB.id).catch(() => {});
    }
  });
}

export async function scenarioTC0198() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc198", async (tenant) => {
      const user = await addTenantUser(api, suffix, "tc198_user", tenant.id);
      const grants: TenantUserGrant[] = [];
      try {
        grants.push(
          await grantTenantPermissions(api, {
            roleKey: `tc198-user-query-${suffix}`,
            roleName: `TC198 User Query ${suffix}`,
            tenantId: tenant.id,
            userId: user.id,
            permissions: ["system:user:query"],
          }),
        );
        const login = await loginRaw(user.username, password);
        expect(login.accessToken).toBeTruthy();
        const tenantApi = await createTenantApiContext(login.accessToken!);
        try {
          await expectUserListContains(tenantApi, tenant.id, user.username);
          expectMembershipRow(user.id, tenant.id);
        } finally {
          await tenantApi.dispose();
        }
      } finally {
        revokeTenantPermissionGrants(grants);
        await cleanupTenantUser(api, user.id, user.memberId);
      }
    });
  });
}

export async function scenarioTC0199() {
  await withAdmin(async ({ api }) => {
    expectBuiltInResolverPolicyRemainsCodeOwned();
    await expectBusinessError(
      await api.post(tenantCoreApiPath("platform/tenants"), {
        data: { code: "www", name: "Reserved" },
      }),
    );
  });
}

export async function scenarioTC0200() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc200", async (tenant) => {
      const user = await addTenantUser(api, suffix, "tc200_user", tenant.id);
      const grants: TenantUserGrant[] = [];
      try {
        grants.push(
          await grantTenantPermissions(api, {
            roleKey: `tc200-user-query-${suffix}`,
            roleName: `TC200 User Query ${suffix}`,
            tenantId: tenant.id,
            userId: user.id,
            permissions: ["system:user:query"],
          }),
        );
        const token = await loginAndSelect(user.username, tenant.id);
        const tenantApi = await createTenantApiContext(token);
        try {
          await expectUserListContains(tenantApi, tenant.id, user.username);
          expectMembershipRow(user.id, tenant.id);
        } finally {
          await tenantApi.dispose();
        }
      } finally {
        revokeTenantPermissionGrants(grants);
        await cleanupTenantUser(api, user.id, user.memberId);
      }
    });
  });
}

export async function scenarioTC0201() {
  await scenarioSwitchToken("tc201");
}

export async function scenarioTC0202() {
  await withAdmin(async ({ api, suffix }) => {
    const tenantA = await createNamedTenant(api, suffix, "tc202-a");
    const tenantB = await createNamedTenant(api, suffix, "tc202-b");
    const user = await createTenantUser(api, suffix, "tc202_user", tenantA.id);
    let memberA = 0;
    let memberB = 0;
    try {
      memberA = (
        await addTenantMember(api, { tenantId: tenantA.id, userId: user.id })
      ).id;
      memberB = (
        await addTenantMember(api, { tenantId: tenantB.id, userId: user.id })
      ).id;
      const login = await loginRaw(user.username, password);
      expect(login.preToken).toBeTruthy();
      expect(login.tenants?.map((tenant) => tenant.id)).toEqual([
        tenantA.id,
        tenantB.id,
      ]);
    } finally {
      await cleanupTenantUser(api, user.id, memberA);
      if (memberB > 0) {
        await removeTenantMember(api, memberB).catch(() => {});
      }
      await deleteTenant(api, tenantA.id).catch(() => {});
      await deleteTenant(api, tenantB.id).catch(() => {});
    }
  });
}

export async function scenarioTC0203() {
  await withAdmin(async ({ api }) => {
    expectBuiltInResolverPolicyRemainsCodeOwned();
    await expectBusinessError(
      await api.post("auth/login", {
        data: { username: "missing-tenant-user", password },
      }),
    );
  });
}

export async function scenarioTC0204() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc204", async (tenant) => {
      const out = await expectSuccess<{
        tenantId: number;
        isImpersonated: boolean;
      }>(
        await api.post(tenantCoreApiPath(`platform/tenants/${tenant.id}/impersonate`), {
          data: { reason: "TC007 override" },
        }),
      );
      expect(out.tenantId).toBe(tenant.id);
      expect(out.isImpersonated).toBeTruthy();
      const user = await addTenantUser(api, suffix, "tc204_user", tenant.id);
      const grants: TenantUserGrant[] = [];
      try {
        grants.push(
          await grantTenantPermissions(api, {
            roleKey: `tc204-user-query-${suffix}`,
            roleName: `TC204 User Query ${suffix}`,
            tenantId: tenant.id,
            userId: user.id,
            permissions: ["system:user:query"],
          }),
        );
        const token = await loginAndSelect(user.username, tenant.id);
        const tenantApi = await createTenantApiContext(token);
        try {
          await expectBusinessError(
            await tenantApi.post(tenantCoreApiPath("auth/switch-tenant"), {
              data: { tenantId: 999999 },
            }),
          );
        } finally {
          await tenantApi.dispose();
        }
      } finally {
        revokeTenantPermissionGrants(grants);
        await cleanupTenantUser(api, user.id, user.memberId);
      }
    });
  });
}

export async function scenarioTC0205() {
  await withAdmin(async ({ api }) => {
    expectBuiltInResolverPolicyRemainsCodeOwned();
    await expectResolverConfigEndpointRemoved(api);
  });
}

export async function scenarioTC0206() {
  await withAdmin(async ({ api }) => {
    await ensurePluginInstalled(api, "linapro-monitor-loginlog");
    const plugin = pluginRow("linapro-monitor-loginlog");
    expect(plugin.scopeNature).toBe("tenant_aware");
    expect(["global", "tenant_scoped"]).toContain(plugin.installMode);
  });
}

export async function scenarioTC0207() {
  await withAdmin(async ({ api }) => {
    await syncPlugins(api);
    const plugin = pluginRow("linapro-tenant-core");
    expect(plugin.scopeNature).toBe("platform_only");
    expect(plugin.installMode).toBe("global");
    await expectBusinessError(
      await api.get(tenantCoreApiPath("tenant/plugins")),
    );
  });
}

export async function scenarioTC0208() {
  await withAdmin(async ({ api, suffix }) => {
    await ensurePluginInstalled(api, "linapro-monitor-loginlog");
    await withTenant(api, suffix, "tc208", async (tenant) => {
      execPgSQL(
        `UPDATE sys_plugin SET install_mode = 'tenant_scoped', status = 1 WHERE plugin_id = 'linapro-monitor-loginlog';`,
      );
      const tenantPlugin = await createTenantPluginApi(api, tenant.id, suffix);
      try {
        const list = await expectSuccess<{
          list: Array<{ id: string; tenantEnabled: number }>;
        }>(await tenantPlugin.api.get(tenantCoreApiPath("tenant/plugins")));
        expect(list.list.map((item) => item.id)).toContain("linapro-monitor-loginlog");
        await expectSuccess(
          await tenantPlugin.api.post(
            tenantCoreApiPath("tenant/plugins/linapro-monitor-loginlog/enable"),
          ),
        );
        expect(
          scalarNumber(
            `SELECT COUNT(1) FROM sys_plugin_state WHERE plugin_id = 'linapro-monitor-loginlog' AND tenant_id = ${tenant.id} AND state_key = '${tenantEnablementKey}' AND enabled = TRUE;`,
          ),
        ).toBe(1);
        await expectSuccess(
          await tenantPlugin.api.post(
            tenantCoreApiPath("tenant/plugins/linapro-monitor-loginlog/disable"),
          ),
        );
        expect(
          scalarNumber(
            `SELECT COUNT(1) FROM sys_plugin_state WHERE plugin_id = 'linapro-monitor-loginlog' AND tenant_id = ${tenant.id} AND state_key = '${tenantEnablementKey}' AND enabled = FALSE;`,
          ),
        ).toBe(1);
      } finally {
        await tenantPlugin.api.dispose();
        revokeTenantPermissionGrants(tenantPlugin.grants);
        await cleanupTenantUser(
          api,
          tenantPlugin.user.id,
          tenantPlugin.user.memberId,
        );
        execPgSQL(
          `DELETE FROM sys_plugin_state WHERE plugin_id = 'linapro-monitor-loginlog' AND tenant_id = ${tenant.id};`,
        );
      }
    });
  });
}

export async function scenarioTC0237() {
  const tenantScopedMonitorPluginIds = [
    "linapro-monitor-loginlog",
    "linapro-monitor-online",
    "linapro-monitor-operlog",
  ];
  const monitorPluginIds = [...tenantScopedMonitorPluginIds, "linapro-monitor-server"];
  const monitorMenuPaths = [
    "/monitor/loginlog",
    "/monitor/online",
    "/monitor/operlog",
    "/monitor/server",
  ];
  await withAdmin(async ({ api, suffix }) => {
    await syncPlugins(api);
    const snapshots = monitorPluginIds.map((pluginId) =>
      monitorPluginSnapshot(pluginId),
    );
    try {
      for (const pluginId of monitorPluginIds) {
        await ensurePluginInstalledAndEnabled(api, pluginId);
      }
      await withTenant(api, suffix, "tc237", async (tenant) => {
        removeTenantMonitorStates(tenant.id, tenantScopedMonitorPluginIds);
        setMonitorPluginsAutoEnabledForTenant(
          tenantScopedMonitorPluginIds,
          tenant.id,
        );
        const user = await addTenantUser(
          api,
          suffix,
          `monitor_menu_${tenant.id}`,
          tenant.id,
        );
        let grant: TenantUserGrant | undefined;
        try {
          grant = await grantTenantPermissions(api, {
            roleKey: `monitor_menu_${tenant.id}_${suffix}`,
            roleName: `Monitor menu ${tenant.id} ${suffix}`,
            tenantId: tenant.id,
            userId: user.id,
            permissions: [
              "monitor:loginlog:list",
              "monitor:online:list",
              "monitor:operlog:list",
              "monitor:server:list",
            ],
          });
          const token = await loginAndSelect(user.username, tenant.id);
          const tenantApi = await createTenantApiContext(token);
          try {
            const routes = await getAccessibleMenus(tenantApi);
            assertAccessibleMenuPaths(routes.list, monitorMenuPaths);
          } finally {
            await tenantApi.dispose();
          }
        } finally {
          revokeTenantPermissionGrants(grant ? [grant] : []);
          await cleanupTenantUser(api, user.id, user.memberId);
          removeTenantMonitorStates(tenant.id, tenantScopedMonitorPluginIds);
        }
      });
    } finally {
      for (const snapshot of snapshots) {
        restoreMonitorPluginSnapshot(snapshot);
      }
    }
  });
}

export async function scenarioTC0239() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc239", async (tenant) => {
      const user = await addTenantUser(api, suffix, "tc239_user", tenant.id);
      const grants: TenantUserGrant[] = [];
      try {
        const blockedMenuIDs = menuIDsByPermissions([
          "system:tenant:list",
          "plugin:install",
          "system:menu:add",
        ]);
        grants.push(
          await grantTenantPermissions(api, {
            roleKey: `tc239-dirty-${suffix}`,
            roleName: `TC239 Dirty ${suffix}`,
            tenantId: tenant.id,
            userId: user.id,
            permissions: [
              "system:tenant:list",
              "system:tenant:add",
              "system:role:add",
              "system:role:edit",
              "system:menu:query",
            ],
          }),
        );
        const token = await loginAndSelect(user.username, tenant.id);
        const tenantApi = await createTenantApiContext(token);
        try {
          await expectBusinessErrorCode(
            await tenantApi.get(
              tenantCoreApiPath("platform/tenants?pageNum=1&pageSize=10"),
            ),
            "MULTI_TENANT_PLATFORM_PERMISSION_REQUIRED",
          );
          const tree = await expectSuccess<RoleMenuTreeResponse>(
            await tenantApi.get(`menu/role/${grants[0]!.roleId}`),
          );
          expectRoleTreeExcludesMenuIDs(tree, blockedMenuIDs);
          await expectBusinessErrorCode(
            await tenantApi.post("role", {
              data: {
                name: `TC239 Create`.slice(0, 30),
                key: `tc239_create_${suffix}`.slice(0, 30),
                sort: 1,
                dataScope: 2,
                status: 1,
                menuIds: [blockedMenuIDs[0]!],
              },
            }),
            "ROLE_MENU_ASSIGNMENT_FORBIDDEN",
          );
          await expectBusinessErrorCode(
            await tenantApi.put(`role/${grants[0]!.roleId}`, {
              data: {
                name: `TC239 Update`.slice(0, 30),
                key: `tc239_dirty_${suffix}`.slice(0, 30),
                sort: 1,
                dataScope: 2,
                status: 1,
                menuIds: [blockedMenuIDs[1]!],
              },
            }),
            "ROLE_MENU_ASSIGNMENT_FORBIDDEN",
          );
        } finally {
          await tenantApi.dispose();
        }
      } finally {
        revokeTenantPermissionGrants(grants);
        await cleanupTenantUser(api, user.id, user.memberId);
      }
    });
  });
}

export async function scenarioTC0240() {
  await withAdmin(async ({ api, suffix }) => {
    await ensurePluginInstalled(api, "linapro-org-core");
    await withTenant(api, suffix, "tc240", async (tenant) => {
      const user = await addTenantUser(api, suffix, "tc240_user", tenant.id);
      const grants: TenantUserGrant[] = [];
      const prefix = `tc240_${suffix}`;
      try {
        const blockedMenuIDs = menuIDsByPermissions([
          "system:menu:add",
          "plugin:install",
        ]);
        grants.push(
          await grantTenantPermissions(api, {
            roleKey: `tc240-governance-${suffix}`,
            roleName: `TC240 Governance ${suffix}`,
            tenantId: tenant.id,
            userId: user.id,
            permissions: [
              "system:menu:query",
              "system:menu:add",
              "system:menu:edit",
              "system:menu:remove",
              "plugin:list",
              "plugin:query",
              "plugin:install",
              "plugin:enable",
              "plugin:disable",
              "plugin:edit",
              "plugin:uninstall",
            ],
          }),
        );
        const token = await loginAndSelect(user.username, tenant.id);
        const tenantApi = await createTenantApiContext(token);
        try {
          const tree = await expectSuccess<RoleMenuTreeResponse>(
            await tenantApi.get(`menu/role/${grants[0]!.roleId}`),
          );
          expectRoleTreeExcludesMenuIDs(tree, blockedMenuIDs);
          const routes = await getAccessibleMenus(tenantApi);
          assertNoRoutePath(routes.list, "/platform/tenants");
          await expectBusinessErrorCode(
            await tenantApi.post("menu", {
              data: {
                parentId: 0,
                name: `TC240 ${suffix}`,
                path: `tc240-${suffix}`,
                component: "system/user/index",
                perms: `${prefix}:menu:list`,
                icon: `lucide:test-tube-${suffix}`,
                type: "M",
                sort: 999,
                visible: 1,
                status: 1,
                isFrame: 0,
                isCache: 0,
                remark: "linapro-tenant-core e2e",
              },
            }),
            "PLATFORM_PERMISSION_REQUIRED",
          );
          await expectBusinessErrorCode(
            await tenantApi.post("plugins/sync"),
            "PLATFORM_PERMISSION_REQUIRED",
          );
          await expectBusinessErrorCode(
            await tenantApi.put("plugins/linapro-org-core/tenant-provisioning-policy", {
              data: { autoEnableForNewTenants: true },
            }),
            "PLATFORM_PERMISSION_REQUIRED",
          );
          await expectBusinessErrorCode(
            await tenantApi.put("plugins/linapro-org-core/enable"),
            "PLATFORM_PERMISSION_REQUIRED",
          );
          expect(
            scalarNumber(
              `SELECT COUNT(1) FROM sys_menu WHERE perms = '${pgEscapeLiteral(`${prefix}:menu:list`)}';`,
            ),
          ).toBe(0);
        } finally {
          await tenantApi.dispose();
        }
      } finally {
        revokeTenantPermissionGrants(grants);
        cleanupRowsByPrefix(prefix);
        await cleanupTenantUser(api, user.id, user.memberId);
      }
    });
  });
}

export async function scenarioTC0241() {
  await withAdmin(async ({ api, suffix }) => {
    const tenantA = await createNamedTenant(api, suffix, "tc241-a");
    const tenantB = await createNamedTenant(api, suffix, "tc241-b");
    const user = await addTenantUser(api, suffix, "tc241_user", tenantA.id);
    const grants: TenantUserGrant[] = [];
    const prefix = `tc241_${suffix}`;
      try {
        cleanupRowsByPrefix(prefix);
        insertTenantDefaultJobGroup(tenantA.id);
        insertTenantDefaultJobGroup(tenantB.id);
        grants.push(
          await grantTenantPermissions(api, {
          roleKey: `tc241-jobgroup-${suffix}`,
          roleName: `TC241 Job Group ${suffix}`,
          tenantId: tenantA.id,
          userId: user.id,
          permissions: [
            "system:jobgroup:list",
            "system:jobgroup:add",
            "system:jobgroup:edit",
            "system:jobgroup:remove",
            "system:job:list",
            "system:job:add",
          ],
        }),
      );
      const tenantADefaultGroup = scalarNumber(
        `SELECT id FROM sys_job_group WHERE tenant_id = ${tenantA.id} AND code = 'default' LIMIT 1;`,
      );
      const tenantBGroup = insertJobGroup(
        tenantB.id,
        `${prefix}_tenant_b`,
        "TC241 Tenant B",
      );
      const platformGroup = insertJobGroup(
        0,
        `${prefix}_platform`,
        "TC241 Platform",
      );
      const token = await loginAndSelect(user.username, tenantA.id);
      const tenantApi = await createTenantApiContext(token);
      try {
        const created = await expectSuccess<{ id: number }>(
          await tenantApi.post("job-group", {
            data: {
              code: `${prefix}_tenant_a`,
              name: "TC241 Tenant A",
              remark: "linapro-tenant-core e2e",
              sortOrder: 20,
            },
          }),
        );
        expect(
          scalarNumber(
            `SELECT tenant_id FROM sys_job_group WHERE id = ${created.id};`,
          ),
        ).toBe(tenantA.id);
        insertShellJob(tenantA.id, created.id, `${prefix}_tenant_a_job`);
        const tenantBJob = insertShellJob(
          tenantB.id,
          tenantBGroup,
          `${prefix}_tenant_b_job`,
        );
        insertShellJob(0, platformGroup, `${prefix}_platform_job`);
        const groups = await expectSuccess<{
          list: JobGroupFallbackItem[];
          total: number;
        }>(
          await tenantApi.get(
            `job-group?pageNum=1&pageSize=100&code=${encodeURIComponent(prefix)}`,
          ),
        );
        expect(groups.list.map((item) => item.id)).toContain(created.id);
        expect(groups.list.map((item) => item.id)).not.toContain(tenantBGroup);
        expect(groups.list.map((item) => item.id)).not.toContain(platformGroup);
        expect(groups.list.find((item) => item.id === created.id)?.jobCount).toBe(
          1,
        );
        await expectBusinessErrorCode(
          await tenantApi.put(`job-group/${tenantBGroup}`, {
            data: {
              code: `${prefix}_tenant_b_update`,
              name: "TC241 Tenant B Update",
              remark: "out of scope",
              sortOrder: 30,
            },
          }),
          "JOB_GROUP_NOT_FOUND",
        );
        await expectBusinessErrorCode(
          await tenantApi.delete(`job-group/${tenantBGroup}`),
          "JOB_GROUP_NOT_FOUND",
        );
        await expectSuccess(await tenantApi.delete(`job-group/${created.id}`));
        expect(
          scalarNumber(
            `SELECT group_id FROM sys_job WHERE tenant_id = ${tenantA.id} AND name = '${pgEscapeLiteral(`${prefix}_tenant_a_job`)}';`,
          ),
        ).toBe(tenantADefaultGroup);
        expect(
          scalarNumber(`SELECT group_id FROM sys_job WHERE id = ${tenantBJob};`),
        ).toBe(tenantBGroup);
      } finally {
        await tenantApi.dispose();
      }
    } finally {
      revokeTenantPermissionGrants(grants);
      cleanupRowsByPrefix(prefix);
      await cleanupTenantUser(api, user.id, user.memberId);
      await deleteTenant(api, tenantA.id).catch(() => {});
      await deleteTenant(api, tenantB.id).catch(() => {});
    }
  });
}

export async function scenarioTC0242(page: Page) {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc242", async (tenant) => {
      const user = await addTenantUser(api, suffix, "tc242_user", tenant.id);
      const grants: TenantUserGrant[] = [];
      const configKey = `tc242.${suffix}`;
      const dictType = `tc242_${suffix}`;
      try {
        cleanupRowsByPrefix("tc242.");
        cleanupRowsByPrefix("tc242_");
        grants.push(
          await grantTenantPermissions(api, {
            roleKey: `tc242-fallback-${suffix}`,
            roleName: `TC242 Fallback ${suffix}`,
            tenantId: tenant.id,
            userId: user.id,
            permissions: [
              "system:config:list",
              "system:config:query",
              "system:dict:list",
              "system:dict:query",
            ],
          }),
        );
        const configID = scalarNumber(`
          INSERT INTO sys_config (tenant_id, name, key, value, is_builtin, remark, created_at, updated_at)
          VALUES (0, 'TC242 Platform Config ${pgEscapeLiteral(suffix)}', '${pgEscapeLiteral(configKey)}', 'platform', 0, 'linapro-tenant-core e2e', NOW(), NOW())
          RETURNING id;
        `);
        const dictTypeID = scalarNumber(`
          INSERT INTO sys_dict_type (tenant_id, name, type, status, is_builtin, allow_tenant_override, remark, created_at, updated_at)
          VALUES (0, 'TC242 Platform Dict ${pgEscapeLiteral(suffix)}', '${pgEscapeLiteral(dictType)}', 1, 0, TRUE, 'linapro-tenant-core e2e', NOW(), NOW())
          RETURNING id;
        `);
        const dictDataID = scalarNumber(`
          INSERT INTO sys_dict_data (tenant_id, dict_type, label, value, sort, tag_style, css_class, status, is_builtin, remark, created_at, updated_at)
          VALUES (0, '${pgEscapeLiteral(dictType)}', 'TC242 Platform Data ${pgEscapeLiteral(suffix)}', 'platform', 1, '', '', 1, 0, 'linapro-tenant-core e2e', NOW(), NOW())
          RETURNING id;
        `);
        const token = await loginAndSelect(user.username, tenant.id);
        const tenantApi = await createTenantApiContext(token);
        try {
          const configList = await expectSuccess<{
            list: ConfigFallbackItem[];
          }>(
            await tenantApi.get(
              `config?pageNum=1&pageSize=10&key=${encodeURIComponent(configKey)}`,
            ),
          );
          const configRow = configList.list.find((item) => item.key === configKey);
          expect(configRow).toBeTruthy();
          expectFallbackMetadata(configRow!, "config");

          const dictTypeList = await expectSuccess<{
            list: DictTypeFallbackItem[];
          }>(
            await tenantApi.get(
              `dict/type?pageNum=1&pageSize=10&type=${encodeURIComponent(dictType)}`,
            ),
          );
          const dictTypeRow = dictTypeList.list.find(
            (item) => item.type === dictType,
          );
          expect(dictTypeRow).toBeTruthy();
          expectFallbackMetadata(dictTypeRow!, "dict type");

          const dictDataList = await expectSuccess<{
            list: DictDataFallbackItem[];
          }>(
            await tenantApi.get(
              `dict/data?pageNum=1&pageSize=10&dictType=${encodeURIComponent(dictType)}`,
            ),
          );
          const dictDataRow = dictDataList.list.find(
            (item) => item.dictType === dictType && item.value === "platform",
          );
          expect(dictDataRow).toBeTruthy();
          expectFallbackMetadata(dictDataRow!, "dict data");

          await expectBusinessError(await tenantApi.get(`config/${configID}`));
          await expectBusinessError(await tenantApi.get(`dict/type/${dictTypeID}`));
          await expectBusinessError(await tenantApi.get(`dict/data/${dictDataID}`));

          const detailRequests: string[] = [];
          page.on("request", (request) => {
            const url = new URL(request.url());
            if (request.method() !== "GET" || !url.pathname.includes("/api/v1/")) {
              return;
            }
            if (
              url.pathname.endsWith(`/config/${configID}`) ||
              url.pathname.endsWith(`/dict/type/${dictTypeID}`) ||
              url.pathname.endsWith(`/dict/data/${dictDataID}`)
            ) {
              detailRequests.push(url.pathname);
            }
          });
          await loginTenantUserInBrowser(page, user.username);

          await page.goto(workspacePath("/system/config"));
          await waitForRouteReady(page, 15000);
          await expect(page.getByText(configKey, { exact: false })).toBeVisible({
            timeout: 15000,
          });
          await expect(page.getByTestId(`config-delete-${configID}`)).toHaveCount(
            0,
          );

          await page.goto(workspacePath("/system/dict"));
          await waitForRouteReady(page, 15000);
          await expect(page.getByText(dictType, { exact: false })).toBeVisible({
            timeout: 15000,
          });
          await expect(
            page.getByTestId(`dict-type-delete-${dictTypeID}`),
          ).toHaveCount(0);
          await page.getByText(dictType, { exact: false }).first().click();
          await waitForRouteReady(page, 15000);
          await expect(
            page.getByText(`TC242 Platform Data ${suffix}`, { exact: false }),
          ).toBeVisible({ timeout: 15000 });
          await expect(
            page.getByTestId(`dict-data-delete-${dictDataID}`),
          ).toHaveCount(0);
          expect(detailRequests).toEqual([]);
        } finally {
          await tenantApi.dispose();
        }
      } finally {
        revokeTenantPermissionGrants(grants);
        await cleanupTenantUser(api, user.id, user.memberId);
        cleanupRowsByPrefix("tc242.");
        cleanupRowsByPrefix("tc242_");
      }
    });
  });
}

export async function scenarioTC0210() {
  await withAdmin(async ({ api, suffix }) => {
    await expectMultiTenantUninstallBlockedWithExistingTenant(
      api,
      suffix,
      "tc210",
    );
  });
}

export async function scenarioTC0211() {
  await withAdmin(async ({ api, suffix }) => {
    await expectMultiTenantForceUninstallBypassesGuard(api, suffix, "tc211");
  });
}

export async function scenarioTC0212() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc212", async (tenant) => {
      await expectSuccess(
        await api.delete(tenantCoreApiPath(`platform/tenants/${tenant.id}`)),
      );
      expect(tableExists("plugin_linapro_tenant_core_event_outbox")).toBeFalsy();
    });
  });
}

export async function scenarioTC0213() {
  await withAdmin(async ({ api, suffix }) => {
    await ensurePluginInstalledAndEnabled(api, "linapro-org-core");
    const orgCenter = pluginRow("linapro-org-core");
    expect(orgCenter.scopeNature).toBe("tenant_aware");
    const previousMode = orgCenter.installMode;
    const previousEnabled = orgCenter.enabled;
    const previousAutoEnableForNewTenants = orgCenter.autoEnableForNewTenants;
    try {
      execPgSQL(`
        UPDATE sys_plugin
        SET install_mode = 'tenant_scoped',
            status = 1,
            updated_at = NOW()
        WHERE plugin_id = 'linapro-org-core'
          AND scope_nature = 'tenant_aware';
      `);
      await expectSuccess(
        await api.put("plugins/linapro-org-core/tenant-provisioning-policy/", {
          data: { autoEnableForNewTenants: true },
        }),
      );
      await withTenant(api, suffix, "tc213", async (tenant) => {
        expect(
          scalarNumber(
            `SELECT COUNT(1) FROM sys_plugin_state WHERE plugin_id = 'linapro-org-core' AND tenant_id = ${tenant.id} AND state_key = '${tenantEnablementKey}' AND enabled = TRUE;`,
          ),
        ).toBeGreaterThan(0);
      });
    } finally {
      execPgSQL(`
        UPDATE sys_plugin
        SET install_mode = '${pgEscapeLiteral(previousMode)}',
            auto_enable_for_new_tenants = ${previousAutoEnableForNewTenants ? "TRUE" : "FALSE"},
            status = ${previousEnabled},
            updated_at = NOW()
        WHERE plugin_id = 'linapro-org-core';
      `);
      execPgSQL(
        `DELETE FROM sys_plugin_state WHERE plugin_id = 'linapro-org-core' AND state_key = '${tenantEnablementKey}' AND tenant_id > 0;`,
      );
    }
  });
}

export async function scenarioTC0214() {
  await withAdmin(async ({ api, suffix }) => {
    const tenantA = await createNamedTenant(api, suffix, "tc214-a");
    const tenantB = await createNamedTenant(api, suffix, "tc214-b");
    try {
      const tenants = await expectSuccess<{ list: Array<{ id: number }> }>(
        await api.get(tenantCoreApiPath("platform/tenants?pageNum=1&pageSize=100")),
      );
      expect(tenants.list.map((tenant) => tenant.id)).toEqual(
        expect.arrayContaining([tenantA.id, tenantB.id]),
      );
    } finally {
      await deleteTenant(api, tenantA.id).catch(() => {});
      await deleteTenant(api, tenantB.id).catch(() => {});
    }
  });
}

export async function scenarioTC0215() {
  await withAdmin(async ({ api, suffix }) => {
    await ensurePluginInstalledAndEnabled(api, "linapro-monitor-loginlog");
    await ensurePluginInstalledAndEnabled(api, "linapro-monitor-operlog");
    await withTenant(api, suffix, "tc215", async (tenant) => {
      await expectSuccess(
        await api.post(tenantCoreApiPath(`platform/tenants/${tenant.id}/impersonate`), {
          data: { reason: "TC003" },
        }),
      );
      expect(
        scalarNumber(
          `SELECT COUNT(1) FROM plugin_linapro_monitor_loginlog WHERE tenant_id = ${tenant.id} AND is_impersonation = TRUE AND on_behalf_of_tenant_id = ${tenant.id};`,
        ),
      ).toBeGreaterThan(0);
      expect(
        scalarNumber(
          `SELECT COUNT(1) FROM plugin_linapro_monitor_operlog WHERE tenant_id = ${tenant.id} AND is_impersonation = TRUE AND on_behalf_of_tenant_id = ${tenant.id};`,
        ),
      ).toBeGreaterThan(0);
    });
  });
}

export async function scenarioTC0216() {
  await withAdmin(async ({ api, suffix }) => {
    await expectMultiTenantForceUninstallBypassesGuard(api, suffix, "tc216");
  });
}

export async function scenarioTC0217() {
  await withAdmin(async ({ api, suffix }) => {
    await withTenant(api, suffix, "tc217", async (tenant) => {
      const adminMenus = await getAccessibleMenus(api);
      const tenantPlugin = await createTenantPluginApi(api, tenant.id, suffix);
      try {
        const tenantMenus = await getAccessibleMenus(tenantPlugin.api);
        const adminText = JSON.stringify(adminMenus.list);
        const tenantText = JSON.stringify(tenantMenus.list);
        expect(adminText).toContain("platform");
        expect(tenantText).not.toBe(adminText);
      } finally {
        await tenantPlugin.api.dispose();
        revokeTenantPermissionGrants(tenantPlugin.grants);
        await cleanupTenantUser(
          api,
          tenantPlugin.user.id,
          tenantPlugin.user.memberId,
        );
      }
    });
  });
}

export async function scenarioTC0218() {
  await withAdmin(async ({ api }) => {
    expectBuiltInResolverPolicyRemainsCodeOwned();
    await expectResolverConfigEndpointRemoved(api);
  });
}

export async function scenarioTC0219() {
  await withAdmin(async ({ api, suffix }) => {
    const tenantA = await createNamedTenant(api, suffix, "tc219-a");
    const tenantB = await createNamedTenant(api, suffix, "tc219-b");
    try {
      execPgSQL(`
        INSERT INTO sys_cache_revision (tenant_id, domain, scope, revision, reason, created_at, updated_at)
        VALUES
          (${tenantA.id}, 'permission-access', 'tenant:${tenantA.id}', 1, 'tc219', NOW(), NOW()),
          (${tenantB.id}, 'permission-access', 'tenant:${tenantB.id}', 1, 'tc219', NOW(), NOW())
        ON CONFLICT (tenant_id, domain, scope) DO UPDATE SET revision = sys_cache_revision.revision + 1, updated_at = NOW();
      `);
      const revA = scalarNumber(
        `SELECT revision FROM sys_cache_revision WHERE tenant_id = ${tenantA.id} AND scope = 'tenant:${tenantA.id}';`,
      );
      const revB = scalarNumber(
        `SELECT revision FROM sys_cache_revision WHERE tenant_id = ${tenantB.id} AND scope = 'tenant:${tenantB.id}';`,
      );
      expect(revA).toBeGreaterThan(0);
      expect(revB).toBeGreaterThan(0);
      execPgSQL(
        `UPDATE sys_cache_revision SET revision = revision + 1 WHERE tenant_id = ${tenantA.id} AND scope = 'tenant:${tenantA.id}';`,
      );
      expect(
        scalarNumber(
          `SELECT revision FROM sys_cache_revision WHERE tenant_id = ${tenantA.id} AND scope = 'tenant:${tenantA.id}';`,
        ),
      ).toBe(revA + 1);
      expect(
        scalarNumber(
          `SELECT revision FROM sys_cache_revision WHERE tenant_id = ${tenantB.id} AND scope = 'tenant:${tenantB.id}';`,
        ),
      ).toBe(revB);
    } finally {
      execPgSQL(
        `DELETE FROM sys_cache_revision WHERE tenant_id IN (${tenantA.id}, ${tenantB.id}) AND domain = 'permission-access';`,
      );
      await deleteTenant(api, tenantA.id).catch(() => {});
      await deleteTenant(api, tenantB.id).catch(() => {});
    }
  });
}

export async function scenarioTC0220() {
  await scenarioSwitchToken("tc220");
}

async function scenarioSwitchToken(prefix: string) {
  await withAdmin(async ({ api, suffix }) => {
    const tenantA = await createNamedTenant(api, suffix, `${prefix}-a`);
    const tenantB = await createNamedTenant(api, suffix, `${prefix}-b`);
    const user = await createTenantUser(
      api,
      suffix,
      `${prefix}_user`,
      tenantA.id,
    );
    let memberA = 0;
    let memberB = 0;
    const grants: TenantUserGrant[] = [];
    try {
      memberA = (
        await addTenantMember(api, { tenantId: tenantA.id, userId: user.id })
      ).id;
      memberB = (
        await addTenantMember(api, { tenantId: tenantB.id, userId: user.id })
      ).id;
      grants.push(
        await grantTenantPermissions(api, {
          roleKey: `${prefix}-user-query-b-${suffix}`,
          roleName: `${prefix.toUpperCase()} User Query B ${suffix}`,
          tenantId: tenantB.id,
          userId: user.id,
          permissions: ["system:user:query"],
        }),
      );
      const login = await loginRaw(user.username, password);
      const oldToken = await selectTenant(login.preToken!, tenantA.id);
      const tenantApi = await createTenantApiContext(oldToken);
      try {
        const switched = await expectSuccess<{ accessToken: string }>(
          await switchTenant(tenantApi, tenantB.id),
        );
        expect(switched.accessToken).toBeTruthy();
        expect(switched.accessToken).not.toBe(oldToken);
        expect((await tenantApi.get("user/info")).status()).toBe(401);
        const switchedApi = await createTenantApiContext(switched.accessToken);
        try {
          await expectSuccess(await switchedApi.get("user/info"));
          await expectUserListContains(switchedApi, tenantB.id, user.username);
        } finally {
          await switchedApi.dispose();
        }
      } finally {
        await tenantApi.dispose();
      }
    } finally {
      revokeTenantPermissionGrants(grants);
      await cleanupTenantUser(api, user.id, memberA);
      if (memberB > 0) {
        await removeTenantMember(api, memberB).catch(() => {});
      }
      await deleteTenant(api, tenantA.id).catch(() => {});
      await deleteTenant(api, tenantB.id).catch(() => {});
    }
  });
}

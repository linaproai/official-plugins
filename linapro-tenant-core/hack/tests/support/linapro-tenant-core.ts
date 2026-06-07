import { config } from "@host-tests/fixtures/config";
import { test as authTest, expect } from "@host-tests/fixtures/auth";
import {
  enablePlugin,
  expectSuccess,
  getPlugin,
  getMenuIdsByPerms,
  installPlugin,
  playwrightRequest,
  syncPlugins,
  type APIRequestContext,
  type APIResponse,
} from "@host-tests/support/api/job";
import {
  execPgSQL,
  pgEscapeLiteral,
  queryPgRows,
  queryPgScalar,
} from "@host-tests/support/postgres";

export type MultiTenantMode =
  | "linapro-tenant-core-disabled"
  | "linapro-tenant-core-enabled";

export type MultiTenantFixtures = {
  multiTenantMode: MultiTenantMode;
};

const apiBaseURL = config.apiBaseURL;
const tenantCorePluginID = "linapro-tenant-core";
const tenantCoreApiBaseURL = `${config.publicBaseURL.replace(/\/$/, "")}/x/${tenantCorePluginID}/api/v1/`;

export function tenantCoreApiPath(pathName: string) {
  return new URL(pathName.replace(/^\/+/, ""), tenantCoreApiBaseURL).toString();
}

export type LoginTenant = {
  id: number;
  code: string;
  name: string;
  status: string;
};

export type TenantCreateResult = {
  id: number;
};

export type TenantMember = {
  id: number;
  tenantId: number;
  userId: number;
  username: string;
  status: number;
};

export type TenantUserGrant = {
  roleId: number;
  tenantId: number;
};

export const test = authTest.extend<MultiTenantFixtures>({
  multiTenantMode: ["linapro-tenant-core-enabled", { option: true }],
});

export { expect };

export async function ensureMultiTenantPluginEnabled(api: APIRequestContext) {
  await syncPlugins(api);
  const plugin = await getPlugin(api, "linapro-tenant-core");
  if (plugin.installed !== 1) {
    await installPlugin(api, "linapro-tenant-core");
  }
  if (plugin.enabled !== 1) {
    await enablePlugin(api, "linapro-tenant-core");
  }
  return plugin;
}

export async function loginRaw(
  username = config.adminUser,
  password = config.adminPass,
) {
  const api = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const response = await api.post("auth/login", {
    data: { username, password, clientType: "web" },
  });
  expect(response.ok()).toBeTruthy();
  const payload = await response.json();
  expect(payload.code).toBe(0);
  await api.dispose();
  return payload.data as {
    accessToken?: string;
    preToken?: string;
    tenants?: LoginTenant[];
  };
}

export async function createTenant(
  api: APIRequestContext,
  payload: { code: string; name: string; remark?: string },
) {
  return expectSuccess<TenantCreateResult>(
    await api.post(tenantCoreApiPath("platform/tenants"), {
      data: {
        remark: "",
        ...payload,
      },
    }),
  );
}

export function updateUserPrimaryTenant(username: string, tenantId: number) {
  execPgSQL(
    `UPDATE sys_user SET tenant_id = ${tenantId} WHERE username = '${pgEscapeLiteral(username)}';`,
  );
}

export async function deleteTenant(api: APIRequestContext, id: number) {
  await api.delete(tenantCoreApiPath(`platform/tenants/${id}`)).catch(() => {});
}

export async function addTenantMember(
  _api: APIRequestContext,
  payload: { tenantId: number; userId: number },
) {
  const membershipColumns = new Set(
    queryPgRows(`
      SELECT column_name
      FROM information_schema.columns
      WHERE table_schema = 'public'
        AND table_name = 'plugin_linapro_tenant_core_user_membership';
    `),
  );
  const insertColumns = ['"user_id"', '"tenant_id"', '"status"'];
  const insertValues = [`${payload.userId}`, `${payload.tenantId}`, "1"];
  if (membershipColumns.has("created_by")) {
    insertColumns.push('"created_by"');
    insertValues.push("0");
  }
  if (membershipColumns.has("updated_by")) {
    insertColumns.push('"updated_by"');
    insertValues.push("0");
  }
  const updateClauses = ['"status" = 1'];
  if (membershipColumns.has("deleted_at")) {
    updateClauses.push('"deleted_at" = NULL');
  }
  if (membershipColumns.has("updated_at")) {
    updateClauses.push('"updated_at" = NOW()');
  }
  const membershipPredicate = `
        "user_id" = ${payload.userId}
        AND "tenant_id" = ${payload.tenantId}
  `;
  execPgSQL(`
    BEGIN;
    LOCK TABLE plugin_linapro_tenant_core_user_membership IN SHARE ROW EXCLUSIVE MODE;
    UPDATE plugin_linapro_tenant_core_user_membership
    SET ${updateClauses.join(",\n        ")}
    WHERE ${membershipPredicate};
    INSERT INTO plugin_linapro_tenant_core_user_membership (${insertColumns.join(", ")})
    SELECT ${insertValues.join(", ")}
    WHERE NOT EXISTS (
      SELECT 1
      FROM plugin_linapro_tenant_core_user_membership
      WHERE ${membershipPredicate}
    );
    COMMIT;
  `);
  return {
    id: Number(
      queryPgScalar(`
        SELECT id
        FROM plugin_linapro_tenant_core_user_membership
        WHERE user_id = ${payload.userId}
          AND tenant_id = ${payload.tenantId}
        LIMIT 1;
      `),
    ),
  };
}

export async function grantTenantPermissions(
  api: APIRequestContext,
  payload: {
    roleKey: string;
    roleName: string;
    tenantId: number;
    userId: number;
    permissions: string[];
  },
): Promise<TenantUserGrant> {
  await getMenuIdsByPerms(api, payload.permissions);
  const roleId = Number(
    queryPgScalar(`
      INSERT INTO sys_role (name, key, sort, data_scope, status, remark, tenant_id, created_at, updated_at)
      VALUES (
        '${pgEscapeLiteral(payload.roleName)}',
        '${pgEscapeLiteral(payload.roleKey)}',
        1,
        2,
        1,
        'Multi-tenant E2E tenant role',
        ${payload.tenantId},
        NOW(),
        NOW()
      )
      RETURNING id;
    `),
  );
  execPgSQL(
    `
      INSERT INTO sys_role_menu (role_id, menu_id, tenant_id)
      SELECT ${roleId}, id, ${payload.tenantId}
      FROM sys_menu
      WHERE perms IN (${payload.permissions
        .map((permission) => `'${pgEscapeLiteral(permission)}'`)
        .join(", ")})
      ON CONFLICT DO NOTHING;

      INSERT INTO sys_user_role (user_id, role_id, tenant_id)
      VALUES (${payload.userId}, ${roleId}, ${payload.tenantId})
      ON CONFLICT DO NOTHING;
    `,
  );
  return { roleId, tenantId: payload.tenantId };
}

export function revokeTenantPermissionGrants(grants: TenantUserGrant[]) {
  const roleIds = grants
    .map((grant) => grant.roleId)
    .filter((roleId) => roleId > 0);
  if (roleIds.length === 0) {
    return;
  }
  const idList = roleIds.join(", ");
  execPgSQL(`
    DELETE FROM sys_role_menu WHERE role_id IN (${idList});
    DELETE FROM sys_user_role WHERE role_id IN (${idList});
    DELETE FROM sys_role WHERE id IN (${idList});
  `);
}

export async function removeTenantMember(_api: APIRequestContext, id: number) {
  if (id <= 0) {
    return;
  }
  execPgSQL(
    `DELETE FROM plugin_linapro_tenant_core_user_membership WHERE id = ${id};`,
  );
}

export function listTenantMembers(_api: APIRequestContext, tenantId: number) {
  const list = queryPgRows(`
    SELECT m.id || '|' || m.tenant_id || '|' || m.user_id || '|' || u.username || '|' || m.status
    FROM plugin_linapro_tenant_core_user_membership m
    JOIN sys_user u ON u.id = m.user_id
    WHERE m.tenant_id = ${tenantId}
      AND m.deleted_at IS NULL
    ORDER BY m.id;
  `).map((row) => {
    const [id, rowTenantId, userId, username, status] = row.split("|");
    return {
      id: Number(id),
      tenantId: Number(rowTenantId),
      userId: Number(userId),
      username,
      status: Number(status),
    };
  });
  return {
    list,
    total: list.length,
  };
}

export async function selectTenant(preToken: string, tenantId: number) {
  const api = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const response = await api.post(tenantCoreApiPath("auth/select-tenant"), {
    data: { preToken, tenantId },
  });
  const data = await expectSuccess<{ accessToken: string }>(response);
  await api.dispose();
  return data.accessToken;
}

export async function createTenantApiContext(accessToken: string) {
  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

export async function switchTenant(
  api: APIRequestContext,
  tenantId: number,
): Promise<APIResponse> {
  return api.post(tenantCoreApiPath("auth/switch-tenant"), {
    data: { tenantId },
  });
}

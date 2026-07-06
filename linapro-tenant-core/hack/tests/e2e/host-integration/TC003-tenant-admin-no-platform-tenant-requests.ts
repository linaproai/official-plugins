import type { Page } from "@host-tests/support/playwright";

import { mkdirSync } from "node:fs";
import path from "node:path";

import { test, expect } from "@host-tests/fixtures/auth";
import { RolePage } from "@host-tests/pages/RolePage";
import { UserPage } from "@host-tests/pages/UserPage";

type TenantRequestCounters = {
  loginTenants: number;
  platformTenants: number;
};

const tenantAdminPermissions = [
  "system:user:query",
  "system:role:query",
  "system:dict:query",
];

const tenantState = {
  currentTenant: {
    code: "alpha",
    id: 101,
    name: "Alpha Tenant",
    status: "active",
  },
  enabled: true,
  impersonation: { active: false },
  tenants: [],
};

test.describe("TC-3 租户管理员宿主页不触发平台租户候选请求", () => {
  test.beforeEach(async ({ adminPage }) => {
    await installTenantAdminWorkbench(adminPage);
  });

  test("TC-3a: 租户管理员访问用户管理不请求平台租户控制面候选接口", async ({
    adminPage,
  }) => {
    // Regression: https://github.com/linaproai/linapro/issues/76
    const counters = await blockRestrictedTenantCandidateApis(adminPage);

    await reloadAsTenantAdmin(adminPage);
    const userPage = new UserPage(adminPage);
    await userPage.goto();
    await captureTenantAdminScreenshot(adminPage, "user-management");

    expect(counters.platformTenants).toBe(0);
    expect(counters.loginTenants).toBe(0);
  });

  test("TC-3b: 租户管理员访问角色管理不请求平台租户控制面候选接口", async ({
    adminPage,
  }) => {
    // Regression: https://github.com/linaproai/linapro/issues/76
    const counters = await blockRestrictedTenantCandidateApis(adminPage);

    await reloadAsTenantAdmin(adminPage);
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();
    await captureTenantAdminScreenshot(adminPage, "role-management");

    expect(counters.platformTenants).toBe(0);
    expect(counters.loginTenants).toBe(0);
  });
});

async function installTenantAdminWorkbench(page: Page) {
  await mockUserInfo(page);
  await mockPluginRuntimeStates(page);
  await mockDictionaries(page);
  await mockUserManagement(page);
  await mockRoleManagement(page);
  await page.evaluate(
    ({ permissions, tenant }) => {
      const accessStorageKey = "lina-web-antd-5.6.0-dev-core-access";
      const rawAccessState = localStorage.getItem(accessStorageKey);
      const accessState = rawAccessState ? JSON.parse(rawAccessState) : {};
      localStorage.setItem(
        accessStorageKey,
        JSON.stringify({
          ...accessState,
          accessCodes: permissions,
          isLockScreen: false,
        }),
      );
      localStorage.setItem("linapro:tenant-state", JSON.stringify(tenant));
    },
    { permissions: tenantAdminPermissions, tenant: tenantState },
  );
}

async function reloadAsTenantAdmin(page: Page) {
  await page.reload({ waitUntil: "domcontentloaded" });
  await page.waitForLoadState("networkidle").catch(() => {});
}

async function blockRestrictedTenantCandidateApis(page: Page) {
  const counters: TenantRequestCounters = {
    loginTenants: 0,
    platformTenants: 0,
  };

  await page.route(
    /\/x\/linapro-tenant-core\/api\/v1\/auth\/login-tenants(?:\?.*)?$/,
    async (route) => {
      counters.loginTenants += 1;
      await route.fulfill(forbidden("system:tenant:auth:login-tenants"));
    },
  );
  await page.route(
    /\/x\/linapro-tenant-core\/api\/v1\/platform\/tenants(?:\?.*)?$/,
    async (route) => {
      counters.platformTenants += 1;
      await route.fulfill(forbidden("system:tenant:list"));
    },
  );

  return counters;
}

async function mockUserInfo(page: Page) {
  await page.unroute("**/api/v1/user/info").catch(() => {});
  await page.route("**/api/v1/user/info", async (route) => {
    await route.fulfill(ok({
      avatar: "",
      desc: "Tenant administrator",
      homePath: "/system/user",
      menus: [
        { children: [], id: 1001, name: "SystemUser", path: "/system/user" },
        { children: [], id: 1002, name: "SystemRole", path: "/system/role" },
        {
          children: [],
          id: 1003,
          name: "PlatformTenantManagement",
          path: "/platform/tenants",
        },
      ],
      permissions: tenantAdminPermissions,
      realName: "Tenant Admin",
      roles: ["tenant-admin"],
      userId: 201,
      username: "tenant-admin",
    }));
  });
}

async function mockPluginRuntimeStates(page: Page) {
  await page.unroute("**/api/v1/plugins/dynamic**").catch(() => {});
  await page.route("**/api/v1/plugins/dynamic**", async (route) => {
    await route.fulfill(ok({
      list: [
        {
          enabled: 1,
          generation: 1,
          id: "linapro-tenant-core",
          installed: 1,
          runtimeState: "normal",
          statusKey: "sys_plugin.status:linapro-tenant-core",
          version: "e2e",
        },
      ],
    }));
  });
  await page.evaluate(() => {
    const registryGlobal = globalThis as any;
    registryGlobal.__linaPluginStatePromise = null;
    registryGlobal.__linaPluginStateSignature = null;
  });
}

async function mockDictionaries(page: Page) {
  await page.route("**/api/v1/dict/data/type/sys_normal_disable", async (route) => {
    await route.fulfill(ok({
      list: [
        { label: "正常", value: "1" },
        { label: "停用", value: "0" },
      ],
    }));
  });
  await page.route("**/api/v1/dict/data/type/sys_data_scope", async (route) => {
    await route.fulfill(ok({
      list: [
        { label: "全部数据", value: "1" },
        { label: "本租户数据", value: "2" },
        { label: "本人数据", value: "4" },
      ],
    }));
  });
}

async function mockUserManagement(page: Page) {
  await page.route(/\/api\/v1\/user(?:\?.*)?$/, async (route) => {
    if (route.request().method() !== "GET") {
      await route.fallback();
      return;
    }
    await route.fulfill(ok({
      list: [
        {
          createdAt: "2026-07-06 10:00:00",
          email: "tenant-admin@example.test",
          id: 201,
          nickname: "Tenant Admin",
          phone: "",
          roleNames: "Tenant Admin",
          sex: 0,
          status: 1,
          tenantId: 101,
          tenantNames: ["Alpha Tenant"],
          username: "tenant-admin",
        },
      ],
      total: 1,
    }));
  });
  await page.route("**/api/v1/role/options**", async (route) => {
    await route.fulfill(ok({
      list: [{ id: 201, key: "tenant-admin", name: "Tenant Admin" }],
    }));
  });
}

async function mockRoleManagement(page: Page) {
  await page.route(/\/api\/v1\/role(?:\?.*)?$/, async (route) => {
    if (route.request().method() !== "GET") {
      await route.fallback();
      return;
    }
    await route.fulfill(ok({
      list: [
        {
          createdAt: "2026-07-06 10:00:00",
          dataScope: 2,
          id: 201,
          key: "tenant-admin",
          name: "Tenant Admin",
          sort: 1,
          status: 1,
        },
      ],
      total: 1,
    }));
  });
}

function ok(data: unknown) {
  return {
    body: JSON.stringify({
      code: 0,
      data,
      message: "success",
    }),
    contentType: "application/json",
    status: 200,
  };
}

function forbidden(permission: string) {
  return {
    body: JSON.stringify({
      code: 61,
      data: null,
      errorCode: "PERMISSION_DENIED_REQUIRED",
      message: `当前用户缺少接口权限: ${permission}`,
      messageKey: "error.permission.denied.required",
      messageParams: { permissions: permission },
    }),
    contentType: "application/json",
    status: 403,
  };
}

async function captureTenantAdminScreenshot(page: Page, description: string) {
  const now = new Date();
  const date = now.toISOString().slice(0, 10).replaceAll("-", "");
  const time = now.toISOString().slice(11, 19).replaceAll(":", "");
  const cwd = process.cwd().replaceAll("\\", "/");
  const repoRoot = cwd.endsWith("/hack/tests")
    ? path.resolve(process.cwd(), "../..")
    : process.cwd();
  const screenshotDir = path.join(repoRoot, "temp", date);
  mkdirSync(screenshotDir, { recursive: true });
  await page.screenshot({
    fullPage: false,
    path: path.join(screenshotDir, `${time}-tc003-${description}.png`),
  });
}

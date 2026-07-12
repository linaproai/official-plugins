/**
 * TC001 linapro-extlogin-core 授权登录目录
 *
 * 验证领域目录由本插件安装创建，而非宿主预留：
 * - 安装并启用后出现「授权登录」
 * - 卸载后目录消失
 * - 一级排序位于扩展中心之前（若环境存在智能中心则位于其后）
 */
import type { APIRequestContext } from "@playwright/test";

import { expect, test } from "@host-tests/fixtures/auth";
import {
  createAdminApiContext,
  findPlugin,
  installPlugin,
  prepareSourcePluginsBaseline,
  syncPlugins,
  uninstallPlugin,
  updatePluginStatus,
} from "@host-tests/fixtures/plugin";
import { MainLayout } from "@host-tests/pages/MainLayout";
import { workspacePath } from "@host-tests/fixtures/config";
import { waitForRouteReady } from "@host-tests/support/ui";

const pluginID = "linapro-extlogin-core";
// Install one settings child so the Auth Login directory is non-empty.
// Frontend route menus intentionally hide empty directory shells.
const catalogChildPluginID = "linapro-auth-ldap";
const dependentPluginIDs = [
  catalogChildPluginID,
  "linapro-oidc-google",
  "linapro-oidc-discord",
] as const;

type MenuNode = {
  children?: MenuNode[];
  name?: string;
  path?: string;
};

function findByName(list: MenuNode[], name: string | RegExp): MenuNode | null {
  for (const item of list) {
    const title = String(item.name ?? "");
    if (
      (typeof name === "string" && title === name) ||
      (name instanceof RegExp && name.test(title))
    ) {
      return item;
    }
    const nested = findByName(item.children ?? [], name);
    if (nested) {
      return nested;
    }
  }
  return null;
}

async function fetchMenuTree(api: APIRequestContext): Promise<MenuNode[]> {
  const response = await api.get("menu");
  expect(response.ok()).toBeTruthy();
  const body = await response.json();
  const data = body?.data ?? body;
  if (Array.isArray(data?.list)) {
    return data.list as MenuNode[];
  }
  if (Array.isArray(data)) {
    return data as MenuNode[];
  }
  return [];
}

async function uninstallDependentsIfPresent(api: APIRequestContext) {
  for (const id of dependentPluginIDs) {
    const plugin = await findPlugin(api, id);
    if (plugin?.installed === 1) {
      await uninstallPlugin(api, id, false);
    }
  }
}

/** Install catalog owner first, then a child settings page (dependency order). */
async function ensureAuthLoginCatalogVisible() {
  // prepareSourcePluginsBaseline sorts IDs alphabetically; keep explicit order
  // so the parent domain catalog exists before child settings menus attach.
  await prepareSourcePluginsBaseline([pluginID]);
  await prepareSourcePluginsBaseline([catalogChildPluginID]);
}

test.describe("TC001 linapro-extlogin-core 授权登录目录", () => {
  test.beforeAll(async () => {
    // Domain catalog plus one child settings page keeps the directory visible
    // in sidebar route menus (empty directories are filtered out).
    await ensureAuthLoginCatalogVisible();
  });

  test("TC001a: 安装启用后出现授权登录目录且位于扩展中心之前", async () => {
    const api = await createAdminApiContext();
    try {
      await syncPlugins(api);
      await ensureAuthLoginCatalogVisible();

      const tree = await fetchMenuTree(api);
      const authLogin = findByName(tree, /授权登录|Auth Login/i);
      expect(
        authLogin,
        "linapro-extlogin-core 安装后应声明授权登录目录",
      ).not.toBeNull();
      expect(authLogin?.path).toBe("auth-login");

      // 宿主不得再以 menu_key=auth-login 预留空目录；本目录 path 来自插件声明。
      const names = tree.map((node) => node.name);
      const authIdx = names.findIndex((name) =>
        /授权登录|Auth Login/i.test(String(name ?? "")),
      );
      const extensionIdx = names.findIndex((name) =>
        /扩展中心|Extensions/i.test(String(name ?? "")),
      );
      expect(authIdx).toBeGreaterThanOrEqual(0);
      expect(extensionIdx).toBeGreaterThanOrEqual(0);
      expect(authIdx).toBeLessThan(extensionIdx);

      const smartIdx = names.findIndex((name) =>
        /智能中心|AI Hub/i.test(String(name ?? "")),
      );
      if (smartIdx >= 0) {
        expect(authIdx).toBeGreaterThan(smartIdx);
      }
    } finally {
      await api.dispose();
    }
  });

  test("TC001b: 侧边栏可见授权登录目录", async ({ adminPage }) => {
    await ensureAuthLoginCatalogVisible();
    const layout = new MainLayout(adminPage);
    await adminPage.goto(workspacePath("/dashboard/workspace"));
    await waitForRouteReady(adminPage);
    await expect(
      layout.sidebarMenuItem(/授权登录|Auth Login/i),
    ).toBeVisible();
  });

  test("TC001c: 卸载本插件后授权登录目录消失", async () => {
    const api = await createAdminApiContext();
    try {
      await syncPlugins(api);
      // Reverse dependents must be removed before uninstalling the owner catalog.
      await uninstallDependentsIfPresent(api);
      const self = await findPlugin(api, pluginID);
      if (self?.installed === 1) {
        await uninstallPlugin(api, pluginID, false);
      }

      const tree = await fetchMenuTree(api);
      const authLogin = findByName(tree, /授权登录|Auth Login/i);
      expect(
        authLogin,
        "卸载 linapro-extlogin-core 后不应再保留授权登录目录",
      ).toBeNull();

      // Restore for subsequent suite consumers.
      await installPlugin(api, pluginID, "global");
      await updatePluginStatus(api, pluginID, true);
    } finally {
      await api.dispose();
    }
  });
});

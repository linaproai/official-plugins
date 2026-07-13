/**
 * TC001 宿主「系统设置」目录与云存储插件挂载
 *
 * - 导航中不出现独立「存储管理」一级目录
 * - 安装 linapro-storage-s3 后在「系统设置」下出现 存储管理-S3 子菜单
 */
import type { APIRequestContext } from "@playwright/test";

import { expect, test } from "@host-tests/fixtures/auth";
import {
  createAdminApiContext,
  findPlugin,
  prepareSourcePluginsBaseline,
  syncPlugins,
  uninstallPlugin,
} from "@host-tests/fixtures/plugin";
import { MainLayout } from "@host-tests/pages/MainLayout";
import { workspacePath } from "@host-tests/fixtures/config";
import { waitForRouteReady } from "@host-tests/support/ui";

const pluginID = "linapro-storage-s3";
const storagePlugins = [
  "linapro-storage-aws",
  "linapro-storage-azure",
  "linapro-storage-cos",
  "linapro-storage-obs",
  "linapro-storage-oss",
  "linapro-storage-qiniu",
  "linapro-storage-s3",
] as const;

type MenuNode = {
  children?: MenuNode[];
  meta?: { title?: string };
  name?: string;
  path?: string;
};

/** Match admin list `name` or navigation route projection `meta.title`. */
function nodeTitle(item: MenuNode): string {
  return String(item.meta?.title ?? item.name ?? "");
}

function findByName(list: MenuNode[], name: string | RegExp): MenuNode | null {
  for (const item of list) {
    const title = nodeTitle(item);
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

/**
 * Navigation route projection (`/menus/all`) is what the sidebar uses.
 */
async function fetchNavMenuTree(api: APIRequestContext): Promise<MenuNode[]> {
  const response = await api.get("menus/all");
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

/** Full menu management tree. */
async function fetchAdminMenuTree(api: APIRequestContext): Promise<MenuNode[]> {
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

async function uninstallStoragePlugins(api: APIRequestContext) {
  for (const id of storagePlugins) {
    const plugin = await findPlugin(api, id);
    if (plugin?.installed === 1) {
      await uninstallPlugin(api, id, false);
    }
  }
}

test.describe("TC001 linapro-storage-s3 系统设置挂载", () => {
  test("TC001a: 导航不展示独立存储管理目录", async () => {
    const api = await createAdminApiContext();
    try {
      await syncPlugins(api);
      await uninstallStoragePlugins(api);
      const tree = await fetchNavMenuTree(api);
      // Only the removed top-level catalog; do not match "存储管理-S3".
      const storage = tree.find((item) =>
        /^(存储管理|Storage Management|Storage)$/i.test(nodeTitle(item)),
      );
      expect(
        storage ?? null,
        "不得展示独立的存储管理一级目录",
      ).toBeNull();
    } finally {
      await api.dispose();
    }
  });

  test("TC001b: 安装 S3 插件后挂载到系统设置", async () => {
    const api = await createAdminApiContext();
    try {
      await syncPlugins(api);
      await prepareSourcePluginsBaseline([pluginID]);
      const adminTree = await fetchAdminMenuTree(api);
      expect(
        adminTree.find((item) =>
          /^(存储管理|Storage Management|Storage)$/i.test(nodeTitle(item)),
        ) ?? null,
        "管理菜单树中不应再有存储管理一级目录",
      ).toBeNull();

      const setting = findByName(adminTree, /系统设置|Settings/i);
      expect(setting, "应存在系统设置目录").not.toBeNull();
      expect(setting?.path).toBe("setting");
      const child = findByName(
        setting?.children ?? [],
        /存储管理-S3|Storage Management - S3/i,
      );
      expect(child, "系统设置下应有 存储管理-S3 配置菜单").not.toBeNull();

      const navTree = await fetchNavMenuTree(api);
      const navSetting = findByName(navTree, /系统设置|Settings/i);
      expect(navSetting, "导航中应展示系统设置").not.toBeNull();
      const navChild = findByName(
        navSetting?.children ?? [],
        /存储管理-S3|Storage Management - S3/i,
      );
      expect(navChild, "导航系统设置下应有 存储管理-S3").not.toBeNull();
    } finally {
      await api.dispose();
    }
  });

  test("TC001c: 侧边栏在系统设置下可见 存储管理-S3", async ({ adminPage }) => {
    await prepareSourcePluginsBaseline([pluginID]);
    const layout = new MainLayout(adminPage);
    await adminPage.goto(workspacePath("/dashboard/workspace"));
    await waitForRouteReady(adminPage);
    await expect(
      layout.sidebar.getByText(/^(存储管理|Storage Management|Storage)$/i),
    ).toHaveCount(0);
    await layout.expandSidebarGroup(/系统设置|Settings/i);
    await expect(
      layout.sidebarMenuItem(/存储管理-S3|Storage Management - S3/i),
    ).toBeVisible();
  });
});

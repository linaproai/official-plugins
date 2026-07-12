/**
 * TC001 宿主「存储管理」目录与云存储插件挂载
 *
 * - 未安装云存储配置插件时导航隐藏「存储管理」
 * - 安装 linapro-storage-s3 后出现目录与 S3 存储 子菜单
 * - 目录排序：存储管理 → 扩展中心 → 开发中心
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

const pluginID = "linapro-storage-s3";
const storagePlugins = [
  "linapro-storage-cos",
  "linapro-storage-oss",
  "linapro-storage-aws",
  "linapro-storage-s3",
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

async function uninstallStoragePlugins(api: APIRequestContext) {
  for (const id of storagePlugins) {
    const plugin = await findPlugin(api, id);
    if (plugin?.installed === 1) {
      await uninstallPlugin(api, id, false);
    }
  }
}

test.describe("TC001 linapro-storage-s3 存储管理目录", () => {
  test("TC001a: 未安装云存储插件时隐藏存储管理", async () => {
    const api = await createAdminApiContext();
    try {
      await syncPlugins(api);
      await uninstallStoragePlugins(api);
      const tree = await fetchMenuTree(api);
      const storage = findByName(tree, /存储管理|Storage/i);
      expect(
        storage,
        "无云存储配置子菜单时不应展示存储管理目录",
      ).toBeNull();
    } finally {
      await api.dispose();
    }
  });

  test("TC001b: 安装 S3 插件后出现存储管理与子菜单", async () => {
    const api = await createAdminApiContext();
    try {
      await syncPlugins(api);
      await prepareSourcePluginsBaseline([pluginID]);
      const tree = await fetchMenuTree(api);
      const storage = findByName(tree, /存储管理|Storage/i);
      expect(storage, "安装云存储插件后应展示存储管理").not.toBeNull();
      expect(storage?.path).toBe("storage");
      const child = findByName(
        storage?.children ?? [],
        /S3 存储|S3存储|S3 Storage|^S3$/i,
      );
      expect(child, "存储管理下应有 S3 存储 配置菜单").not.toBeNull();

      const names = tree.map((node) => node.name);
      const storageIdx = names.findIndex((name) =>
        /存储管理|Storage/i.test(String(name ?? "")),
      );
      const extensionIdx = names.findIndex((name) =>
        /扩展中心|Extensions/i.test(String(name ?? "")),
      );
      const developerIdx = names.findIndex((name) =>
        /开发中心|Dev Tools/i.test(String(name ?? "")),
      );
      expect(storageIdx).toBeGreaterThanOrEqual(0);
      if (extensionIdx >= 0) {
        expect(storageIdx).toBeLessThan(extensionIdx);
      }
      if (developerIdx >= 0) {
        expect(storageIdx).toBeLessThan(developerIdx);
        if (extensionIdx >= 0) {
          expect(extensionIdx).toBeLessThan(developerIdx);
        }
      }
    } finally {
      await api.dispose();
    }
  });

  test("TC001c: 侧边栏可见存储管理", async ({ adminPage }) => {
    await prepareSourcePluginsBaseline([pluginID]);
    const layout = new MainLayout(adminPage);
    await adminPage.goto(workspacePath("/dashboard/workspace"));
    await waitForRouteReady(adminPage);
    await expect(
      layout.sidebarMenuItem(/存储管理|Storage/i),
    ).toBeVisible();
  });
});

import type { APIRequestContext } from "@host-tests/support/playwright";

import { execFileSync } from "node:child_process";
import { rmSync } from "node:fs";
import path from "node:path";

import { test, expect } from '@host-tests/fixtures/auth';
import { config } from '@host-tests/fixtures/config';
import { DemoDynamicPage } from '../../pages/DemoDynamicPage';
import {
  createAdminApiContext,
  disablePlugin,
  enablePlugin,
  getPlugin,
  installPlugin,
  listPlugins,
  syncPlugins,
  uninstallPlugin,
} from '@host-tests/support/api/job';
import {
  execPgSQLStatements,
  pgEscapeLiteral,
  pgIdentifier,
} from '@host-tests/support/postgres';
import { waitForRouteReady } from '@host-tests/support/ui';

const pluginID = "linapro-demo-dynamic";
const sourcePluginID = "linapro-demo-source";
const pluginMenuNamePattern = /Dynamic Plugin Demo|动态插件示例/u;
const recordTable = "plugin_linapro_demo_dynamic_record";
const publicBaseURL = config.publicBaseURL;
const repoRoot = path.resolve(process.cwd(), "../..");
const legacyRuntimeArtifactPath = path.join(
  repoRoot,
  "apps",
  "lina-plugins",
  pluginID,
  "runtime",
  `${pluginID}.wasm`,
);

let adminApi: APIRequestContext;
let originalInstalled = 0;
let originalEnabled = 0;
let originalSourceInstalled = 0;
let originalSourceEnabled = 0;

type DemoRecordListPayload = {
  list?: Array<{ title?: string }>;
  total?: number;
};

function unwrapApiData(payload: any) {
  if (payload && typeof payload === "object" && "data" in payload) {
    return payload.data;
  }
  return payload;
}

function ensureRuntimePluginArtifact() {
  execFileSync("make", ["wasm", `p=${pluginID}`, "out=../../temp/output"], {
    cwd: repoRoot,
    stdio: "inherit",
  });
  rmSync(legacyRuntimeArtifactPath, { force: true });
}

function cleanupRuntimePluginData() {
  execPgSQLStatements([
    `DROP TABLE IF EXISTS ${pgIdentifier(recordTable)};`,
    `DELETE FROM sys_plugin_migration WHERE plugin_id = '${pgEscapeLiteral(pluginID)}';`,
  ]);
}

function cleanupRuntimePluginRows() {
  const escapedPluginID = pgEscapeLiteral(pluginID);
  execPgSQLStatements([
    `DELETE FROM sys_role_menu WHERE menu_id IN (SELECT id FROM sys_menu WHERE menu_key LIKE 'plugin:${escapedPluginID}:%');`,
    `DELETE FROM sys_menu WHERE menu_key LIKE 'plugin:${escapedPluginID}:%';`,
    `DELETE FROM sys_plugin_state WHERE plugin_id = '${escapedPluginID}';`,
    `DELETE FROM sys_plugin_node_state WHERE plugin_id = '${escapedPluginID}';`,
    `DELETE FROM sys_plugin_resource_ref WHERE plugin_id = '${escapedPluginID}';`,
    `DELETE FROM sys_plugin_migration WHERE plugin_id = '${escapedPluginID}';`,
    `DELETE FROM sys_plugin_release WHERE plugin_id = '${escapedPluginID}';`,
    `DELETE FROM sys_plugin WHERE plugin_id = '${escapedPluginID}';`,
  ]);
}

async function ensureSourcePluginInstalledAndEnabled() {
  let sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  if (sourcePlugin.installed !== 1) {
    await installPlugin(adminApi, sourcePluginID, { installMode: "global" });
    sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  }
  if (sourcePlugin.enabled !== 1) {
    await enablePlugin(adminApi, sourcePluginID);
  }
}

async function ensurePluginInstalledAndEnabled() {
  await ensureSourcePluginInstalledAndEnabled();
  const plugin = await getPlugin(adminApi, pluginID);
  if (plugin.installed !== 1) {
    await installPlugin(adminApi, pluginID, { installMode: "global" });
  }

  const refreshedPlugin = await getPlugin(adminApi, pluginID);
  if (refreshedPlugin.enabled !== 1) {
    await enablePlugin(adminApi, pluginID);
  }
}

async function demoRecordListSnapshot(pageSize = 20) {
  try {
    const response = await adminApi.get(
      `${publicBaseURL}/x/${pluginID}/api/v1/demo-records`,
      {
        params: {
          pageNum: 1,
          pageSize,
        },
      },
    );
    if (!response.ok()) {
      return {
        ok: false,
        status: response.status(),
        titles: [] as string[],
      };
    }

    const payload = unwrapApiData(
      (await response.json()) as DemoRecordListPayload,
    ) as DemoRecordListPayload;
    const records = Array.isArray(payload?.list) ? payload.list : [];
    return {
      ok: true,
      status: response.status(),
      titles: records
        .map((item) => item.title ?? "")
        .filter((title) => title.length > 0),
    };
  } catch (error) {
    return {
      error: error instanceof Error ? error.message : String(error),
      ok: false,
      status: 0,
      titles: [] as string[],
    };
  }
}

async function waitForDemoRecord(title: string) {
  await expect
    .poll(
      async () => {
        const snapshot = await demoRecordListSnapshot();
        return snapshot.ok && snapshot.titles.includes(title);
      },
      {
        message: `等待 ${pluginID} 动态路由返回记录: ${title}`,
        timeout: 20_000,
      },
    )
    .toBe(true);
}

async function restorePluginState() {
  let plugin = await getPlugin(adminApi, pluginID);

  if (originalInstalled !== 1) {
    await removePluginForCleanState();
    return;
  }

  if (originalEnabled !== 1 && plugin.enabled === 1) {
    try {
      await disablePlugin(adminApi, pluginID);
    } catch (error) {
      cleanupRuntimePluginRows();
      await syncPlugins(adminApi);
      await installPlugin(adminApi, pluginID, { installMode: "global" });
    }
  }
}

async function removePluginForCleanState() {
  let plugin = await getPlugin(adminApi, pluginID);
  try {
    if (plugin.enabled === 1) {
      await disablePlugin(adminApi, pluginID);
      plugin = await getPlugin(adminApi, pluginID);
    }
    if (plugin.installed === 1) {
      await uninstallPlugin(adminApi, pluginID);
    }
  } catch (error) {
    cleanupRuntimePluginRows();
    await syncPlugins(adminApi);
  }
}

async function restoreSourcePluginState() {
  let sourcePlugin = await getPlugin(adminApi, sourcePluginID);

  if (originalSourceInstalled !== 1) {
    if (sourcePlugin.enabled === 1) {
      await disablePlugin(adminApi, sourcePluginID);
      sourcePlugin = await getPlugin(adminApi, sourcePluginID);
    }
    if (sourcePlugin.installed === 1) {
      await uninstallPlugin(adminApi, sourcePluginID);
    }
    return;
  }

  if (sourcePlugin.installed !== 1) {
    await installPlugin(adminApi, sourcePluginID, { installMode: "global" });
    sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  }
  if (originalSourceEnabled === 1 && sourcePlugin.enabled !== 1) {
    await enablePlugin(adminApi, sourcePluginID);
  } else if (originalSourceEnabled !== 1 && sourcePlugin.enabled === 1) {
    await disablePlugin(adminApi, sourcePluginID);
  }
}

test.describe("TC003 英文运行时页面巡检", () => {
  test.beforeAll(async () => {
    ensureRuntimePluginArtifact();
    adminApi = await createAdminApiContext();
    await syncPlugins(adminApi);
    const sourcePlugin = await getPlugin(adminApi, sourcePluginID);
    const plugin = await getPlugin(adminApi, pluginID);
    originalSourceInstalled = sourcePlugin.installed;
    originalSourceEnabled = sourcePlugin.enabled;
    originalInstalled = plugin.installed;
    originalEnabled = plugin.enabled;
  });

  test.beforeEach(async () => {
    cleanupRuntimePluginData();
    await removePluginForCleanState();
  });

  test.afterAll(async () => {
    try {
      await restorePluginState();
      await restoreSourcePluginState();
    } finally {
      if (originalInstalled !== 1) {
        cleanupRuntimePluginData();
      }
      await adminApi.dispose();
    }
  });

  test("TC-3a: 英文环境下动态插件管理列表元数据保持英文", async ({
    adminPage,
    mainLayout,
  }) => {
    test.setTimeout(120_000);
    await ensurePluginInstalledAndEnabled();

    await mainLayout.switchLanguage("English");

    const pluginPage = new DemoDynamicPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    await expect(pluginPage.pluginRow(pluginID)).toContainText(
      "Dynamic Plugin Demo",
    );
    await expect(pluginPage.pluginDescriptionCell(pluginID)).toContainText(
      "Dynamic wasm sample that demonstrates a host-embedded menu page, plugin-owned SQL CRUD, and a hosted standalone page.",
    );
    const rowText = await pluginPage.pluginRow(pluginID).innerText();
    expect(rowText).not.toContain("动态插件示例");
    expect(rowText).not.toContain("提供独立的 dynamic wasm 插件样例");
  });

  test("TC-3b: 英文环境下未安装动态插件列表元数据保持英文", async ({
    adminPage,
    mainLayout,
  }) => {
    await removePluginForCleanState();

    await mainLayout.switchLanguage("English");

    const apiList = await listPlugins(adminApi, pluginID, "en-US");
    const apiPlugin = apiList.list.find((item) => item.id === pluginID);
    expect(apiPlugin).toBeTruthy();
    expect(apiPlugin?.installed).toBe(0);
    expect(apiPlugin?.name).toBe("Dynamic Plugin Demo");
    expect(apiPlugin?.description).toBe(
      "Dynamic wasm sample that demonstrates a host-embedded menu page, plugin-owned SQL CRUD, and a hosted standalone page.",
    );

    const pluginPage = new DemoDynamicPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    await expect(pluginPage.pluginRow(pluginID)).toContainText(
      "Dynamic Plugin Demo",
    );
    await expect(pluginPage.pluginDescriptionCell(pluginID)).toContainText(
      "Dynamic wasm sample that demonstrates a host-embedded menu page, plugin-owned SQL CRUD, and a hosted standalone page.",
    );
    const rowText = await pluginPage.pluginRow(pluginID).innerText();
    expect(rowText).not.toContain("动态插件示例");
    expect(rowText).not.toContain("提供独立的 dynamic wasm 插件样例");
  });

  test("TC-3c: 英文环境下动态插件页面与独立页种子内容保持英文", async ({
    adminPage,
    mainLayout,
  }) => {
    await ensurePluginInstalledAndEnabled();
    await waitForDemoRecord("Dynamic Plugin SQL Demo Record");
    await adminPage.reload({ waitUntil: "domcontentloaded" });
    await waitForRouteReady(adminPage);

    const pluginPage = new DemoDynamicPage(adminPage);

    await mainLayout.switchLanguage("English");
    await pluginPage.clickSidebarMenuItem(pluginMenuNamePattern);
    await waitForRouteReady(adminPage);

    await expect(pluginPage.pluginDemoDynamicTitle()).toHaveText(
      "Dynamic Plugin Demo Is Live",
    );
    await expect(
      pluginPage.pluginDemoDynamicRecordRow("Dynamic Plugin SQL Demo Record"),
    ).toBeVisible();
    await expect(
      adminPage.getByText(
        "This record is seeded by the linapro-demo-dynamic install SQL",
        { exact: false },
      ),
    ).toBeVisible();

    const [standalonePage] = await Promise.all([
      adminPage.context().waitForEvent("page"),
      pluginPage.pluginDemoDynamicOpenStandaloneButton().click(),
    ]);
    await standalonePage.waitForLoadState("domcontentloaded");
    await standalonePage.waitForLoadState("networkidle").catch(() => {});

    await expect.poll(async () => standalonePage.url()).toContain("lang=en-US");
    await expect(
      standalonePage.getByText("Dynamic Plugin Standalone Page", {
        exact: true,
      }),
    ).toBeVisible();
    await expect(
      standalonePage.getByText(
        "This page is served directly by linapro-demo-dynamic as a hosted static asset",
        { exact: false },
      ),
    ).toBeVisible();
    await standalonePage.close();
  });
});

import type { APIRequestContext } from "@host-tests/support/playwright";

import { test, expect } from '@host-tests/fixtures/auth';
import { pluginApiPath } from '@host-tests/fixtures/config';
import {
  createAdminApiContext,
  enablePlugin,
  expectSuccess,
  getPlugin,
  installPlugin,
  syncPlugins,
} from '@host-tests/support/api/job';
import { waitForRouteReady } from '@host-tests/support/ui';

type DeptTreeNode = {
  children?: DeptTreeNode[];
  id: number;
  label: string;
};

const sourcePluginIDs = ["linapro-org-core", "linapro-demo-source"] as const;
const demoSourcePluginID = "linapro-demo-source";
const chineseSystemCopyPattern =
  /未分配部门|服务运行时长|小时|分钟|秒|这是一条来自 linapro-demo-source 接口的简要介绍/;

async function ensureSourcePluginsEnabled(
  api: APIRequestContext,
  pluginIDs: readonly string[],
) {
  await syncPlugins(api);
  for (const pluginID of pluginIDs) {
    let plugin = await getPlugin(api, pluginID);
    if (plugin.installed !== 1) {
      await installPlugin(api, pluginID);
      plugin = await getPlugin(api, pluginID);
    }
    if (plugin.enabled !== 1) {
      await enablePlugin(api, pluginID);
    }
  }
}

function flattenDeptTree(nodes: DeptTreeNode[]): DeptTreeNode[] {
  return nodes.flatMap((node) => [node, ...flattenDeptTree(node.children ?? [])]);
}

test.describe("TC-3 Backend hardcoded Chinese regression", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    await ensureSourcePluginsEnabled(adminApi, sourcePluginIDs);
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test("TC-3a: English backend projections do not return legacy Chinese system copy", async () => {
    const deptTree = await expectSuccess<{ list: DeptTreeNode[] }>(
      await adminApi.get("user/dept-tree", {
        headers: { "Accept-Language": "en-US" },
      }),
    );
    const unassignedNode = flattenDeptTree(deptTree.list).find((node) => node.id === 0);
    expect(unassignedNode?.label).toContain("Unassigned");
    expect(unassignedNode?.label).not.toContain("未分配部门");

    const systemInfo = await expectSuccess<{
      runDuration: string;
      runDurationSeconds: number;
    }>(
      await adminApi.get("system/info", {
        headers: { "Accept-Language": "en-US" },
      }),
    );
    expect(systemInfo.runDurationSeconds).toBeGreaterThanOrEqual(0);
    expect(systemInfo.runDuration).toMatch(/second|minute|hour|day/);
    expect(systemInfo.runDuration).not.toMatch(/小时|分钟|秒/);

    const summary = await expectSuccess<{ message: string }>(
      await adminApi.get(
        pluginApiPath(demoSourcePluginID, "plugins/linapro-demo-source/summary"),
        {
          headers: { "Accept-Language": "en-US" },
        },
      ),
    );
    expect(summary.message).toContain("linapro-demo-source API");
    expect(summary.message).not.toMatch(chineseSystemCopyPattern);
  });

  test("TC-3b: English source-plugin page does not show legacy backend Chinese copy", async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage("English");
    await adminPage.goto("/linapro-demo-source-sidebar-entry?lang=en-US", {
      waitUntil: "domcontentloaded",
    });
    await waitForRouteReady(adminPage, 15_000);

    const bodyText = await adminPage.locator("body").innerText();
    expect(bodyText).toContain("Source Plugin Demo");
    expect(bodyText).toContain("Demo Records");
    expect(bodyText).not.toMatch(chineseSystemCopyPattern);
  });
});

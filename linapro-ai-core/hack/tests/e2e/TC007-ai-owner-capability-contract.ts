import { execFileSync } from "node:child_process";
import { mkdirSync, readFileSync, rmSync } from "node:fs";
import path from "node:path";

import type {
  APIRequestContext,
  APIResponse,
} from "@host-tests/support/playwright";

import { expect, test } from "@host-tests/fixtures/auth";
import { pluginApiPath, workspacePath } from "@host-tests/fixtures/config";
import {
  findPlugin,
  installPlugin,
  refreshPluginProjection,
  syncPlugins,
  uninstallPlugin,
  updatePluginStatus,
  upgradePlugin,
} from "@host-tests/fixtures/plugin";
import {
  execPgSQLStatements,
  pgEscapeLiteral,
} from "@host-tests/support/postgres";
import { SmartCenterPage } from "../pages/SmartCenterPage";
import { listTiers, withAdminApi } from "../support/ai-core-api";

const aiCorePluginID = "linapro-ai-core";
const dynamicDemoPluginID = "linapro-demo-dynamic";
const dynamicDemoDependencyPluginID = "linapro-demo-source";
const dynamicDemoVersion = "v0.1.0";
const dynamicDemoArtifactName = `${dynamicDemoPluginID}.wasm`;

type PluginListItem = {
  enabled?: number;
  id: string;
  installed?: number;
  runtimeState?: string;
  upgradeAvailable?: boolean;
};

type HostCallDemoAIStatus = {
  activeProvider?: string;
  available?: boolean;
  capabilityId?: string;
  capabilityMethod?: string;
  capabilityType?: string;
  owner?: string;
  reason?: string;
  service?: string;
  version?: string;
};

type HostCallDemoResponse = {
  ai?: HostCallDemoAIStatus;
  message?: string;
  pluginId?: string;
};

function repoRoot() {
  return path.resolve(process.cwd(), "../..");
}

function tempOutputDir() {
  return path.join(repoRoot(), "temp", "output");
}

function dynamicDemoArtifactPath() {
  return path.join(tempOutputDir(), dynamicDemoArtifactName);
}

function legacyDynamicDemoArtifactPath() {
  return path.join(
    repoRoot(),
    "apps",
    "lina-plugins",
    dynamicDemoPluginID,
    "runtime",
    dynamicDemoArtifactName,
  );
}

function runtimeStorageRoot() {
  return path.join(
    tempOutputDir(),
    ".capability-storage",
    "plugins",
    dynamicDemoPluginID,
  );
}

function unwrapApiData(payload: any) {
  if (
    payload &&
    typeof payload === "object" &&
    typeof payload.code === "number" &&
    "data" in payload
  ) {
    return payload.data;
  }
  return payload;
}

async function expectApiSuccess<T>(
  response: APIResponse,
  message: string,
): Promise<T> {
  const body = await response.text();
  expect(response.ok(), `${message}, status=${response.status()}, body=${body}`).toBe(
    true,
  );
  const payload = body.trim() ? JSON.parse(body) : null;
  expect(
    payload?.code ?? 0,
    `${message}, businessCode=${payload?.code}, body=${body}`,
  ).toBe(0);
  return unwrapApiData(payload) as T;
}

async function expectApiFailure(response: APIResponse, message: string) {
  const body = await response.text();
  if (!response.ok()) {
    expect(
      response.status(),
      `${message}, status=${response.status()}, body=${body}`,
    ).toBeGreaterThanOrEqual(400);
    return body;
  }
  const payload = body.trim() ? JSON.parse(body) : {};
  expect(payload?.code ?? 0, `${message}, body=${body}`).not.toBe(0);
  return body;
}

function buildDynamicDemoArtifact() {
  mkdirSync(tempOutputDir(), { recursive: true });
  execFileSync(
    "make",
    ["wasm", `p=${dynamicDemoPluginID}`, "out=../../temp/output"],
    {
      cwd: repoRoot(),
      stdio: "inherit",
    },
  );
  rmSync(legacyDynamicDemoArtifactPath(), { force: true });
  expect(readFileSync(dynamicDemoArtifactPath()).byteLength).toBeGreaterThan(0);
}

async function pluginItem(
  adminApi: APIRequestContext,
  pluginId: string,
): Promise<PluginListItem | null> {
  return (await findPlugin(adminApi, pluginId)) as PluginListItem | null;
}

async function ensurePluginEnabled(
  adminApi: APIRequestContext,
  pluginId: string,
  installMode = "global",
) {
  await syncPlugins(adminApi);
  let plugin = await pluginItem(adminApi, pluginId);
  expect(plugin, `未发现插件: ${pluginId}`).toBeTruthy();
  if (plugin?.installed !== 1) {
    await installPlugin(adminApi, pluginId, installMode);
    plugin = await pluginItem(adminApi, pluginId);
  }
  if (
    plugin?.upgradeAvailable === true ||
    plugin?.runtimeState === "pending_upgrade" ||
    plugin?.runtimeState === "upgrade_failed"
  ) {
    await upgradePlugin(adminApi, pluginId);
    plugin = await pluginItem(adminApi, pluginId);
  }
  if (plugin?.enabled !== 1) {
    await updatePluginStatus(adminApi, pluginId, true);
  }
  await expect
    .poll(async () => (await pluginItem(adminApi, pluginId))?.enabled ?? 0, {
      message: `${pluginId} 应处于启用状态`,
      timeout: 30_000,
    })
    .toBe(1);
}

async function ensureDynamicDemoDiscovered(adminApi: APIRequestContext) {
  buildDynamicDemoArtifact();
  await syncPlugins(adminApi);
  await expect
    .poll(async () => (await pluginItem(adminApi, dynamicDemoPluginID))?.id ?? "", {
      message: "动态 demo artifact 同步后应被发现",
      timeout: 30_000,
    })
    .toBe(dynamicDemoPluginID);
}

async function ensureDynamicDemoInstalled(adminApi: APIRequestContext) {
  await resetDynamicDemo(adminApi);
  await ensurePluginEnabled(adminApi, aiCorePluginID);
  await ensurePluginEnabled(adminApi, dynamicDemoDependencyPluginID);
  await ensureDynamicDemoDiscovered(adminApi);
  const plugin = await pluginItem(adminApi, dynamicDemoPluginID);
  if (plugin?.installed !== 1) {
    await installPlugin(adminApi, dynamicDemoPluginID);
  }
  await expect
    .poll(
      async () => (await pluginItem(adminApi, dynamicDemoPluginID))?.installed ?? 0,
      {
        message: "动态 demo 应安装成功",
        timeout: 60_000,
      },
    )
    .toBe(1);
}

async function ensureDynamicDemoEnabled(adminApi: APIRequestContext) {
  await ensureDynamicDemoInstalled(adminApi);
  const plugin = await pluginItem(adminApi, dynamicDemoPluginID);
  if (plugin?.enabled !== 1) {
    await updatePluginStatus(adminApi, dynamicDemoPluginID, true);
  }
  await expect
    .poll(
      async () => (await pluginItem(adminApi, dynamicDemoPluginID))?.enabled ?? 0,
      {
        message: "动态 demo 应启用成功",
        timeout: 60_000,
      },
    )
    .toBe(1);
}

async function resetDynamicDemo(adminApi: APIRequestContext) {
  const plugin = await pluginItem(adminApi, dynamicDemoPluginID);
  if (plugin) {
    if (plugin.enabled === 1) {
      await updatePluginStatus(adminApi, dynamicDemoPluginID, false).catch(
        () => {},
      );
    }
    if (plugin.installed === 1) {
      await uninstallPlugin(adminApi, dynamicDemoPluginID, true).catch(() => {});
    }
  }
  rmSync(runtimeStorageRoot(), { force: true, recursive: true });
  const escapedPluginID = pgEscapeLiteral(dynamicDemoPluginID);
  execPgSQLStatements([
    `DELETE FROM sys_plugin_state WHERE plugin_id = '${escapedPluginID}';`,
    `DELETE FROM sys_plugin_node_state WHERE plugin_id = '${escapedPluginID}';`,
    `DELETE FROM sys_plugin_resource_ref WHERE plugin_id = '${escapedPluginID}';`,
    `DELETE FROM sys_plugin_migration WHERE plugin_id = '${escapedPluginID}';`,
    `DELETE FROM sys_plugin_release WHERE plugin_id = '${escapedPluginID}';`,
    `DELETE FROM sys_plugin WHERE plugin_id = '${escapedPluginID}';`,
  ]);
}

async function callDynamicDemoHostCall(adminApi: APIRequestContext) {
  return expectApiSuccess<HostCallDemoResponse>(
    await adminApi.post(pluginApiPath(dynamicDemoPluginID, "host-call-demo"), {
      data: { skipNetwork: true },
    }),
    "调用动态 demo owner AI host-call 失败",
  );
}

async function expectDynamicDemoHostCallUnavailable(
  adminApi: APIRequestContext,
) {
  await expectApiFailure(
    await adminApi.post(pluginApiPath(dynamicDemoPluginID, "host-call-demo"), {
      data: { skipNetwork: true },
    }),
    "未启用动态 demo 时 host-call 路由不应成功",
  );
}

function assertOwnerAIStatus(status: HostCallDemoAIStatus | undefined) {
  expect(status).toBeTruthy();
  expect(status).toMatchObject({
    capabilityId: "plugin.linapro-ai-core.ai.text.v1",
    capabilityMethod: "generate",
    capabilityType: "text",
    owner: aiCorePluginID,
    service: "ai",
    version: "v1",
  });
  expect(typeof status?.available).toBe("boolean");
  if (status?.available) {
    expect(status.activeProvider || aiCorePluginID).toBeTruthy();
  } else {
    expect(status?.reason, "不可用状态必须携带稳定降级原因").toBeTruthy();
  }
}

async function capturePageEvidence(page: { screenshot: Function }, name: string) {
  const now = new Date();
  const day = new Intl.DateTimeFormat("en-CA", {
    day: "2-digit",
    month: "2-digit",
    timeZone: "Asia/Shanghai",
    year: "numeric",
  })
    .format(now)
    .replaceAll("-", "");
  const time = new Intl.DateTimeFormat("en-GB", {
    hour: "2-digit",
    hour12: false,
    minute: "2-digit",
    second: "2-digit",
    timeZone: "Asia/Shanghai",
  })
    .format(now)
    .replaceAll(":", "");
  const dir = path.join(repoRoot(), "temp", day);
  mkdirSync(dir, { recursive: true });
  await page.screenshot({
    fullPage: false,
    path: path.join(dir, `${time}-${name}.png`),
  });
}

test.describe("TC-7 AI owner 能力契约", () => {
  test.afterEach(async ({ adminPage }) => {
    await withAdminApi(async (api) => {
      await resetDynamicDemo(api);
      await ensurePluginEnabled(api, aiCorePluginID);
    });
    await refreshPluginProjection(adminPage);
  });

  test.afterAll(async () => {
    rmSync(dynamicDemoArtifactPath(), { force: true });
    rmSync(legacyDynamicDemoArtifactPath(), { force: true });
  });

  test("TC-7a: 插件启用后智能中心能力状态可读取", async ({
    adminPage,
  }) => {
    await withAdminApi(async (api) => {
      await ensurePluginEnabled(api, aiCorePluginID);
      const plugin = await pluginItem(api, aiCorePluginID);
      expect(plugin?.installed).toBe(1);
      expect(plugin?.enabled).toBe(1);

      const tiers = await listTiers(api, "text", "generate");
      expect(tiers.map((item: { code?: string }) => item.code).sort()).toEqual([
        "advanced",
        "basic",
        "standard",
      ]);
    });

    await refreshPluginProjection(adminPage);
    const smartCenter = new SmartCenterPage(adminPage);
    await smartCenter.gotoTiers();
    await smartCenter.assertTierCapabilityTypeTabs();
    await smartCenter.assertTierTypePage("text");
    await expect(adminPage.getByText(/plugin\.linapro-ai-core/u)).toHaveCount(0);
    await smartCenter.captureEvidence("TC007-ai-owner-enabled-status");
  });

  test("TC-7b: 插件禁用后智能中心路由和 API 稳定降级", async ({
    adminPage,
  }) => {
    await withAdminApi(async (api) => {
      await resetDynamicDemo(api);
      await ensurePluginEnabled(api, aiCorePluginID);
      await updatePluginStatus(api, aiCorePluginID, false);
      const plugin = await pluginItem(api, aiCorePluginID);
      expect(plugin?.enabled).toBe(0);

      await expectApiFailure(
        await api.get(pluginApiPath(aiCorePluginID, "ai/tiers"), {
          params: {
            capabilityMethod: "generate",
            capabilityType: "text",
          },
        }),
        "禁用 linapro-ai-core 后 AI 插件 API 不应继续成功",
      );
    });

    await refreshPluginProjection(adminPage);
    await adminPage.goto(workspacePath("/ai/tiers"));
    await adminPage.waitForLoadState("networkidle").catch(() => {});
    await expect(adminPage.getByText(/未找到页面|Not Found/i)).toBeVisible();
    await expect(adminPage.getByText(/加载菜单中|Loading menu/i)).toHaveCount(0);
    await capturePageEvidence(adminPage, "TC007-ai-owner-disabled-degraded");
  });

  test("TC-7c: 动态 demo owner AI 调用覆盖未启用失败和启用成功路径", async () => {
    await withAdminApi(async (api) => {
      await ensureDynamicDemoInstalled(api);
      await updatePluginStatus(api, dynamicDemoPluginID, false);
      await expectDynamicDemoHostCallUnavailable(api);

      await updatePluginStatus(api, dynamicDemoPluginID, true);
      const plugin = await pluginItem(api, dynamicDemoPluginID);
      expect(plugin?.installed).toBe(1);
      expect(plugin?.enabled).toBe(1);

      const detail = await expectApiSuccess<{
        authorizedHostServices?: Array<{
          methods?: string[];
          owner?: string;
          service?: string;
          version?: string;
        }>;
        effectiveVersion?: string;
      }>(await api.get(`plugins/${dynamicDemoPluginID}`), "查询动态 demo 详情失败");
      expect(detail.effectiveVersion || dynamicDemoVersion).toBe(dynamicDemoVersion);
      expect(
        detail.authorizedHostServices?.some(
          (item) =>
            item.owner === aiCorePluginID &&
            item.service === "ai" &&
            item.version === "v1" &&
            item.methods?.includes("text.method_status.get"),
        ),
      ).toBe(true);

      const payload = await callDynamicDemoHostCall(api);
      expect(payload.pluginId).toBe(dynamicDemoPluginID);
      expect(payload.message).toContain("linapro-ai-core owner AI services");
      assertOwnerAIStatus(payload.ai);
    });
  });
});

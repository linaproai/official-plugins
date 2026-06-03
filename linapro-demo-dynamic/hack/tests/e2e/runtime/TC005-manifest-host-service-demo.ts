import type { APIRequestContext } from "@host-tests/support/playwright";

import { execFileSync } from "node:child_process";
import { copyFileSync, existsSync, rmSync, statSync } from "node:fs";
import path from "node:path";

import { test, expect } from "@host-tests/fixtures/auth";
import {
  createAdminApiContext,
  disablePlugin,
  enablePlugin,
  expectSuccess,
  getPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
} from "@host-tests/support/api/job";
import { waitForRouteReady } from "@host-tests/support/ui";
import { DemoDynamicPage } from "../../pages/DemoDynamicPage";

const pluginID = "linapro-demo-dynamic";
const sourcePluginID = "linapro-demo-source";
const pluginMenuNamePattern = /Dynamic Plugin Demo|动态插件示例/u;
const repoRoot = path.resolve(process.cwd(), "../..");
const pluginDir = path.join(repoRoot, "apps", "lina-plugins", pluginID);
const manifestConfigPath = path.join(
  pluginDir,
  "manifest",
  "config",
  "config.yaml",
);
const manifestConfigTemplatePath = path.join(
  pluginDir,
  "manifest",
  "config",
  "config.example.yaml",
);
const legacyRuntimeArtifactPath = path.join(
  repoRoot,
  "apps",
  "lina-plugins",
  pluginID,
  "runtime",
  `${pluginID}.wasm`,
);
const dynamicHostServiceAuthorization = {
  services: [
    {
      service: "storage",
      methods: ["put", "get", "delete", "list", "stat"],
      paths: ["host-call-demo/", "demo-record-files/"],
    },
    {
      service: "network",
      methods: ["request"],
      resourceRefs: ["https://example.com"],
    },
    {
      service: "data",
      methods: ["list", "get", "create", "update", "delete", "transaction"],
      tables: ["sys_plugin_node_state", "plugin_linapro_demo_dynamic_record"],
    },
    {
      service: "manifest",
      methods: ["get"],
      paths: ["config/config.yaml", "config/profile.yaml"],
    },
    {
      service: "hostConfig",
      methods: ["get"],
      keys: ["workspace.basePath", "i18n.default", "i18n.enabled"],
    },
  ],
};

let adminApi: APIRequestContext;
let originalInstalled = 0;
let originalEnabled = 0;
let originalSourceInstalled = 0;
let originalSourceEnabled = 0;
let createdManifestConfigFixture = false;

function ensureRuntimePluginArtifact() {
  createdManifestConfigFixture = prepareIgnoredManifestConfigFixture();
  execFileSync("make", ["wasm", `p=${pluginID}`, "out=../../temp/output"], {
    cwd: repoRoot,
    stdio: "inherit",
  });
  rmSync(legacyRuntimeArtifactPath, { force: true });
}

function prepareIgnoredManifestConfigFixture() {
  if (existsSync(manifestConfigPath)) {
    if (statSync(manifestConfigPath).isDirectory()) {
      throw new Error(
        `dynamic demo manifest config path is a directory: ${manifestConfigPath}`,
      );
    }
    return false;
  }

  copyFileSync(manifestConfigTemplatePath, manifestConfigPath);
  return true;
}

function cleanupIgnoredManifestConfigFixture() {
  if (!createdManifestConfigFixture) {
    return;
  }
  rmSync(manifestConfigPath, { force: true });
  createdManifestConfigFixture = false;
}

async function installDynamicPluginWithAuthorization() {
  return expectSuccess<{ id: string; installed: number; enabled: number }>(
    await adminApi.post(`plugins/${pluginID}/install`, {
      data: {
        installMode: "global",
        authorization: dynamicHostServiceAuthorization,
      },
    }),
  );
}

async function enableDynamicPluginWithAuthorization() {
  return expectSuccess<{ id: string; enabled: number }>(
    await adminApi.put(`plugins/${pluginID}/enable`, {
      data: {
        authorization: dynamicHostServiceAuthorization,
      },
    }),
  );
}

async function ensurePluginInstalledAndEnabled() {
  await syncPlugins(adminApi);
  let sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  let plugin = await getPlugin(adminApi, pluginID);
  originalSourceInstalled = sourcePlugin.installed;
  originalSourceEnabled = sourcePlugin.enabled;
  originalInstalled = plugin.installed;
  originalEnabled = plugin.enabled;

  if (sourcePlugin.installed !== 1) {
    await installPlugin(adminApi, sourcePluginID, { installMode: "global" });
    sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  }
  if (sourcePlugin.enabled !== 1) {
    await enablePlugin(adminApi, sourcePluginID);
  }

  if (plugin.installed === 1) {
    await forceUninstallPlugin(pluginID, false);
    await syncPlugins(adminApi);
    plugin = await getPlugin(adminApi, pluginID);
  }
  await installDynamicPluginWithAuthorization();
  plugin = await getPlugin(adminApi, pluginID);
  if (plugin.enabled !== 1) {
    await enableDynamicPluginWithAuthorization();
  }
}

async function restorePluginState() {
  let plugin = await getPlugin(adminApi, pluginID);

  if (originalInstalled !== 1) {
    if (plugin.installed === 1) {
      await forceUninstallPlugin(pluginID, true);
    }
    await restoreSourcePluginState();
    return;
  }

  if (originalEnabled !== 1 && plugin.enabled === 1) {
    await forceUninstallPlugin(pluginID, false);
    await syncPlugins(adminApi);
    plugin = await getPlugin(adminApi, pluginID);
  }
  if (plugin.installed !== 1) {
    await installDynamicPluginWithAuthorization();
    plugin = await getPlugin(adminApi, pluginID);
  }
  if (originalEnabled === 1 && plugin.enabled !== 1) {
    await enableDynamicPluginWithAuthorization();
  }

  await restoreSourcePluginState();
}

async function forceUninstallPlugin(pluginId: string, purgeStorageData: boolean) {
  await expectSuccess(
    await adminApi.delete(`plugins/${pluginId}`, {
      params: {
        force: true,
        purgeStorageData: purgeStorageData ? 1 : 0,
      },
    }),
  );
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

test.describe("TC-5 Manifest host service demo", () => {
  test.beforeAll(async () => {
    try {
      ensureRuntimePluginArtifact();
      adminApi = await createAdminApiContext();
      await ensurePluginInstalledAndEnabled();
    } catch (error) {
      cleanupIgnoredManifestConfigFixture();
      throw error;
    }
  });

  test.afterAll(async () => {
    const api = adminApi as APIRequestContext | undefined;
    try {
      if (api) {
        await restorePluginState();
      }
    } finally {
      cleanupIgnoredManifestConfigFixture();
      if (api) {
        await api.dispose();
      }
    }
  });

  test("TC-5a: Manifest declaration is visible through the dynamic plugin page", async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage("English");
    await adminPage.reload({ waitUntil: "domcontentloaded" });
    await waitForRouteReady(adminPage);

    const pluginPage = new DemoDynamicPage(adminPage);
    await pluginPage.clickSidebarMenuItem(pluginMenuNamePattern);
    await waitForRouteReady(adminPage);

    await expect(pluginPage.pluginDemoDynamicManifestDemo()).toBeVisible();
    await expect(pluginPage.pluginDemoDynamicManifestProfilePath()).toContainText(
      "config/profile.yaml",
    );
    await expect(pluginPage.pluginDemoDynamicManifestProfileName()).toContainText(
      "demo-dynamic-profile",
    );
    await expect(pluginPage.pluginDemoDynamicManifestConfigPath()).toContainText(
      "config/config.yaml",
    );
    await expect(pluginPage.pluginDemoDynamicManifestConfigPreview()).toContainText(
      "Hello from dynamic plugin",
    );
  });
});

import type { APIRequestContext } from "@host-tests/support/playwright";

import {
  disablePlugin,
  enablePlugin,
  getPlugin,
  installPlugin,
  uninstallPlugin,
} from "@host-tests/support/api/job";

/** Hard plugin dependencies declared by linapro-demo-dynamic/plugin.yaml. */
export const demoDynamicDependencyPluginIDs = [
  "linapro-ai-core",
  "linapro-demo-source",
] as const;

export type DemoDynamicDependencyPluginID =
  (typeof demoDynamicDependencyPluginIDs)[number];

export type DependencyPluginState = {
  id: DemoDynamicDependencyPluginID;
  installed: number;
  enabled: number;
};

/**
 * Install and optionally enable every hard dependency required by
 * linapro-demo-dynamic before the dynamic plugin itself is installed.
 * Host dependency checks never auto-install declared plugins.
 */
export async function ensureDemoDynamicDependenciesInstalled(
  adminApi: APIRequestContext,
  options: { enable?: boolean } = {},
) {
  const enable = options.enable ?? false;
  for (const dependencyID of demoDynamicDependencyPluginIDs) {
    let dependency = await getPlugin(adminApi, dependencyID);
    if (dependency.installed !== 1) {
      await installPlugin(adminApi, dependencyID, { installMode: "global" });
      dependency = await getPlugin(adminApi, dependencyID);
    }
    if (enable && dependency.enabled !== 1) {
      await enablePlugin(adminApi, dependencyID);
    }
  }
}

/** Capture installed/enabled state for later restore after a test suite. */
export async function captureDemoDynamicDependencyStates(
  adminApi: APIRequestContext,
): Promise<DependencyPluginState[]> {
  const states: DependencyPluginState[] = [];
  for (const id of demoDynamicDependencyPluginIDs) {
    const plugin = await getPlugin(adminApi, id);
    states.push({
      id,
      installed: plugin.installed ?? 0,
      enabled: plugin.enabled ?? 0,
    });
  }
  return states;
}

/**
 * Restore dependency plugins to the captured pre-suite state so later suites
 * are not polluted by installs performed only for linapro-demo-dynamic.
 */
export async function restoreDemoDynamicDependencyStates(
  adminApi: APIRequestContext,
  originalStates: DependencyPluginState[],
) {
  // Restore dependents first is not required here because only hard deps of
  // demo-dynamic are tracked and they do not depend on each other.
  for (const original of originalStates) {
    let plugin = await getPlugin(adminApi, original.id);

    if (original.installed !== 1) {
      if (plugin.enabled === 1) {
        await disablePlugin(adminApi, original.id);
        plugin = await getPlugin(adminApi, original.id);
      }
      if (plugin.installed === 1) {
        await uninstallPlugin(adminApi, original.id);
      }
      continue;
    }

    if (plugin.installed !== 1) {
      await installPlugin(adminApi, original.id, { installMode: "global" });
      plugin = await getPlugin(adminApi, original.id);
    }
    if (original.enabled === 1 && plugin.enabled !== 1) {
      await enablePlugin(adminApi, original.id);
    } else if (original.enabled !== 1 && plugin.enabled === 1) {
      await disablePlugin(adminApi, original.id);
    }
  }
}

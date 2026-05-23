import type { APIRequestContext } from '@host-tests/support/playwright';

import { execFileSync } from 'node:child_process';
import { rmSync } from 'node:fs';
import path from 'node:path';

import { test, expect } from '@host-tests/fixtures/auth';
import { MenuPage } from '@host-tests/pages/MenuPage';
import {
  createAdminApiContext,
  disablePlugin,
  enablePlugin,
  expectSuccess,
  getPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
} from '@host-tests/support/api/job';

const pluginID = 'linapro-demo-dynamic';
const sourcePluginID = 'linapro-demo-source';
const repoRoot = path.resolve(process.cwd(), '../..');
const runtimeArtifactPath = path.join(
  repoRoot,
  'temp',
  'output',
  `${pluginID}.wasm`,
);
const legacyRuntimeArtifactPath = path.join(
  repoRoot,
  'apps',
  'lina-plugins',
  pluginID,
  'runtime',
  `${pluginID}.wasm`,
);

type MenuNode = {
  children?: MenuNode[];
  id: number;
  name: string;
  path?: string;
  perms?: string;
  sort?: number;
  type: string;
};

type FlatMenuNode = {
  ancestors: MenuNode[];
  node: MenuNode;
};

let adminApi: APIRequestContext;
let originalInstalled = 0;
let originalEnabled = 0;
let originalSourceInstalled = 0;
let originalSourceEnabled = 0;

function ensureRuntimePluginArtifact() {
  execFileSync('make', ['wasm', `p=${pluginID}`, 'out=../../temp/output'], {
    cwd: repoRoot,
    stdio: 'inherit',
  });
  rmSync(legacyRuntimeArtifactPath, { force: true });
}

function flattenMenus(nodes: MenuNode[], ancestors: MenuNode[] = []): FlatMenuNode[] {
  return nodes.flatMap((node) => [
    { ancestors, node },
    ...flattenMenus(node.children ?? [], [...ancestors, node]),
  ]);
}

async function ensurePluginInstalledAndEnabled() {
  await syncPlugins(adminApi);
  let sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  let plugin = await getPlugin(adminApi, pluginID);
  originalSourceInstalled = sourcePlugin.installed;
  originalSourceEnabled = sourcePlugin.enabled;
  originalInstalled = plugin.installed;
  originalEnabled = plugin.enabled;

  if (plugin.enabled === 1) {
    await disablePlugin(adminApi, pluginID);
    plugin = await getPlugin(adminApi, pluginID);
  }
  if (plugin.installed === 1) {
    await uninstallPlugin(adminApi, pluginID);
  }

  sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  if (sourcePlugin.enabled === 1) {
    await disablePlugin(adminApi, sourcePluginID);
    sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  }
  if (sourcePlugin.installed === 1) {
    await uninstallPlugin(adminApi, sourcePluginID);
  }

  await syncPlugins(adminApi);
  sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  plugin = await getPlugin(adminApi, pluginID);

  if (sourcePlugin.installed !== 1) {
    await installPlugin(adminApi, sourcePluginID, { installMode: 'global' });
    sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  }
  if (sourcePlugin.enabled !== 1) {
    await enablePlugin(adminApi, sourcePluginID);
  }

  if (plugin.installed !== 1) {
    await installPlugin(adminApi, pluginID, { installMode: 'global' });
    plugin = await getPlugin(adminApi, pluginID);
  }
  if (plugin.enabled !== 1) {
    await enablePlugin(adminApi, pluginID);
  }
}

async function restorePluginState() {
  let plugin = await getPlugin(adminApi, pluginID);

  if (originalInstalled !== 1) {
    if (plugin.enabled === 1) {
      await disablePlugin(adminApi, pluginID);
      plugin = await getPlugin(adminApi, pluginID);
    }
    if (plugin.installed === 1) {
      await uninstallPlugin(adminApi, pluginID);
    }
  } else {
    if (plugin.installed !== 1) {
      await installPlugin(adminApi, pluginID, { installMode: 'global' });
      plugin = await getPlugin(adminApi, pluginID);
    }
    if (originalEnabled === 1 && plugin.enabled !== 1) {
      await enablePlugin(adminApi, pluginID);
    } else if (originalEnabled !== 1 && plugin.enabled === 1) {
      await disablePlugin(adminApi, pluginID);
    }
  }

  let sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  if (originalSourceInstalled !== 1) {
    if (sourcePlugin.enabled === 1) {
      await disablePlugin(adminApi, sourcePluginID);
      sourcePlugin = await getPlugin(adminApi, sourcePluginID);
    }
    if (sourcePlugin.installed === 1) {
      await uninstallPlugin(adminApi, sourcePluginID);
    }
  } else {
    if (sourcePlugin.installed !== 1) {
      await installPlugin(adminApi, sourcePluginID, { installMode: 'global' });
      sourcePlugin = await getPlugin(adminApi, sourcePluginID);
    }
    if (originalSourceEnabled === 1 && sourcePlugin.enabled !== 1) {
      await enablePlugin(adminApi, sourcePluginID);
    } else if (originalSourceEnabled !== 1 && sourcePlugin.enabled === 1) {
      await disablePlugin(adminApi, sourcePluginID);
    }
  }
}

test.describe('TC-4 Dynamic plugin permission menu tree regression', () => {
  test.beforeAll(async () => {
    ensureRuntimePluginArtifact();
    adminApi = await createAdminApiContext();
    await ensurePluginInstalledAndEnabled();
  });

  test.afterAll(async () => {
    try {
      await restorePluginState();
    } finally {
      await adminApi.dispose();
    }
  });

  test('TC-4a: Dynamic route permission buttons are children of the plugin menu', async () => {
    const menuData = await expectSuccess<{ list: MenuNode[] }>(
      await adminApi.get('menu'),
    );
    const extensionMenu = menuData.list.find(
      (node) => node.path === 'extension' && node.type === 'D',
    );
    expect(extensionMenu, 'missing extension center menu').toBeTruthy();
    const extensionChildren = (extensionMenu?.children ?? []).filter(
      (node) => node.type !== 'B',
    );
    const pluginManagementIndex = extensionChildren.findIndex(
      (node) => node.path === '/system/plugin',
    );
    const sourceDemoIndex = extensionChildren.findIndex(
      (node) => node.path === 'linapro-demo-source-sidebar-entry',
    );
    const dynamicDemoIndex = extensionChildren.findIndex(
      (node) => (node.path ?? '').includes(`/x-assets/${pluginID}/`),
    );
    expect(
      pluginManagementIndex,
      'plugin management menu should be under extension',
    ).toBeGreaterThanOrEqual(0);
    expect(
      sourceDemoIndex,
      'source demo menu should be under extension',
    ).toBeGreaterThan(pluginManagementIndex);
    expect(
      dynamicDemoIndex,
      'dynamic demo menu should be under extension',
    ).toBeGreaterThan(pluginManagementIndex);
    expect(
      dynamicDemoIndex,
      'dynamic demo should be ordered after source demo',
    ).toBeGreaterThan(sourceDemoIndex);

    const flatMenus = flattenMenus(menuData.list);
    const pluginMenu = flatMenus.find(
      ({ node }) =>
        node.perms === `${pluginID}:view` ||
        (node.path ?? '').includes(`/x-assets/${pluginID}/`),
    );
    expect(pluginMenu, 'missing dynamic plugin main menu').toBeTruthy();

    const dynamicRouteButtons = flatMenus.filter(({ node }) => {
      return (
        node.type === 'B' &&
        (node.perms ?? '').startsWith(`${pluginID}:`) &&
        node.perms !== `${pluginID}:view`
      );
    });
    expect(dynamicRouteButtons.length).toBeGreaterThan(0);

    for (const item of dynamicRouteButtons) {
      expect(
        item.ancestors.map((ancestor) => ancestor.id),
        `${item.node.name} should be nested below the dynamic plugin menu`,
      ).toContain(pluginMenu!.node.id);
    }
  });

  test('TC-4b: Menu tree expandable names show pointer cursor and toggle on click', async ({
    adminPage,
    mainLayout,
  }) => {
    const menuPage = new MenuPage(adminPage);

    await mainLayout.switchLanguage('简体中文');
    await menuPage.goto();
    await menuPage.collapseAll();

    const accessRow = adminPage
      .locator('.vxe-body--row:visible', { hasText: '权限管理' })
      .first();
    await expect(accessRow).toBeVisible();

    const accessNameCell = accessRow
      .locator('.system-menu-name-column .vxe-cell')
      .first();
    await accessNameCell.hover();
    await expect
      .poll(async () =>
        accessNameCell.evaluate((node) => getComputedStyle(node).cursor),
      )
      .toBe('pointer');

    const childUserRow = adminPage
      .locator('.vxe-body--row:visible', { hasText: '用户管理' })
      .first();
    await expect(childUserRow).toBeHidden();

    await accessNameCell.click();
    await expect(childUserRow).toBeVisible();

    await accessNameCell.click();
    await expect(childUserRow).toBeHidden();
  });

  test('TC-4c: Dynamic plugin button names are readable in English menu management', async ({
    adminPage,
    mainLayout,
  }) => {
    const menuPage = new MenuPage(adminPage);

    await mainLayout.switchLanguage('English');
    await menuPage.goto();

    const searchInput = adminPage.getByRole('textbox', {
      name: /菜单名称|Menu Name/i,
    });
    await searchInput.fill(
      'Dynamic Route Permission:linapro-demo-dynamic:record:create',
    );
    await adminPage.getByRole('button', { name: /搜索|Search/i }).click();

    await expect(
      adminPage.locator('.vxe-body--row:visible', { hasText: 'Record Create' }),
    ).toBeVisible();

    await expect(
      adminPage.locator('.vxe-body--row:visible', {
        hasText: /Dynamic Route Permission:linapro-demo-dynamic/u,
      }),
    ).toHaveCount(0);
  });
});

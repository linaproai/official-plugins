/**
 * TC001 linapro-oidc-generic 菜单挂载与登录入口
 *
 * - 依赖 linapro-extlogin-core 的授权登录目录
 * - 设置菜单挂在该目录下
 * - 未配置 Client 凭证时登录入口 fail-closed 回登录页
 */
import type { APIRequestContext } from '@host-tests/support/playwright';

import { expect, test } from '@host-tests/fixtures/auth';
import {
  createAdminApiContext,
  prepareSourcePluginsBaseline,
} from '@host-tests/fixtures/plugin';
import { LoginPage } from '@host-tests/pages/LoginPage';
import { waitForRouteReady } from '@host-tests/support/ui';

import { GenericOidcPage } from '../pages/GenericOidcPage';

const ownerPluginID = 'linapro-extlogin-core';
const pluginID = 'linapro-oidc-generic';

type MenuNode = {
  children?: MenuNode[];
  name?: string;
  path?: string;
  perms?: string;
};

function findByName(list: MenuNode[], name: string | RegExp): MenuNode | null {
  for (const item of list) {
    const title = String(item.name ?? '');
    if (
      (typeof name === 'string' && title === name) ||
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

async function fetchMenuTree(
  api: APIRequestContext,
  locale?: string,
): Promise<MenuNode[]> {
  const response = await api.get(
    'menu',
    locale ? { headers: { 'Accept-Language': locale } } : undefined,
  );
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

test.describe('TC-1 linapro-oidc-generic 设置菜单与登录入口', () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline([ownerPluginID, pluginID]);
  });

  test('TC-1a: 设置菜单挂在授权登录目录下', async () => {
    const api = await createAdminApiContext();
    try {
      const tree = await fetchMenuTree(api);
      const authLogin = findByName(tree, /授权登录|Auth Login/i);
      expect(authLogin, '依赖的授权登录目录应存在').not.toBeNull();

      const generic = (authLogin?.children ?? []).find(
        (node) =>
          node.path === 'linapro-oidc-generic-settings' ||
          /^OIDC$/i.test(String(node.name ?? '').trim()),
      );
      expect(generic, 'OIDC 菜单应挂在授权登录下').toBeTruthy();
    } finally {
      await api.dispose();
    }
  });

  test('TC-1b: 中英文菜单标题本地化', async () => {
    const api = await createAdminApiContext();
    try {
      const zhTree = await fetchMenuTree(api, 'zh-CN');
      const enTree = await fetchMenuTree(api, 'en-US');
      const zhAuth = findByName(zhTree, /授权登录|Auth Login/i);
      const enAuth = findByName(enTree, /授权登录|Auth Login/i);
      const zhNode = (zhAuth?.children ?? []).find(
        (n) => n.path === 'linapro-oidc-generic-settings',
      );
      const enNode = (enAuth?.children ?? []).find(
        (n) => n.path === 'linapro-oidc-generic-settings',
      );
      expect(zhNode?.name, 'zh-CN 菜单标题').toBe('OIDC');
      expect(enNode?.name, 'en-US 菜单标题').toBe('OIDC');
    } finally {
      await api.dispose();
    }
  });

  test('TC-1c: OIDC 入口使用统一全宽响应式按钮', async ({ page }) => {
    await page.setViewportSize({ height: 844, width: 390 });
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await waitForRouteReady(page);
    if ((await page.locator('html').getAttribute('lang')) !== 'zh-CN') {
      await loginPage.switchLanguage('简体中文');
    }

    const genericPage = new GenericOidcPage(page);
    await expect(genericPage.loginEntry).toBeVisible();
    await expect(genericPage.loginEntryButton).toHaveText('使用 OIDC 登录');

    const layout = await genericPage.getLoginEntryLayout();
    // Match host AuthenticationLogin VbenButton default size (h-9 ≈ 36px).
    expect(layout.buttonHeight).toBeGreaterThanOrEqual(32);
    expect(layout.buttonHeight).toBeLessThanOrEqual(40);
    expect(
      Math.abs(layout.buttonWidth - layout.entryWidth),
    ).toBeLessThanOrEqual(1);
    // Single-line label: content must not overflow / wrap the button box.
    expect(layout.scrollWidth).toBeLessThanOrEqual(layout.clientWidth + 1);
    expect(layout.scrollHeight).toBeLessThanOrEqual(layout.clientHeight + 1);
    expect(layout.buttonRight).toBeLessThanOrEqual(layout.viewportWidth);
  });

  test('TC-1d: 未配置凭证时登录入口 fail-closed', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await waitForRouteReady(page);
    const genericPage = new GenericOidcPage(page);
    await expect(genericPage.loginEntry).toBeVisible();
    await Promise.all([
      page.waitForURL(/auth\/login|externalLogin=1|PLUGIN_OIDC_GENERIC/i, {
        timeout: 15_000,
      }),
      genericPage.loginEntryButton.click(),
    ]);
    expect(page.url()).not.toMatch(/accounts\.google\.com|okta\.com|keycloak/i);
  });
});

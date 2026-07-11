/**
 * TC001 linapro-auth-ldap 菜单与登录入口
 */
import type { APIRequestContext } from '@host-tests/support/playwright';
import { expect, test } from '@host-tests/fixtures/auth';
import {
  createAdminApiContext,
  prepareSourcePluginsBaseline,
} from '@host-tests/fixtures/plugin';
import { LoginPage } from '@host-tests/pages/LoginPage';
import { waitForRouteReady } from '@host-tests/support/ui';
import { LdapAuthPage } from '../pages/LdapAuthPage';

const ownerPluginID = 'linapro-extlogin-core';
const pluginID = 'linapro-auth-ldap';

type MenuNode = { children?: MenuNode[]; name?: string; path?: string };

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
    if (nested) return nested;
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
  if (Array.isArray(data?.list)) return data.list;
  if (Array.isArray(data)) return data;
  return [];
}

test.describe('TC-1 linapro-auth-ldap 设置菜单与登录入口', () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline([ownerPluginID, pluginID]);
  });

  test('TC-1a: 设置菜单挂在授权登录下', async () => {
    const api = await createAdminApiContext();
    try {
      const tree = await fetchMenuTree(api);
      const authLogin = findByName(tree, /授权登录|Auth Login/i);
      expect(authLogin).not.toBeNull();
      const ldap = (authLogin?.children ?? []).find(
        (n) =>
          n.path === 'linapro-auth-ldap-settings' ||
          /^LDAP$/i.test(String(n.name ?? '').trim()),
      );
      expect(ldap, 'LDAP 菜单应挂在授权登录下').toBeTruthy();
    } finally {
      await api.dispose();
    }
  });

  test('TC-1a2: 中英文菜单标题本地化', async () => {
    const api = await createAdminApiContext();
    try {
      const zhTree = await fetchMenuTree(api, 'zh-CN');
      const enTree = await fetchMenuTree(api, 'en-US');
      const zhAuth = findByName(zhTree, /授权登录|Auth Login/i);
      const enAuth = findByName(enTree, /授权登录|Auth Login/i);
      const zhNode = (zhAuth?.children ?? []).find(
        (n) => n.path === 'linapro-auth-ldap-settings',
      );
      const enNode = (enAuth?.children ?? []).find(
        (n) => n.path === 'linapro-auth-ldap-settings',
      );
      expect(zhNode?.name, 'zh-CN 菜单标题').toBe('LDAP');
      expect(enNode?.name, 'en-US 菜单标题').toBe('LDAP');
    } finally {
      await api.dispose();
    }
  });

  test('TC-1b: 目录入口使用统一全宽按钮样式', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await waitForRouteReady(page);
    if ((await page.locator('html').getAttribute('lang')) !== 'zh-CN') {
      await loginPage.switchLanguage('简体中文');
    }

    const ldap = new LdapAuthPage(page);
    await expect(ldap.loginEntry).toBeVisible();
    await expect(ldap.loginEntryButton).toHaveText('使用 LDAP 登录');

    const layout = await ldap.getLoginEntryLayout();
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

  test('TC-1c: 目录凭证使用统一响应式弹层和表单校验', async ({ page }) => {
    await page.setViewportSize({ height: 844, width: 390 });
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await waitForRouteReady(page);
    if ((await page.locator('html').getAttribute('lang')) !== 'zh-CN') {
      await loginPage.switchLanguage('简体中文');
    }

    const ldap = new LdapAuthPage(page);
    await ldap.openLoginModal();
    await expect(ldap.loginModal).toBeVisible();
    await expect(ldap.loginModalTitle).toBeVisible();
    await expect(ldap.usernameInput).toBeVisible();
    await expect(ldap.passwordInput).toBeVisible();

    const modalLayout = await ldap.getLoginModalLayout();
    expect(modalLayout.left).toBeGreaterThanOrEqual(16);
    expect(modalLayout.right).toBeLessThanOrEqual(
      modalLayout.viewportWidth - 16,
    );
    expect(modalLayout.width).toBeLessThanOrEqual(
      modalLayout.viewportWidth - 32,
    );

    // Empty submit: username and password must show distinct required tips.
    await ldap.confirmButton.click();
    await expect(ldap.usernameRequiredMessage).toBeVisible();
    await expect(ldap.passwordRequiredMessage).toBeVisible();
    await expect(
      ldap.loginModal.getByText('请输入用户名和密码', { exact: true }),
    ).toHaveCount(0);

    // Username only: password tip remains; username tip clears.
    await ldap.usernameInput.fill('directory-user');
    await ldap.confirmButton.click();
    await expect(ldap.usernameRequiredMessage).toHaveCount(0);
    await expect(ldap.passwordRequiredMessage).toBeVisible();

    await ldap.passwordInput.fill('not-submitted');
    await ldap.cancelButton.click();
    await expect(ldap.loginModal).toBeHidden();

    await ldap.openLoginModal();
    await expect(ldap.usernameInput).toHaveValue('');
    await expect(ldap.passwordInput).toHaveValue('');
  });
});

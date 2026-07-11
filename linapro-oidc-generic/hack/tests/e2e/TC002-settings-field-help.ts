/**
 * TC002 linapro-oidc-generic 设置页字段帮助提示
 *
 * - 难懂字段标题右侧展示问号图标
 * - 悬停后展示通俗易懂的帮助文案（非原始 i18n key）
 */
import { expect, test } from '@host-tests/fixtures/auth';
import { prepareSourcePluginsBaseline } from '@host-tests/fixtures/plugin';

import { GenericOidcPage } from '../pages/GenericOidcPage';

const ownerPluginID = 'linapro-extid-core';
const pluginID = 'linapro-oidc-generic';

test.describe('TC-2 linapro-oidc-generic 设置页字段帮助', () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline([ownerPluginID, pluginID]);
  });

  test('TC-2a: 难懂字段标题旁有问号，悬停显示帮助文案', async ({
    adminPage,
  }) => {
    const page = new GenericOidcPage(adminPage);
    await page.openSettingsPage();

    // Form card title: "OIDC 设置" (no "通用" prefix).
    await expect(page.settingsCardTitle).toHaveText('OIDC 设置');

    await expect(page.fieldHelpIcons.first()).toBeVisible();
    expect(await page.fieldHelpIcons.count()).toBeGreaterThanOrEqual(5);

    // Default host E2E locale is zh-CN.
    await page.expectFieldHelpTooltip(
      /身份认证服务|客户端编号|认证服务|https:\/\//i,
    );

    // Must not show raw i18n keys.
    const tooltip = adminPage.locator('.ant-tooltip:visible').last();
    await expect(tooltip).not.toContainText('plugin.linapro-oidc-generic');
  });
});

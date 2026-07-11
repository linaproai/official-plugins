/**
 * TC002 linapro-auth-ldap 设置页字段帮助提示
 *
 * - 难懂字段标题右侧展示问号图标
 * - 悬停后展示通俗易懂的帮助文案（非原始 i18n key）
 */
import { expect, test } from '@host-tests/fixtures/auth';
import { prepareSourcePluginsBaseline } from '@host-tests/fixtures/plugin';

import { LdapAuthPage } from '../pages/LdapAuthPage';

const ownerPluginID = 'linapro-extid-core';
const pluginID = 'linapro-auth-ldap';

test.describe('TC-2 linapro-auth-ldap 设置页字段帮助', () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline([ownerPluginID, pluginID]);
  });

  test('TC-2a: 难懂字段标题旁有问号，悬停显示帮助文案', async ({
    adminPage,
  }) => {
    const page = new LdapAuthPage(adminPage);
    await page.openSettingsPage();

    // Form card title matches menu wording: "LDAP 设置".
    await expect(page.settingsCardTitle).toHaveText('LDAP 设置');

    await expect(page.fieldHelpIcons.first()).toBeVisible();
    expect(await page.fieldHelpIcons.count()).toBeGreaterThanOrEqual(6);

    // Default host E2E locale is zh-CN.
    await page.expectFieldHelpTooltip(
      /目录|查找用户|用户名|稳定|属性|ou=people|entryUUID|objectGUID/i,
    );

    const tooltip = adminPage.locator('.ant-tooltip:visible').last();
    await expect(tooltip).not.toContainText('plugin.linapro-auth-ldap');
  });
});

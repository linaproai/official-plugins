/**
 * TC002 linapro-auth-ldap 设置页字段帮助与表单布局
 *
 * - 页面不展示与菜单重复的 Card 标题
 * - 表单为左右布局，核心字段带必填红星
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

  test('TC-2a: 水平表单、必填红星与字段帮助', async ({ adminPage }) => {
    const page = new LdapAuthPage(adminPage);
    await page.openSettingsPage();

    // Menu already identifies the page; the form card must not repeat the title.
    await expect(adminPage.locator('.ant-card-head-title')).toHaveCount(0);

    await expect(page.settingsIntroAlert).toBeVisible();
    await expect(
      page.settingsIntroAlert.locator('.ant-alert-icon'),
    ).toBeVisible();
    await expect(page.settingsIntroAlert).not.toContainText(
      'plugin.linapro-auth-ldap',
    );

    await expect(page.settingsForm).toBeVisible();
    expect(await page.requiredFieldLabels.count()).toBeGreaterThanOrEqual(2);
    await expect(
      page.settingsForm
        .locator('.ant-form-item-required')
        .filter({ hasText: /主机|Host|TLS/i })
        .first(),
    ).toBeVisible();

    const sampleLabel = page.settingsForm
      .locator('.ant-form-item-label > label')
      .filter({ hasText: /主机/ })
      .first();
    await expect(sampleLabel).toBeVisible();
    await expect(sampleLabel).toHaveClass(/ant-form-item-no-colon/);
    const fontWeight = await sampleLabel.evaluate(
      (el) => Number.parseInt(getComputedStyle(el).fontWeight, 10) || 0,
    );
    expect(fontWeight).toBeGreaterThanOrEqual(500);
    const allLabelText = await page.settingsForm
      .locator('.ant-form-item-label > label')
      .allTextContents();
    for (const text of allLabelText) {
      expect(text.trim().endsWith(':')).toBe(false);
      expect(text).not.toMatch(/（可选）|\(optional\)/i);
    }
    // LDAP previously embedded optional markers in bind DN / user DN labels.
    await expect(
      page.settingsForm.getByText(/服务账号 DN/, { exact: false }).first(),
    ).toBeVisible();
    await expect(
      page.settingsForm.getByText(/（可选）|\(optional\)/i),
    ).toHaveCount(0);

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

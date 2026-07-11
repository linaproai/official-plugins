/**
 * TC002 linapro-oidc-discord 设置页字段帮助与表单布局
 *
 * - 页面不展示与菜单重复的 Card 标题
 * - 表单为左右布局，核心字段带必填红星
 * - 难懂字段标题右侧展示问号图标
 * - 悬停后展示通俗易懂的帮助文案（非原始 i18n key）
 */
import { expect, test } from "@host-tests/fixtures/auth";
import { prepareSourcePluginsBaseline } from "@host-tests/fixtures/plugin";

import { DiscordOidcPage } from "../pages/DiscordOidcPage";

const ownerPluginID = "linapro-extlogin-core";
const pluginID = "linapro-oidc-discord";

test.describe("TC-2 linapro-oidc-discord 设置页字段帮助", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline([ownerPluginID, pluginID]);
  });

  test("TC-2a: 水平表单、必填红星与字段帮助", async ({ adminPage }) => {
    const page = new DiscordOidcPage(adminPage);
    await page.openSettingsPage();

    // Menu already identifies the page; the form card must not repeat the title.
    await expect(
      adminPage.locator(".ant-card-head-title"),
    ).toHaveCount(0);

    await expect(page.settingsIntroAlert).toBeVisible();
    await expect(
      page.settingsIntroAlert.locator(".ant-alert-icon"),
    ).toBeVisible();
    await expect(page.settingsIntroAlert).not.toContainText(
      "plugin.linapro-oidc-discord",
    );

    await expect(page.settingsForm).toBeVisible();
    expect(await page.requiredFieldLabels.count()).toBeGreaterThanOrEqual(1);
    await expect(
      page.settingsForm
        .locator(".ant-form-item-required")
        .filter({ hasText: /客户端 ID|Client ID/i })
        .first(),
    ).toBeVisible();

    const sampleLabel = page.settingsForm
      .locator(".ant-form-item-label > label")
      .filter({ hasText: /客户端 ID/ })
      .first();
    await expect(sampleLabel).toBeVisible();
    await expect(sampleLabel).toHaveClass(/ant-form-item-no-colon/);
    const fontWeight = await sampleLabel.evaluate(
      (el) => Number.parseInt(getComputedStyle(el).fontWeight, 10) || 0,
    );
    expect(fontWeight).toBeGreaterThanOrEqual(500);
    const allLabelText = await page.settingsForm
      .locator(".ant-form-item-label > label")
      .allTextContents();
    for (const text of allLabelText) {
      expect(text.trim().endsWith(":")).toBe(false);
      expect(text).not.toMatch(/（可选）|\(optional\)/i);
    }

    await expect(page.fieldHelpIcons.first()).toBeVisible();
    expect(await page.fieldHelpIcons.count()).toBeGreaterThanOrEqual(4);

    // Default host E2E locale is zh-CN.
    await page.expectFieldHelpTooltip(/Discord|客户端|Developer Portal|应用/i);

    const tooltip = adminPage.locator(".ant-tooltip:visible").last();
    await expect(tooltip).not.toContainText("plugin.linapro-oidc-discord");
  });
});

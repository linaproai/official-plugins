/**
 * TC002 linapro-oidc-discord 设置页字段帮助提示
 *
 * - 难懂字段标题右侧展示问号图标
 * - 悬停后展示通俗易懂的帮助文案（非原始 i18n key）
 */
import { expect, test } from "@host-tests/fixtures/auth";
import { prepareSourcePluginsBaseline } from "@host-tests/fixtures/plugin";

import { DiscordOidcPage } from "../pages/DiscordOidcPage";

const ownerPluginID = "linapro-extid-core";
const pluginID = "linapro-oidc-discord";

test.describe("TC-2 linapro-oidc-discord 设置页字段帮助", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline([ownerPluginID, pluginID]);
  });

  test("TC-2a: 难懂字段标题旁有问号，悬停显示帮助文案", async ({
    adminPage,
  }) => {
    const page = new DiscordOidcPage(adminPage);
    await page.openSettingsPage();

    // Form card title: brand login settings (menu stays "Discord 登录").
    await expect(
      adminPage.getByText("Discord 登录设置", { exact: true }),
    ).toBeVisible();

    await expect(page.fieldHelpIcons.first()).toBeVisible();
    expect(await page.fieldHelpIcons.count()).toBeGreaterThanOrEqual(4);

    // Default host E2E locale is zh-CN.
    await page.expectFieldHelpTooltip(/Discord|客户端编号|账号/i);

    const tooltip = adminPage.locator(".ant-tooltip:visible").last();
    await expect(tooltip).not.toContainText("plugin.linapro-oidc-discord");
  });
});

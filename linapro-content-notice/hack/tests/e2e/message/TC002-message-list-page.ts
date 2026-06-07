import { test, expect } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";
import {
  config,
  pluginApiPath,
  workspacePath,
} from "@host-tests/fixtures/config";
import { LoginPage } from "@host-tests/pages/LoginPage";
import {
  createAdminApiContext,
  expectSuccess,
} from "@host-tests/support/api/job";

import {
  createMessageRecipient,
  unreadCount,
} from "../../support/message-recipient";

const PLUGIN_ID = "linapro-content-notice";

test.describe("TC002 用户消息列表页面", () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, "linapro-content-notice");
  });

  test("TC002a: 消息列表页面可访问", async ({ browser }) => {
    const recipient = await createMessageRecipient("tc002a");
    const context = await browser.newContext({ baseURL: config.baseURL });
    const page = await context.newPage();
    const loginPage = new LoginPage(page);

    try {
      await loginPage.goto();
      await loginPage.loginAndWaitForRedirect(
        recipient.username,
        recipient.password,
      );

      await page.goto(workspacePath("/system/message"));
      await page.waitForLoadState("networkidle");

      const card = page.locator(".ant-card");
      await expect(card).toBeVisible({ timeout: 10000 });
      await expect(card.locator(".ant-card-head-title")).toHaveText("消息列表");

      await expect(page.getByRole("button", { name: /全部已读/ })).toBeVisible({
        timeout: 5000,
      });
      await expect(page.getByRole("button", { name: /清空消息/ })).toBeVisible({
        timeout: 5000,
      });
    } finally {
      await recipient.cleanup();
      await context.close();
    }
  });

  test("TC002b: 消息列表展示通知消息", async ({ browser }) => {
    const adminApi = await createAdminApiContext();
    const recipient = await createMessageRecipient("tc002b");

    const title = `消息列表测试_${Date.now()}`;
    const createData = await expectSuccess<{ id: number }>(
      await adminApi.post(pluginApiPath(PLUGIN_ID, "notice"), {
        data: {
          title,
          type: 1,
          content: "消息列表测试内容",
          status: 1,
        },
      }),
    );
    const noticeId = createData.id;
    const context = await browser.newContext({ baseURL: config.baseURL });
    const page = await context.newPage();
    const loginPage = new LoginPage(page);

    try {
      await expect
        .poll(() => unreadCount(recipient.api), {
          message:
            "expected recipient unread-count to include the published notice",
          timeout: 10_000,
        })
        .toBeGreaterThan(0);

      await loginPage.goto();
      await loginPage.loginAndWaitForRedirect(
        recipient.username,
        recipient.password,
      );

      await page.goto(workspacePath("/system/message"));
      await page.waitForLoadState("networkidle");

      const card = page.locator(".ant-card");
      await expect(card).toBeVisible({ timeout: 10000 });
      await expect(card.locator(".ant-card-head-title")).toHaveText("消息列表");
      await expect(page.getByText(title, { exact: true }).first()).toBeVisible({
        timeout: 10_000,
      });
    } finally {
      await adminApi
        .delete(pluginApiPath(PLUGIN_ID, `notice/${noticeId}`))
        .catch(() => {});
      await recipient.cleanup();
      await adminApi.dispose();
      await context.close();
    }
  });
});

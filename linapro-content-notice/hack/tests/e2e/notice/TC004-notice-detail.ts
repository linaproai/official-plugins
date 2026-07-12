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

test.describe("TC004 消息列表预览弹窗查看通知详情", () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, "linapro-content-notice");
  });

  test("TC004a: 从消息列表点击消息弹出预览窗口", async ({ browser }) => {
    const adminApi = await createAdminApiContext();
    const recipient = await createMessageRecipient("tc004a");
    let noticeId = 0;
    const title = `预览测试通知_${Date.now()}`;
    const content = "<p>这是预览测试的通知内容</p>";
    const context = await browser.newContext({ baseURL: config.baseURL });
    const page = await context.newPage();
    const loginPage = new LoginPage(page);

    try {
      const createData = await expectSuccess<{ id: number }>(
        await adminApi.post(pluginApiPath(PLUGIN_ID, "notice"), {
          data: { title, type: 1, content, status: 1 },
        }),
      );
      noticeId = createData.id;

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
      await expect(page.getByText(title, { exact: true }).first()).toBeVisible({
        timeout: 10_000,
      });
      await page.getByText(title, { exact: true }).first().click();

      const modal = page.locator('[role="dialog"]');
      await expect(modal).toBeVisible({ timeout: 10000 });
      await expect(modal.getByText("这是预览测试的通知内容")).toBeVisible({
        timeout: 5000,
      });
      await expect(modal.locator(".ant-descriptions")).toBeVisible({
        timeout: 5000,
      });
    } finally {
      if (noticeId > 0) {
        await adminApi
          .delete(pluginApiPath(PLUGIN_ID, `notice?ids[]=${noticeId}`))
          .catch(() => {});
      }
      await recipient.cleanup();
      await adminApi.dispose();
      await context.close();
    }
  });
});

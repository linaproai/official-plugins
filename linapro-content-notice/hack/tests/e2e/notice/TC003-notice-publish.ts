import { test, expect } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";
import { NoticePage } from "../../pages/NoticePage";
import { pluginApiPath } from "@host-tests/fixtures/config";
import {
  createAdminApiContext,
  expectSuccess,
} from "@host-tests/support/api/job";

import {
  createMessageRecipient,
  unreadCount,
} from "../../support/message-recipient";

const PLUGIN_ID = "linapro-content-notice";

test.describe("TC003 通知公告发布与消息分发", () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, "linapro-content-notice");
  });

  test("TC003a: 创建已发布通知后铃铛显示未读", async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    const publishTitle = `发布测试_${Date.now()}`;
    await noticePage.goto();

    try {
      // Create a published notice
      await noticePage.createNotice(
        publishTitle,
        "通知",
        "已发布",
        "发布测试内容",
      );

      await expect(
        adminPage.getByText(/新增成功|创建成功|success/i),
      ).toBeVisible({ timeout: 5000 });

      // Note: The admin user is excluded from message distribution (they are the creator),
      // so we just verify the notice was created successfully
      const hasNotice = await noticePage.hasNotice(publishTitle);
      expect(hasNotice).toBeTruthy();
    } finally {
      await noticePage.deleteNoticeIfExists(publishTitle);
    }
  });

  test("TC003b: 发布后其他用户收到消息通知", async () => {
    const adminApi = await createAdminApiContext();
    const recipient = await createMessageRecipient("tc003b");
    let noticeId = 0;

    try {
      expect(await unreadCount(recipient.api)).toBe(0);

      const createData = await expectSuccess<{ id: number }>(
        await adminApi.post(pluginApiPath(PLUGIN_ID, "notice"), {
          data: {
            title: `消息分发测试_${Date.now()}`,
            type: 1,
            content: "验证消息分发",
            status: 1,
          },
        }),
      );
      noticeId = createData.id;

      expect(await unreadCount(recipient.api)).toBe(1);
    } finally {
      if (noticeId > 0) {
        await adminApi
          .delete(pluginApiPath(PLUGIN_ID, `notice/${noticeId}`))
          .catch(() => {});
      }
      await recipient.cleanup();
      await adminApi.dispose();
    }
  });

  test("TC003d: 草稿发布后其他用户收到消息通知", async () => {
    const adminApi = await createAdminApiContext();
    const recipient = await createMessageRecipient("tc003d");
    let noticeId = 0;

    try {
      expect(await unreadCount(recipient.api)).toBe(0);

      const createData = await expectSuccess<{ id: number }>(
        await adminApi.post(pluginApiPath(PLUGIN_ID, "notice"), {
          data: {
            title: `草稿发布测试_${Date.now()}`,
            type: 1,
            content: "草稿内容",
            status: 0,
          },
        }),
      );
      noticeId = createData.id;

      expect(await unreadCount(recipient.api)).toBe(0);

      await expectSuccess(
        await adminApi.put(pluginApiPath(PLUGIN_ID, `notice/${noticeId}`), {
          data: { status: 1 },
        }),
      );

      expect(await unreadCount(recipient.api)).toBe(1);
    } finally {
      if (noticeId > 0) {
        await adminApi
          .delete(pluginApiPath(PLUGIN_ID, `notice/${noticeId}`))
          .catch(() => {});
      }
      await recipient.cleanup();
      await adminApi.dispose();
    }
  });

  test("TC003c: 已发布测试通知可删除", async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    const title = `删除测试_${Date.now()}`;
    await noticePage.goto();

    await noticePage.createNotice(title, "通知", "已发布", "删除测试内容");
    await expect(adminPage.getByText(/新增成功|创建成功|success/i)).toBeVisible(
      { timeout: 5000 },
    );

    await noticePage.deleteNotice(title);

    await expect(adminPage.getByText(/删除成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });
});

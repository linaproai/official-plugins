import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { config, pluginApiPath, workspacePath } from '@host-tests/fixtures/config';
import { LoginPage } from '@host-tests/pages/LoginPage';

const PLUGIN_ID = 'linapro-content-notice';
const HOST_API_BASE = config.apiBaseURL.replace(/\/$/, '');
const PLUGIN_API_BASE = `${config.publicBaseURL.replace(/\/$/, '')}${pluginApiPath(PLUGIN_ID)}`;

async function apiLogin(
  username: string,
  password: string,
): Promise<string> {
  const resp = await fetch(`${HOST_API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password, clientType: 'web' }),
    redirect: 'manual',
  });
  const data = await resp.json();
  return data.data.accessToken;
}

async function apiClearMessages(token: string): Promise<void> {
  await fetch(`${HOST_API_BASE}/user/message/clear`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
}

async function apiUnreadCount(
  token: string,
): Promise<number> {
  const resp = await fetch(`${HOST_API_BASE}/user/message/count`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const data = await resp.json();
  return data.data.count;
}

test.describe('TC004 消息列表预览弹窗查看通知详情', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'linapro-content-notice');
  });

  test('TC004a: 从消息列表点击消息弹出预览窗口', async ({ browser }) => {
    const adminToken = await apiLogin(config.adminUser, config.adminPass);
    const userToken = await apiLogin('user001', config.adminPass);
    await apiClearMessages(userToken);
    let noticeId = 0;
    const title = `预览测试通知_${Date.now()}`;
    const content = '<p>这是预览测试的通知内容</p>';
    const context = await browser.newContext({ baseURL: config.baseURL });
    const page = await context.newPage();
    const loginPage = new LoginPage(page);

    try {
      const resp = await fetch(`${PLUGIN_API_BASE}/notice`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${adminToken}`,
        },
        body: JSON.stringify({ title, type: 1, content, status: 1 }),
      });
      const createData = await resp.json();
      expect(createData.code).toBe(0);
      noticeId = createData.data.id;

      await expect.poll(() => apiUnreadCount(userToken), {
        message: 'expected recipient unread-count to include the published notice',
        timeout: 10_000,
      }).toBeGreaterThan(0);

      await loginPage.goto();
      await loginPage.loginAndWaitForRedirect('user001', config.adminPass);

      await page.goto(workspacePath('/system/message'));
      await page.waitForLoadState('networkidle');
      await expect(page.getByText(title, { exact: true }).first()).toBeVisible({
        timeout: 10_000,
      });
      await page.getByText(title, { exact: true }).first().click();

      const modal = page.locator('[role="dialog"]');
      await expect(modal).toBeVisible({ timeout: 10000 });
      await expect(
        modal.getByText('这是预览测试的通知内容'),
      ).toBeVisible({ timeout: 5000 });
      await expect(
        modal.locator('.ant-descriptions'),
      ).toBeVisible({ timeout: 5000 });
    } finally {
      if (noticeId > 0) {
        await fetch(`${PLUGIN_API_BASE}/notice/${noticeId}`, {
          method: 'DELETE',
          headers: { Authorization: `Bearer ${adminToken}` },
        });
      }
      await apiClearMessages(userToken);
      await context.close();
    }
  });
});
